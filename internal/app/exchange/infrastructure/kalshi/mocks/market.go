package kalshi_mocks

import (
	"prediction-risk/internal/app/exchange/infrastructure/kalshi"

	"github.com/stretchr/testify/mock"
)

type MockMarketService struct {
	mock.Mock
}

func (m *MockMarketService) GetMarket(ticker string) (*kalshi.MarketResponse, error) {
	args := m.Called(ticker)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*kalshi.MarketResponse), args.Error(1)
}

func (m *MockMarketService) GetMarkets(params kalshi.GetMarketsOptions) (*kalshi.MarketsResult, error) {
	args := m.Called(params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*kalshi.MarketsResult), args.Error(1)
}
