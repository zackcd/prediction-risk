package mocks

import (
	"prediction-risk/internal/infrastructure/external/kalshi"

	"github.com/stretchr/testify/mock"
)

// MockPortfolioClient mocks the Portfolio client
type MockPortfolioClient struct {
	mock.Mock
}

func (m *MockPortfolioClient) GetPositions(opts kalshi.GetPositionsOptions) (*kalshi.PositionsResult, error) {
	args := m.Called(opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*kalshi.PositionsResult), args.Error(1)
}

func (m *MockPortfolioClient) CreateOrder(order kalshi.CreateOrderRequest) (*kalshi.CreateOrderResponse, error) {
	args := m.Called(order)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*kalshi.CreateOrderResponse), args.Error(1)
}
