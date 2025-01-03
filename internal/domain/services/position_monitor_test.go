package services

import (
	"prediction-risk/internal/domain/entities"
	"prediction-risk/internal/domain/services/mocks"
	"prediction-risk/internal/infrastructure/external/kalshi"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPositionMonitor(t *testing.T) {
	t.Run("lifecycle", func(t *testing.T) {
		// Test basic start/stop functionality
		mockExchange := new(mocks.MockExchangeService)
		mockStopOrder := new(mocks.MockStopOrderService)

		// Setup initial sync expectations
		mockExchange.On("GetPositions").Return(&kalshi.PositionsResult{
			MarketPositions: []kalshi.MarketPosition{},
		}, nil).Once()
		mockStopOrder.On("GetActiveOrders").Return([]*entities.StopOrder{}, nil).Once()

		monitor := NewPositionMonitor(mockExchange, mockStopOrder, time.Millisecond*100)
		monitor.Start()
		time.Sleep(time.Millisecond * 50) // Let it run briefly
		monitor.Stop()

		mockExchange.AssertExpectations(t)
		mockStopOrder.AssertExpectations(t)
	})

	t.Run("creates stop orders for new positions", func(t *testing.T) {
		mockExchange := new(mocks.MockExchangeService)
		mockStopOrder := new(mocks.MockStopOrderService)

		// Setup market data for stop price calculation
		mockExchange.On("GetMarket", "AAPL").Return(&kalshi.Market{
			LastPrice: 100,
		}, nil)

		// Setup position data with one long position
		mockExchange.On("GetPositions").Return(&kalshi.PositionsResult{
			MarketPositions: []kalshi.MarketPosition{
				{Ticker: "AAPL", Position: 10},
			},
		}, nil)

		// No existing stop orders
		mockStopOrder.On("GetActiveOrders").Return([]*entities.StopOrder{}, nil)

		// Expect creation of new stop order
		stopPrice, _ := entities.NewContractPrice(90) // 90 is 10% below 100
		mockStopOrder.On("CreateOrder",
			"AAPL",
			entities.SideYes,
			stopPrice,
			(*entities.ContractPrice)(nil),
		).Return(&entities.StopOrder{}, nil)

		monitor := NewPositionMonitor(mockExchange, mockStopOrder, time.Second)
		monitor.syncPositions() // Test single sync

		mockExchange.AssertExpectations(t)
		mockStopOrder.AssertExpectations(t)
	})

	t.Run("handles short positions correctly", func(t *testing.T) {
		mockExchange := new(mocks.MockExchangeService)
		mockStopOrder := new(mocks.MockStopOrderService)

		mockExchange.On("GetMarket", "AAPL").Return(&kalshi.Market{
			LastPrice: 100,
		}, nil)

		mockExchange.On("GetPositions").Return(&kalshi.PositionsResult{
			MarketPositions: []kalshi.MarketPosition{
				{Ticker: "AAPL", Position: -10}, // Short position
			},
		}, nil)

		mockStopOrder.On("GetActiveOrders").Return([]*entities.StopOrder{}, nil)

		stopPrice, _ := entities.NewContractPrice(90)
		mockStopOrder.On("CreateOrder",
			"AAPL",
			entities.SideNo,
			stopPrice,
			(*entities.ContractPrice)(nil),
		).Return(&entities.StopOrder{}, nil)

		monitor := NewPositionMonitor(mockExchange, mockStopOrder, time.Second)
		monitor.syncPositions()

		mockExchange.AssertExpectations(t)
		mockStopOrder.AssertExpectations(t)
	})

	t.Run("cancels stop orders for closed positions", func(t *testing.T) {
		mockExchange := new(mocks.MockExchangeService)
		mockStopOrder := new(mocks.MockStopOrderService)

		orderID := entities.NewOrderID()
		price, err := entities.NewContractPrice(90)
		assert.NoError(t, err)
		existingOrder := entities.NewStopOrder("AAPL", entities.SideYes, price, nil, &orderID)

		mockExchange.On("GetPositions").Return(&kalshi.PositionsResult{
			MarketPositions: []kalshi.MarketPosition{
				{Ticker: "AAPL", Position: 0}, // Closed position
			},
		}, nil)

		mockStopOrder.On("GetActiveOrders").Return([]*entities.StopOrder{
			existingOrder,
		}, nil)

		mockStopOrder.On("CancelOrder", orderID).Return(existingOrder, nil)

		monitor := NewPositionMonitor(mockExchange, mockStopOrder, time.Second)
		monitor.syncPositions()

		mockExchange.AssertExpectations(t)
		mockStopOrder.AssertExpectations(t)
	})

	t.Run("cancels orphaned stop orders", func(t *testing.T) {
		mockExchange := new(mocks.MockExchangeService)
		mockStopOrder := new(mocks.MockStopOrderService)

		orderID := entities.NewOrderID()
		price, err := entities.NewContractPrice(90)
		assert.NoError(t, err)
		existingOrder := entities.NewStopOrder("AAPL", entities.SideYes, price, nil, &orderID)

		// No positions returned from exchange
		mockExchange.On("GetPositions").Return(&kalshi.PositionsResult{
			MarketPositions: []kalshi.MarketPosition{},
		}, nil)

		// But we have an active stop order
		mockStopOrder.On("GetActiveOrders").Return([]*entities.StopOrder{
			existingOrder,
		}, nil)

		mockStopOrder.On("CancelOrder", orderID).Return(existingOrder, nil)

		monitor := NewPositionMonitor(mockExchange, mockStopOrder, time.Second)
		monitor.syncPositions()

		mockExchange.AssertExpectations(t)
		mockStopOrder.AssertExpectations(t)
	})

	t.Run("handles exchange errors gracefully", func(t *testing.T) {
		mockExchange := new(mocks.MockExchangeService)
		mockStopOrder := new(mocks.MockStopOrderService)

		mockExchange.On("GetPositions").Return((*kalshi.PositionsResult)(nil),
			assert.AnError)

		monitor := NewPositionMonitor(mockExchange, mockStopOrder, time.Second)
		err := monitor.syncPositions()

		assert.Error(t, err)
		mockExchange.AssertExpectations(t)
		mockStopOrder.AssertExpectations(t)
	})

	t.Run("handles stop order service errors gracefully", func(t *testing.T) {
		mockExchange := new(mocks.MockExchangeService)
		mockStopOrder := new(mocks.MockStopOrderService)

		mockExchange.On("GetPositions").Return(&kalshi.PositionsResult{
			MarketPositions: []kalshi.MarketPosition{
				{Ticker: "AAPL", Position: 10},
			},
		}, nil)

		mockStopOrder.On("GetActiveOrders").Return(nil, assert.AnError)

		monitor := NewPositionMonitor(mockExchange, mockStopOrder, time.Second)
		err := monitor.syncPositions()

		assert.Error(t, err)
		mockExchange.AssertExpectations(t)
		mockStopOrder.AssertExpectations(t)
	})
}
