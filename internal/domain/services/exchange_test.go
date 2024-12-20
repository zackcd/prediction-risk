package services

import (
	"prediction-risk/internal/domain/entities"
	"prediction-risk/internal/infrastructure/external/kalshi"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockMarketGetter struct {
	mock.Mock
}

func (m *mockMarketGetter) GetMarket(ticker string) (*kalshi.MarketResponse, error) {
	args := m.Called(ticker)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*kalshi.MarketResponse), args.Error(1)
}

type mockPortfolioManager struct {
	mock.Mock
}

func (m *mockPortfolioManager) GetPositions(opts kalshi.GetPositionsOptions) (*kalshi.PositionsResult, error) {
	args := m.Called(opts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*kalshi.PositionsResult), args.Error(1)
}

func (m *mockPortfolioManager) CreateOrder(order kalshi.CreateOrderRequest) (*kalshi.CreateOrderResponse, error) {
	args := m.Called(order)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*kalshi.CreateOrderResponse), args.Error(1)
}

func TestGetMarket(t *testing.T) {
	tests := []struct {
		name         string
		ticker       string
		mockSetup    func(*mockMarketGetter)
		expectError  bool
		expectMarket *kalshi.Market
	}{
		{
			name:   "successful market retrieval",
			ticker: "TEST-MKT",
			mockSetup: func(m *mockMarketGetter) {
				m.On("GetMarket", "TEST-MKT").Return(&kalshi.MarketResponse{
					Market: kalshi.Market{
						Ticker: "TEST-MKT",
						Title:  "Test Market",
						Status: "open",
					},
				}, nil)
			},
			expectError: false,
			expectMarket: &kalshi.Market{
				Ticker: "TEST-MKT",
				Title:  "Test Market",
				Status: "open",
			},
		},
		{
			name:   "market not found",
			ticker: "INVALID-MKT",
			mockSetup: func(m *mockMarketGetter) {
				m.On("GetMarket", "INVALID-MKT").Return(nil, &kalshi.KalshiError{
					Reason:     "Market not found",
					StatusCode: 404,
					Body:       "",
				})
			},
			expectError:  true,
			expectMarket: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			marketGetter := &mockMarketGetter{}
			tt.mockSetup(marketGetter)

			service := NewExchangeService(marketGetter, nil) // Portfolio not needed for this test
			market, err := service.GetMarket(tt.ticker)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, market)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectMarket, market)
			}

			marketGetter.AssertExpectations(t)
		})
	}
}

func TestGetPositions(t *testing.T) {
	tests := []struct {
		name            string
		mockSetup       func(*mockPortfolioManager)
		expectError     bool
		expectPositions *kalshi.PositionsResult
	}{
		{
			name: "successful positions retrieval",
			mockSetup: func(m *mockPortfolioManager) {
				m.On("GetPositions", kalshi.GetPositionsOptions{}).Return(&kalshi.PositionsResult{
					MarketPositions: []kalshi.MarketPosition{
						{
							Ticker:   "TEST-MKT",
							Position: 10,
						},
					},
				}, nil)
			},
			expectError: false,
			expectPositions: &kalshi.PositionsResult{
				MarketPositions: []kalshi.MarketPosition{
					{
						Ticker:   "TEST-MKT",
						Position: 10,
					},
				},
			},
		},
		{
			name: "api error",
			mockSetup: func(m *mockPortfolioManager) {
				m.On("GetPositions", kalshi.GetPositionsOptions{}).Return(nil, &kalshi.KalshiError{
					Reason:     "API error",
					StatusCode: 500,
					Body:       "",
				})
			},
			expectError:     true,
			expectPositions: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			portfolioManager := &mockPortfolioManager{}
			tt.mockSetup(portfolioManager)

			service := NewExchangeService(nil, portfolioManager) // Market getter not needed for this test
			positions, err := service.GetPositions()

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, positions)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectPositions, positions)
			}

			portfolioManager.AssertExpectations(t)
		})
	}
}

