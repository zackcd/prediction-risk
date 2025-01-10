package mocks

import (
	"prediction-risk/internal/domain/order"

	"github.com/stretchr/testify/mock"
)

// Mock repository
type MockStopOrderRepo struct {
	mock.Mock
}

func (m *MockStopOrderRepo) GetByID(id order.OrderID) (*order.StopOrder, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*order.StopOrder), args.Error(1)
}

func (m *MockStopOrderRepo) Persist(order *order.StopOrder) error {
	args := m.Called(order)
	return args.Error(0)
}

func (m *MockStopOrderRepo) GetAll() ([]*order.StopOrder, error) {
	args := m.Called()
	return args.Get(0).([]*order.StopOrder), args.Error(1)
}
