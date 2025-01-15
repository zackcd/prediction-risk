package kalshi_mocks

import (
	"prediction-risk/internal/app/exchange/infrastructure/kalshi"

	"github.com/stretchr/testify/mock"
)

type MockPortfolioService struct {
	mock.Mock
}

func (m *MockPortfolioService) CreateOrder(request kalshi.CreateOrderRequest) (*kalshi.CreateOrderResponse, error) {
	args := m.Called(request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*kalshi.CreateOrderResponse), args.Error(1)
}

func (m *MockPortfolioService) GetPositions(options kalshi.GetPositionsOptions) (*kalshi.PositionsResult, error) {
	args := m.Called(options)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*kalshi.PositionsResult), args.Error(1)
}
