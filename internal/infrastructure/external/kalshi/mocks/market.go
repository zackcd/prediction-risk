package mocks

import (
	"prediction-risk/internal/infrastructure/external/kalshi"

	"github.com/stretchr/testify/mock"
)

// MockMarketClient mocks the Market client
type MockMarketClient struct {
	mock.Mock
}

func (m *MockMarketClient) GetMarket(ticker string) (*kalshi.MarketResponse, error) {
	args := m.Called(ticker)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*kalshi.MarketResponse), args.Error(1)
}
