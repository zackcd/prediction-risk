package mocks

import (
	"prediction-risk/internal/domain/entities"

	"github.com/stretchr/testify/mock"
)

type MockStopOrderService struct {
	mock.Mock
}

func (m *MockStopOrderService) GetOrder(orderId entities.OrderID) (*entities.StopOrder, error) {
	args := m.Called(orderId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.StopOrder), args.Error(1)
}

func (m *MockStopOrderService) CreateOrder(
	ticker string,
	side entities.Side,
	triggerPrice entities.ContractPrice,
	limitPrice *entities.ContractPrice,
) (*entities.StopOrder, error) {
	args := m.Called(ticker, side, triggerPrice, limitPrice)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.StopOrder), args.Error(1)
}

func (m *MockStopOrderService) UpdateOrder(
	orderId entities.OrderID,
	triggerPrice *entities.ContractPrice,
	limitPrice *entities.ContractPrice,
) (*entities.StopOrder, error) {
	args := m.Called(orderId, triggerPrice, limitPrice)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.StopOrder), args.Error(1)
}

func (m *MockStopOrderService) CancelOrder(orderId entities.OrderID) (*entities.StopOrder, error) {
	args := m.Called(orderId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.StopOrder), args.Error(1)
}

func (m *MockStopOrderService) GetActiveOrders() ([]*entities.StopOrder, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.StopOrder), args.Error(1)
}

func (m *MockStopOrderService) ExecuteOrder(orderId entities.OrderID, isDryRun bool) (*entities.StopOrder, error) {
	args := m.Called(orderId, isDryRun)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.StopOrder), args.Error(1)
}
