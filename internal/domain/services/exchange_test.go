package services

import (
	"prediction-risk/internal/domain/entities"
	"prediction-risk/internal/infrastructure/external/kalshi"
	"prediction-risk/internal/infrastructure/external/kalshi/mocks"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetMarket(t *testing.T) {
	tests := []struct {
		name         string
		ticker       string
		mockSetup    func(*mocks.MockKalshiClient)
		expectError  bool
		expectMarket *kalshi.Market
	}{
		{
			name:   "successful market retrieval",
			ticker: "TEST-MKT",
			mockSetup: func(m *mocks.MockKalshiClient) {
				m.Market.On("GetMarket", "TEST-MKT").Return(&kalshi.MarketResponse{
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
			mockSetup: func(m *mocks.MockKalshiClient) {
				m.Market.On("GetMarket", "INVALID-MKT").Return(nil, &kalshi.Error{
					Message: "Market not found",
					Status:  404,
				})
			},
			expectError:  true,
			expectMarket: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client and set up expectations
			mockClient := mocks.NewMockKalshiClient()
			tt.mockSetup(mockClient)

			// Create service with mock client
			service := NewExchangeService(mockClient)

			// Call the method
			market, err := service.GetMarket(tt.ticker)

			// Assert results
			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, market)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectMarket, market)
			}

			// Verify expectations
			mockClient.Market.AssertExpectations(t)
		})
	}
}

func TestGetPositions(t *testing.T) {
	tests := []struct {
		name            string
		mockSetup       func(*mocks.MockKalshiClient)
		expectError     bool
		expectPositions *kalshi.PositionsResult
	}{
		{
			name: "successful positions retrieval",
			mockSetup: func(m *mocks.MockKalshiClient) {
				m.Portfolio.On("GetPositions", kalshi.GetPositionsOptions{}).Return(&kalshi.PositionsResult{
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
			mockSetup: func(m *mocks.MockKalshiClient) {
				m.Portfolio.On("GetPositions", kalshi.GetPositionsOptions{}).Return(nil, &kalshi.Error{
					Message: "API error",
					Status:  500,
				})
			},
			expectError:     true,
			expectPositions: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := mocks.NewMockKalshiClient()
			tt.mockSetup(mockClient)

			service := NewExchangeService(mockClient)
			positions, err := service.GetPositions()

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, positions)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectPositions, positions)
			}

			mockClient.Portfolio.AssertExpectations(t)
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
		mockSetup   func(*mocks.MockKalshiClient)
		expectError bool
		expectOrder *entities.Order
	}{
		{
			name:    "successful yes sell order",
			ticker:  "TEST-MKT",
			count:   5,
			side:    entities.SideYes,
			orderID: "order-123",
			mockSetup: func(m *mocks.MockKalshiClient) {
				expectedRequest := kalshi.CreateOrderRequest{
					Ticker:        "TEST-MKT",
					ClientOrderID: "order-123",
					Side:          kalshi.OrderSideYes,
					Action:        kalshi.OrderActionSell,
					Count:         5,
					Type:          "market",
				}
				m.Portfolio.On("CreateOrder", expectedRequest).Return(&kalshi.CreateOrderResponse{
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
			mockSetup: func(m *mocks.MockKalshiClient) {
				expectedRequest := kalshi.CreateOrderRequest{
					Ticker:        "TEST-MKT",
					ClientOrderID: "order-123",
					Side:          kalshi.OrderSideNo,
					Action:        kalshi.OrderActionSell,
					Count:         5,
					Type:          "market",
				}
				m.Portfolio.On("CreateOrder", expectedRequest).Return(&kalshi.CreateOrderResponse{
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
			mockSetup: func(m *mocks.MockKalshiClient) {
				expectedRequest := kalshi.CreateOrderRequest{
					Ticker:        "TEST-MKT",
					ClientOrderID: "order-123",
					Side:          kalshi.OrderSideYes,
					Action:        kalshi.OrderActionSell,
					Count:         5,
					Type:          "market",
				}
				m.Portfolio.On("CreateOrder", expectedRequest).Return(nil, &kalshi.Error{
					Message: "Insufficient balance",
					Status:  400,
				})
			},
			expectError: true,
			expectOrder: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := mocks.NewMockKalshiClient()
			tt.mockSetup(mockClient)

			service := NewExchangeService(mockClient)
			order, err := service.CreateSellOrder(tt.ticker, tt.count, tt.side, tt.orderID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, order)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectOrder, order)
			}

			mockClient.Portfolio.AssertExpectations(t)
		})
	}
}
