package trigger_mock

import (
	"context"
	trigger_domain "prediction-risk/internal/app/risk/trigger/domain"

	"github.com/stretchr/testify/mock"
)

// MockTriggerRepository is a mock implementation of TriggerRepository
type MockTriggerRepository struct {
	mock.Mock
}

func (m *MockTriggerRepository) Persist(ctx context.Context, trigger *trigger_domain.Trigger) error {
	args := m.Called(ctx, trigger)
	return args.Error(0)
}

func (m *MockTriggerRepository) Get(ctx context.Context, id trigger_domain.TriggerID) (*trigger_domain.Trigger, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*trigger_domain.Trigger), args.Error(1)
}

func (m *MockTriggerRepository) GetAll(ctx context.Context) ([]*trigger_domain.Trigger, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*trigger_domain.Trigger), args.Error(1)
}
