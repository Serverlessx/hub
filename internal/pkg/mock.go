package pkg

import (
	"context"

	"github.com/artifacthub/hub/internal/hub"
	"github.com/stretchr/testify/mock"
)

// ManagerMock is a mock implementation of the PackageManager interface.
type ManagerMock struct {
	mock.Mock
}

// GetJSON implements the PackageManager interface.
func (m *ManagerMock) GetJSON(ctx context.Context, input *hub.GetPackageInput) ([]byte, error) {
	args := m.Called(ctx, input)
	data, _ := args.Get(0).([]byte)
	return data, args.Error(1)
}

// GetStarredByUserJSON implements the PackageManager interface.
func (m *ManagerMock) GetStarredByUserJSON(ctx context.Context) ([]byte, error) {
	args := m.Called(ctx)
	data, _ := args.Get(0).([]byte)
	return data, args.Error(1)
}

// GetStarsJSON implements the PackageManager interface.
func (m *ManagerMock) GetStarsJSON(ctx context.Context, packageID string) ([]byte, error) {
	args := m.Called(ctx)
	data, _ := args.Get(0).([]byte)
	return data, args.Error(1)
}

// GetStatsJSON implements the PackageManager interface.
func (m *ManagerMock) GetStatsJSON(ctx context.Context) ([]byte, error) {
	args := m.Called(ctx)
	data, _ := args.Get(0).([]byte)
	return data, args.Error(1)
}

// GetUpdatesJSON implements the PackageManager interface.
func (m *ManagerMock) GetUpdatesJSON(ctx context.Context) ([]byte, error) {
	args := m.Called(ctx)
	data, _ := args.Get(0).([]byte)
	return data, args.Error(1)
}

// Register implements the PackageManager interface.
func (m *ManagerMock) Register(ctx context.Context, pkg *hub.Package) error {
	args := m.Called(ctx, pkg)
	return args.Error(0)
}

// SearchJSON implements the PackageManager interface.
func (m *ManagerMock) SearchJSON(ctx context.Context, input *hub.SearchPackageInput) ([]byte, error) {
	args := m.Called(ctx)
	data, _ := args.Get(0).([]byte)
	return data, args.Error(1)
}

// ToggleStar implements the PackageManager interface.
func (m *ManagerMock) ToggleStar(ctx context.Context, packageID string) error {
	args := m.Called(ctx, packageID)
	return args.Error(0)
}

// Unregister implements the PackageManager interface.
func (m *ManagerMock) Unregister(ctx context.Context, pkg *hub.Package) error {
	args := m.Called(ctx, pkg)
	return args.Error(0)
}