func TestCreateSellOrder(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name        string
		ticker      string
		count       int
		side        entities.Side
		orderID     string
		mockSetup   func(*mockPortfolioManager)
		expectError bool
		expectOrder *entities.Order
	}{
		{
			name:    "successful yes sell order",
			ticker:  "TEST-MKT",
			count:   5,
			side:    entities.SideYes,
			orderID: "order-123",
			mockSetup: func(m *mockPortfolioManager) {
				expectedRequest := kalshi.CreateOrderRequest{
					Ticker:        "TEST-MKT",
					ClientOrderID: "order-123",
					Side:          kalshi.OrderSideYes,
					Action:        kalshi.OrderActionSell,
					Count:         5,
					Type:          "market",
				}
				m.On("CreateOrder", expectedRequest).Return(&kalshi.CreateOrderResponse{
					Order: kalshi.Order{
						ID:            "exchange-123",
						ClientOrderID: "order-123",
						Ticker:        "TEST-MKT",
						Side:          kalshi.OrderSideYes,
						Action:        string(kalshi.OrderActionSell),
						Type:          "market",
						Status:        "open",
						CreatedTime:   now,
					},
				}, nil)
			},
			expectError: false,
			expectOrder: &entities.Order{
				ExchangeOrderID: "exchange-123",
				Exchange:        entities.ExchangeKalshi,
				InternalOrderID: "order-123",
				Ticker:          "TEST-MKT",
				Side:            entities.SideYes,
				Action:          entities.OrderActionSell,
				OrderType:       entities.OrderTypeMarket,
				Status:          entities.OrderStatus("open"),
			},
		},
		{
			name:    "successful no sell order",
			ticker:  "TEST-MKT",
			count:   5,
			side:    entities.SideNo,
			orderID: "order-123",
			mockSetup: func(m *mockPortfolioManager) {
				expectedRequest := kalshi.CreateOrderRequest{
					Ticker:        "TEST-MKT",
					ClientOrderID: "order-123",
					Side:          kalshi.OrderSideNo,
					Action:        kalshi.OrderActionSell,
					Count:         5,
					Type:          "market",
				}
				m.On("CreateOrder", expectedRequest).Return(&kalshi.CreateOrderResponse{
					Order: kalshi.Order{
						ID:            "exchange-123",
						ClientOrderID: "order-123",
						Ticker:        "TEST-MKT",
						Side:          kalshi.OrderSideNo,
						Action:        string(kalshi.OrderActionSell),
						Type:          "market",
						Status:        "open",
						CreatedTime:   now,
					},
				}, nil)
			},
			expectError: false,
			expectOrder: &entities.Order{
				ExchangeOrderID: "exchange-123",
				Exchange:        entities.ExchangeKalshi,
				InternalOrderID: "order-123",
				Ticker:          "TEST-MKT",
				Side:            entities.SideNo,
				Action:          entities.OrderActionSell,
				OrderType:       entities.OrderTypeMarket,
				Status:          entities.OrderStatus("open"),
			},
		},
		{
			name:    "api error",
			ticker:  "TEST-MKT",
			count:   5,
			side:    entities.SideYes,
			orderID: "order-123",
			mockSetup: func(m *mockPortfolioManager) {
				expectedRequest := kalshi.CreateOrderRequest{
					Ticker:        "TEST-MKT",
					ClientOrderID: "order-123",
					Side:          kalshi.OrderSideYes,
					Action:        kalshi.OrderActionSell,
					Count:         5,
					Type:          "market",
				}
				m.On("CreateOrder", expectedRequest).Return(nil, &kalshi.KalshiError{
					Reason:     "Insufficient balance",
					StatusCode: 400,
					Body:       "",
				})
			},
			expectError: true,
			expectOrder: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			portfolioManager := &mockPortfolioManager{}
			tt.mockSetup(portfolioManager)

			service := NewExchangeService(nil, portfolioManager) // Market getter not needed for this test
			order, err := service.CreateSellOrder(tt.ticker, tt.count, tt.side, tt.orderID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, order)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectOrder, order)
			}

			portfolioManager.AssertExpectations(t)
		})
	}
}
