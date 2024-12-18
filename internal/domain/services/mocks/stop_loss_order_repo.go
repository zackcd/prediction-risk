package mocks

import (
	"prediction-risk/internal/domain/entities"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// Mock repository
type MockStopLossOrderRepo struct {
	mock.Mock
}

func (m *MockStopLossOrderRepo) GetByID(id uuid.UUID) (*entities.StopLossOrder, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.StopLossOrder), args.Error(1)
}

func (m *MockStopLossOrderRepo) Persist(order *entities.StopLossOrder) error {
	args := m.Called(order)
	return args.Error(0)
}

func (m *MockStopLossOrderRepo) GetAll() ([]*entities.StopLossOrder, error) {
	args := m.Called()
	return args.Get(0).([]*entities.StopLossOrder), args.Error(1)
}
