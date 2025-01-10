package mocks

import (
	"prediction-risk/internal/domain/contract"
	"prediction-risk/internal/domain/order"

	"github.com/stretchr/testify/mock"
)

type MockStopOrderService struct {
	mock.Mock
}

func (m *MockStopOrderService) GetOrder(orderId order.OrderID) (*order.StopOrder, error) {
	args := m.Called(orderId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*order.StopOrder), args.Error(1)
}

func (m *MockStopOrderService) CreateOrder(
	ticker string,
	side contract.Side,
	triggerPrice contract.ContractPrice,
	limitPrice *contract.ContractPrice,
) (*order.StopOrder, error) {
	args := m.Called(ticker, side, triggerPrice, limitPrice)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*order.StopOrder), args.Error(1)
}

func (m *MockStopOrderService) UpdateOrder(
	orderId order.OrderID,
	triggerPrice *contract.ContractPrice,
	limitPrice *contract.ContractPrice,
) (*order.StopOrder, error) {
	args := m.Called(orderId, triggerPrice, limitPrice)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*order.StopOrder), args.Error(1)
}

func (m *MockStopOrderService) CancelOrder(orderId order.OrderID) (*order.StopOrder, error) {
	args := m.Called(orderId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*order.StopOrder), args.Error(1)
}

func (m *MockStopOrderService) GetActiveOrders() ([]*order.StopOrder, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*order.StopOrder), args.Error(1)
}

func (m *MockStopOrderService) ExecuteOrder(orderId order.OrderID, isDryRun bool) (*order.StopOrder, error) {
	args := m.Called(orderId, isDryRun)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*order.StopOrder), args.Error(1)
}
