package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/artifacthub/hub/cmd/hub/handlers"
	"github.com/artifacthub/hub/internal/chartrepo"
	"github.com/artifacthub/hub/internal/email"
	"github.com/artifacthub/hub/internal/hub"
	"github.com/artifacthub/hub/internal/img/pg"
	"github.com/artifacthub/hub/internal/org"
	"github.com/artifacthub/hub/internal/pkg"
	"github.com/artifacthub/hub/internal/user"
	"github.com/artifacthub/hub/internal/util"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

func main() {
	// Setup configuration and logger
	cfg, err := util.SetupConfig("hub")
	if err != nil {
		log.Fatal().Err(err).Msg("Configuration setup failed")
	}
	fields := map[string]interface{}{"cmd": "hub"}
	if err := util.SetupLogger(cfg, fields); err != nil {
		log.Fatal().Err(err).Msg("Logger setup failed")
	}

	// Setup services required by the handlers to operate
	db, err := util.SetupDB(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Database setup failed")
	}
	var es hub.EmailSender
	if s := email.NewSender(cfg); s != nil {
		es = s
	}
	svc := &handlers.Services{
		OrganizationManager:    org.NewManager(db, es),
		UserManager:            user.NewManager(db, es),
		PackageManager:         pkg.NewManager(db),
		ChartRepositoryManager: chartrepo.NewManager(db),
		ImageStore:             pg.NewImageStore(db),
	}

	// Setup and launch server
	addr := cfg.GetString("server.addr")
	srv := &http.Server{
		Addr:         addr,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  1 * time.Minute,
		Handler:      handlers.Setup(cfg, svc).Router,
	}
	go func() {
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Hub server ListenAndServe failed")
		}
	}()
	log.Info().Str("addr", addr).Int("pid", os.Getpid()).Msg("Hub server running!")

	// Setup and launch metrics server
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		err := http.ListenAndServe(cfg.GetString("server.metricsAddr"), nil)
		if err != nil {
			log.Fatal().Err(err).Msg("Metrics server ListenAndServe failed")
		}
	}()

	// Shutdown server gracefully when SIGINT or SIGTERM signal is received
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)
	<-shutdown
	log.Info().Msg("Hub server shutting down..")
	ctx, cancel := context.WithTimeout(context.Background(), cfg.GetDuration("server.shutdownTimeout"))
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Hub server shutdown failed")
	}
	log.Info().Msg("Hub server stopped")
}
