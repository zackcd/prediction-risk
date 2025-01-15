package exchange_mock

import (
	"prediction-risk/internal/app/exchange/infrastructure/kalshi"

	"github.com/stretchr/testify/mock"
)

type MockMarketGetter struct {
	mock.Mock
}

func (m *MockMarketGetter) GetMarket(ticker string) (*kalshi.MarketResponse, error) {
	args := m.Called(ticker)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*kalshi.MarketResponse), args.Error(1)
}
