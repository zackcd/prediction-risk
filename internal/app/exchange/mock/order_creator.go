package exchange_mock

import (
	"prediction-risk/internal/app/exchange/infrastructure/kalshi"

	"github.com/stretchr/testify/mock"
)

type MockOrderCreator struct {
	mock.Mock
}

func (m *MockOrderCreator) CreateOrder(request kalshi.CreateOrderRequest) (*kalshi.CreateOrderResponse, error) {
	args := m.Called(request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*kalshi.CreateOrderResponse), args.Error(1)
}
