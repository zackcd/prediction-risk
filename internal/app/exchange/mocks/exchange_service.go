package exchange_mocks

import (
	"prediction-risk/internal/app/contract"
	exchange_domain "prediction-risk/internal/app/exchange/domain"
	"prediction-risk/internal/app/exchange/infrastructure/kalshi"

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
	limitPrice *contract.ContractPrice,
	orderID *exchange_domain.OrderID,
) (*exchange_domain.Order, error) {
	args := m.Called(ticker, count, side, limitPrice, orderID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*exchange_domain.Order), args.Error(1)
}
