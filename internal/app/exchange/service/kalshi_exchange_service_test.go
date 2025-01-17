package exchange_service

import (
	"errors"
	"testing"
	"time"

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

func TestKalshiExchangeService_GetPositions(t *testing.T) {
	t.Run("successfully retrieves multiple positions", func(t *testing.T) {
		service, _, positions, _ := newTestService()

		// Mock the positions response
		positions.On("GetPositions", kalshi.GetPositionsOptions{}).Return(&kalshi.PositionsResult{
			MarketPositions: []kalshi.MarketPosition{
				{
					Ticker:   "MARKET-1",
					Position: 100, // YES position
				},
				{
					Ticker:   "MARKET-2",
					Position: -50, // NO position
				},
			},
		}, nil)

		// Execute test
		result, err := service.GetPositions()

		// Verify results
		require.NoError(t, err)
		require.Len(t, result, 2)

		// Verify first position (YES side)
		assert.Equal(t, contract.Ticker("MARKET-1"), result[0].ContractID.Ticker)
		assert.Equal(t, contract.SideYes, result[0].ContractID.Side)
		assert.Equal(t, uint(100), result[0].Quantity)

		// Verify second position (NO side)
		assert.Equal(t, contract.Ticker("MARKET-2"), result[1].ContractID.Ticker)
		assert.Equal(t, contract.SideNo, result[1].ContractID.Side)
		assert.Equal(t, uint(50), result[1].Quantity)

		positions.AssertExpectations(t)
	})

	t.Run("handles empty positions list", func(t *testing.T) {
		service, _, positions, _ := newTestService()

		positions.On("GetPositions", kalshi.GetPositionsOptions{}).Return(&kalshi.PositionsResult{
			MarketPositions: []kalshi.MarketPosition{},
		}, nil)

		result, err := service.GetPositions()

		require.NoError(t, err)
		assert.Empty(t, result)
		positions.AssertExpectations(t)
	})

	t.Run("handles API error", func(t *testing.T) {
		service, _, positions, _ := newTestService()

		positions.On("GetPositions", kalshi.GetPositionsOptions{}).Return(nil, errors.New("API error"))

		result, err := service.GetPositions()

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "fetch market from kalshi")
		positions.AssertExpectations(t)
	})
}

func TestKalshiExchangeService_GetMarket(t *testing.T) {
	t.Run("successfully retrieves market details", func(t *testing.T) {
		service, markets, _, _ := newTestService()

		// Mock the market response
		markets.On("GetMarket", "TEST-MARKET").Return(&kalshi.MarketResponse{
			Market: kalshi.Market{
				Ticker:         "TEST-MARKET",
				Title:          "Test Market",
				Category:       "TEST",
				Status:         "active",
				OpenTime:       time.Now(),
				CloseTime:      time.Now().Add(24 * time.Hour),
				ExpirationTime: time.Now().Add(48 * time.Hour),
				YesBid:         60,
				YesAsk:         65,
				NoBid:          35,
				NoAsk:          40,
				LastPrice:      62,
				PreviousYesBid: 58,
				PreviousYesAsk: 63,
				Volume:         1000,
				Volume24H:      500,
				OpenInterest:   750,
				Liquidity:      1500,
				NotionalValue:  100,
				TickSize:       1,
				RiskLimitCents: 10000,
			},
		}, nil)

		// Execute test
		result, err := service.GetMarket("TEST-MARKET")

		// Verify results
		require.NoError(t, err)
		require.NotNil(t, result)

		// Verify basic market info
		assert.Equal(t, contract.Ticker("TEST-MARKET"), result.Ticker)
		assert.Equal(t, "Test Market", result.Info.Title)
		assert.Equal(t, "TEST", result.Info.Category)
		assert.Equal(t, exchange_domain.MarketTypeBinary, result.Info.Type)

		// Verify YES side pricing
		assert.Equal(t, contract.ContractPrice(60), result.Pricing.YesSide.Bid)
		assert.Equal(t, contract.ContractPrice(65), result.Pricing.YesSide.Ask)
		assert.Equal(t, contract.ContractPrice(62), result.Pricing.YesSide.LastPrice)
		assert.Equal(t, contract.ContractPrice(58), result.Pricing.YesSide.PreviousBid)
		assert.Equal(t, contract.ContractPrice(63), result.Pricing.YesSide.PreviousAsk)

		// Verify NO side pricing (only current bid/ask)
		assert.Equal(t, contract.ContractPrice(35), result.Pricing.NoSide.Bid)
		assert.Equal(t, contract.ContractPrice(40), result.Pricing.NoSide.Ask)

		// Verify trading constraints
		assert.Equal(t, contract.ContractPrice(100), result.Constraints.NotionalValue)
		assert.Equal(t, contract.ContractPrice(1), result.Constraints.TickSize)
		assert.Equal(t, contract.ContractPrice(10000), result.Constraints.RiskLimit)

		// Verify liquidity metrics
		assert.Equal(t, 1000, result.Liquidity.Volume)
		assert.Equal(t, 500, result.Liquidity.Volume24H)
		assert.Equal(t, 750, result.Liquidity.OpenInterest)
		assert.Equal(t, 1500, result.Liquidity.Liquidity)

		markets.AssertExpectations(t)
	})

	t.Run("handles API error", func(t *testing.T) {
		service, markets, _, _ := newTestService()

		markets.On("GetMarket", "TEST-MARKET").Return(nil, errors.New("API error"))

		result, err := service.GetMarket("TEST-MARKET")

		require.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "fetch market from kalshi")
		markets.AssertExpectations(t)
	})
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
