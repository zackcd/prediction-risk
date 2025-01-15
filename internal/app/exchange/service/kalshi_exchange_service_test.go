package exchange_service

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"prediction-risk/internal/app/contract"
	exchange_domain "prediction-risk/internal/app/exchange/domain"
	"prediction-risk/internal/app/exchange/infrastructure/kalshi"
	exchange_mock "prediction-risk/internal/app/exchange/mock"
)

func newTestService() (
	*KalshiExchangeService,
	*exchange_mock.MockMarketGetter,
	*exchange_mock.MockPositionGetter,
	*exchange_mock.MockOrderCreator,
) {
	markets := new(exchange_mock.MockMarketGetter)
	positions := new(exchange_mock.MockPositionGetter)
	orders := new(exchange_mock.MockOrderCreator)

	service := &KalshiExchangeService{
		markets:   markets,
		positions: positions,
		orders:    orders,
	}

	return service, markets, positions, orders
}

func TestKalshiExchangeService_CreateOrder(t *testing.T) {
	t.Run("successful sell order creation", func(t *testing.T) {
		// Set up mocks
		service, _, positions, orders := newTestService()

		contractID := contract.ContractIdentifier{
			Ticker: "TEST-1234",
			Side:   contract.SideYes,
		}
		reference := "test-ref"
		quantity := uint(10)
		limitPrice := contract.ContractPrice(50)

		// Mock GetPositions response
		positions.On("GetPositions", kalshi.GetPositionsOptions{
			Ticker: stringPtr("TEST-1234"),
		}).Return(&kalshi.PositionsResult{
			MarketPositions: []kalshi.MarketPosition{
				{
					Ticker:   "TEST-1234",
					Position: 20, // Available position > requested quantity
				},
			},
		}, nil)

		// Mock CreateOrder response
		expectedRequest := kalshi.CreateOrderRequest{
			Ticker:        "TEST-1234",
			ClientOrderID: reference,
			Side:          kalshi.OrderSideYes,
			Action:        kalshi.OrderActionSell,
			Count:         10,
			Type:          "limit",
			YesPrice:      intPtr(50),
		}
		orders.On("CreateOrder", expectedRequest).Return(&kalshi.CreateOrderResponse{
			Order: kalshi.Order{
				ID:     "order-123",
				Ticker: "TEST-1234",
				Status: "open",
			},
		}, nil)

		// Execute test
		params := OrderParams{
			ContractID: contractID,
			Action:     exchange_domain.OrderActionSell,
			Quantity:   &quantity,
			LimitPrice: &limitPrice,
			Reference:  reference,
		}
		order, err := service.CreateOrder(params)

		// Verify results
		require.NoError(t, err)
		require.NotNil(t, order)
		assert.Equal(t, "order-123", order.ExchangeOrderID)
		assert.Equal(t, exchange_domain.OrderActionSell, order.Action)
		assert.Equal(t, contractID.Side, order.Side)
		positions.AssertExpectations(t)
		orders.AssertExpectations(t)
	})

	t.Run("sell order with no position available", func(t *testing.T) {
		service, _, positions, _ := newTestService()

		// Mock empty positions response
		positions.On("GetPositions", mock.Anything).Return(&kalshi.PositionsResult{
			MarketPositions: []kalshi.MarketPosition{},
		}, nil)

		params := OrderParams{
			ContractID: contract.ContractIdentifier{
				Ticker: "TEST-1234",
				Side:   contract.SideYes,
			},
			Action:    exchange_domain.OrderActionSell,
			Reference: "test-ref",
		}
		order, err := service.CreateOrder(params)

		require.Error(t, err)
		assert.Nil(t, order)
		assert.Contains(t, err.Error(), "position not found")
		positions.AssertExpectations(t)
	})

	t.Run("position service error", func(t *testing.T) {
		service, _, positions, _ := newTestService()

		positions.On("GetPositions", mock.Anything).Return(nil, errors.New("service error"))

		params := OrderParams{
			ContractID: contract.ContractIdentifier{
				Ticker: "TEST-1234",
				Side:   contract.SideYes,
			},
			Action:    exchange_domain.OrderActionSell,
			Reference: "test-ref",
		}
		order, err := service.CreateOrder(params)

		require.Error(t, err)
		assert.Nil(t, order)
		assert.Contains(t, err.Error(), "get positions")
		positions.AssertExpectations(t)
	})

	t.Run("sell quantity calculation", func(t *testing.T) {
		testCases := []struct {
			name          string
			position      int
			requestSize   *uint
			expectedCount int
		}{
			{
				name:          "full position when size not specified",
				position:      15,
				requestSize:   nil,
				expectedCount: 15,
			},
			{
				name:          "limited by requested size",
				position:      20,
				requestSize:   uintPtr(10),
				expectedCount: 10,
			},
			{
				name:          "limited by position size",
				position:      5,
				requestSize:   uintPtr(10),
				expectedCount: 5,
			},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				service, _, positions, orders := newTestService()

				positions.On("GetPositions", mock.Anything).Return(&kalshi.PositionsResult{
					MarketPositions: []kalshi.MarketPosition{
						{
							Ticker:   "TEST-1234",
							Position: tc.position,
						},
					},
				}, nil)

				orders.On("CreateOrder", mock.MatchedBy(func(req kalshi.CreateOrderRequest) bool {
					return req.Count == tc.expectedCount
				})).Return(&kalshi.CreateOrderResponse{
					Order: kalshi.Order{
						ID:     "order-123",
						Ticker: "TEST-1234",
						Status: "open",
					},
				}, nil)

				params := OrderParams{
					ContractID: contract.ContractIdentifier{
						Ticker: "TEST-1234",
						Side:   contract.SideYes,
					},
					Action:    exchange_domain.OrderActionSell,
					Quantity:  tc.requestSize,
					Reference: "test-ref",
				}
				order, err := service.CreateOrder(params)

				require.NoError(t, err)
				require.NotNil(t, order)
				positions.AssertExpectations(t)
				orders.AssertExpectations(t)
			})
		}
	})
}

// Helper functions for creating pointers to values
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func uintPtr(u uint) *uint {
	return &u
}
