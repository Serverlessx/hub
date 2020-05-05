package subscription

import (
	"context"

	"github.com/artifacthub/hub/internal/hub"
	"github.com/stretchr/testify/mock"
)

// ManagerMock is a mock implementation of the SubscriptionManager interface.
type ManagerMock struct {
	mock.Mock
}

// Add implements the SubscriptionManager interface.
func (m *ManagerMock) Add(ctx context.Context, s *hub.Subscription) error {
	args := m.Called(ctx, s)
	return args.Error(0)
}

// Delete implements the SubscriptionManager interface.
func (m *ManagerMock) Delete(ctx context.Context, s *hub.Subscription) error {
	args := m.Called(ctx, s)
	return args.Error(0)
}

// GetByPackageJSON implements the SubscriptionManager interface.
func (m *ManagerMock) GetByPackageJSON(ctx context.Context, packageID string) ([]byte, error) {
	args := m.Called(ctx, packageID)
	data, _ := args.Get(0).([]byte)
	return data, args.Error(1)
}

// GetByUserJSON implements the SubscriptionManager interface.
func (m *ManagerMock) GetByUserJSON(ctx context.Context) ([]byte, error) {
	args := m.Called(ctx)
	data, _ := args.Get(0).([]byte)
	return data, args.Error(1)
}