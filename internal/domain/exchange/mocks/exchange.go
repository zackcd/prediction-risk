package mocks

import (
	"prediction-risk/internal/domain/contract"
	"prediction-risk/internal/domain/exchange"
	"prediction-risk/internal/infrastructure/external/kalshi"

	"github.com/stretchr/testify/mock"
)

// MockExchangeService is a mock implementation of the ExchangeService interface
type MockExchangeService struct {
	mock.Mock
}

func (m *MockExchangeService) GetMarket(ticker string) (*kalshi.Market, error) {
	args := m.Called(ticker)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*kalshi.Market), args.Error(1)
}

func (m *MockExchangeService) GetPositions() (*kalshi.PositionsResult, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*kalshi.PositionsResult), args.Error(1)
}

func (m *MockExchangeService) CreateSellOrder(
	ticker string,
	count int,
	side contract.Side,
	orderID string,
	limitPrice *contract.ContractPrice,
) (*exchange.ExchangeOrder, error) {
	args := m.Called(ticker, count, side, orderID, limitPrice)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*exchange.ExchangeOrder), args.Error(1)
}
