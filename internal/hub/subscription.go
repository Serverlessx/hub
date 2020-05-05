package hub

import "context"

// NotificationKind represents the kind of a notification.
type NotificationKind int64

const (
	// NewPackageRelease represents a notification for a new package release.
	NewPackageRelease NotificationKind = 0

	// SecurityAlert represents a notification for a security alert.
	SecurityAlert NotificationKind = 1
)

// Subscription represents a user's subscription to receive notifications about
// a given package.
type Subscription struct {
	UserID           string           `json:"user_id"`
	PackageID        string           `json:"package_id"`
	NotificationKind NotificationKind `json:"notification_kind"`
}

// SubscriptionManager describes the methods a SubscriptionManager
// implementation must provide.
type SubscriptionManager interface {
	Add(ctx context.Context, s *Subscription) error
	Delete(ctx context.Context, s *Subscription) error
	GetByPackageJSON(ctx context.Context, packageID string) ([]byte, error)
	GetByUserJSON(ctx context.Context) ([]byte, error)
	GetJSON(ctx context.Context, s *Subscription) ([]byte, error)
}
