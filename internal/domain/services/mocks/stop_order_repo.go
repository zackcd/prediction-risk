package mocks

import (
	"prediction-risk/internal/domain/entities"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// Mock repository
type MockStopOrderRepo struct {
	mock.Mock
}

func (m *MockStopOrderRepo) GetByID(id uuid.UUID) (*entities.StopOrder, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.StopOrder), args.Error(1)
}

func (m *MockStopOrderRepo) Persist(order *entities.StopOrder) error {
	args := m.Called(order)
	return args.Error(0)
}

func (m *MockStopOrderRepo) GetAll() ([]*entities.StopOrder, error) {
	args := m.Called()
	return args.Get(0).([]*entities.StopOrder), args.Error(1)
}
