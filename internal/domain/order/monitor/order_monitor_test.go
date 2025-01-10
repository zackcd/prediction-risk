package monitor

import (
	"prediction-risk/internal/domain/contract"
	exchangeMocks "prediction-risk/internal/domain/exchange/mocks"
	"prediction-risk/internal/domain/order"
	orderMocks "prediction-risk/internal/domain/order/mocks"
	"prediction-risk/internal/infrastructure/external/kalshi"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStopLossMonitor(t *testing.T) {
	t.Run("should execute YES stop loss when bid drops below threshold", func(t *testing.T) {
		// Arrange
		mockStopOrder := new(orderMocks.MockStopOrderService)
		mockExchange := new(exchangeMocks.MockExchangeService)

		threshold, _ := contract.NewContractPrice(60)
		newOrder := order.NewStopOrder("MARKET-1", contract.SideYes, threshold, nil, nil)

		market := &kalshi.Market{
			Ticker: "MARKET-1",
			YesBid: 55, // Below threshold
			YesAsk: 56,
		}

		mockStopOrder.On("GetActiveOrders").Return([]*order.StopOrder{newOrder}, nil)
		mockStopOrder.On("ExecuteOrder", newOrder.ID(), false).Return(newOrder, nil)
		mockExchange.On("GetMarket", "MARKET-1").Return(market, nil)

		monitor := NewOrderMonitor(mockStopOrder, mockExchange, time.Second, false)

		// Act
		err := monitor.checkOrders()

		// Assert
		assert.NoError(t, err)
		mockStopOrder.AssertExpectations(t)
		mockExchange.AssertExpectations(t)
	})

	t.Run("should execute NO stop loss when bid drops below threshold", func(t *testing.T) {
		// Arrange
		mockStopOrder := new(orderMocks.MockStopOrderService)
		mockExchange := new(exchangeMocks.MockExchangeService)

		threshold, _ := contract.NewContractPrice(60)
		newOrder := order.NewStopOrder("MARKET-1", contract.SideNo, threshold, nil, nil)

		market := &kalshi.Market{
			Ticker: "MARKET-1",
			NoBid:  55, // Below threshold
			NoAsk:  56,
		}

		mockStopOrder.On("GetActiveOrders").Return([]*order.StopOrder{newOrder}, nil)
		mockStopOrder.On("ExecuteOrder", newOrder.ID(), false).Return(newOrder, nil)
		mockExchange.On("GetMarket", "MARKET-1").Return(market, nil)

		monitor := NewOrderMonitor(mockStopOrder, mockExchange, time.Second, false)

		// Act
		err := monitor.checkOrders()

		// Assert
		assert.NoError(t, err)
		mockStopOrder.AssertExpectations(t)
		mockExchange.AssertExpectations(t)
	})

	t.Run("should not execute when bid is above threshold", func(t *testing.T) {
		// Arrange
		mockStopOrder := new(orderMocks.MockStopOrderService)
		mockExchange := new(exchangeMocks.MockExchangeService)

		threshold, _ := contract.NewContractPrice(60)
		newOrder := order.NewStopOrder("MARKET-1", contract.SideYes, threshold, nil, nil)

		market := &kalshi.Market{
			Ticker: "MARKET-1",
			YesBid: 65, // Above threshold
			YesAsk: 66,
		}

		mockStopOrder.On("GetActiveOrders").Return([]*order.StopOrder{newOrder}, nil)
		mockExchange.On("GetMarket", "MARKET-1").Return(market, nil)
		// Note: ExecuteOrder should not be called

		monitor := NewOrderMonitor(mockStopOrder, mockExchange, time.Second, false)

		// Act
		err := monitor.checkOrders()

		// Assert
		assert.NoError(t, err)
		mockStopOrder.AssertExpectations(t)
		mockExchange.AssertExpectations(t)
	})

	t.Run("should handle market fetch error", func(t *testing.T) {
		// Arrange
		mockStopOrder := new(orderMocks.MockStopOrderService)
		mockExchange := new(exchangeMocks.MockExchangeService)

		threshold, _ := contract.NewContractPrice(60)
		newOrder := order.NewStopOrder("MARKET-1", contract.SideYes, threshold, nil, nil)

		mockStopOrder.On("GetActiveOrders").Return([]*order.StopOrder{newOrder}, nil)
		mockExchange.On("GetMarket", "MARKET-1").Return(nil, assert.AnError)

		monitor := NewOrderMonitor(mockStopOrder, mockExchange, time.Second, false)

		// Act
		err := monitor.checkOrders()

		// Assert
		assert.NoError(t, err) // Should not return error as it continues processing
		mockStopOrder.AssertExpectations(t)
		mockExchange.AssertExpectations(t)
	})

	t.Run("should handle GetActiveOrders error", func(t *testing.T) {
		// Arrange
		mockStopOrder := new(orderMocks.MockStopOrderService)
		mockExchange := new(exchangeMocks.MockExchangeService)

		mockStopOrder.On("GetActiveOrders").Return(nil, assert.AnError)

		monitor := NewOrderMonitor(mockStopOrder, mockExchange, time.Second, false)

		// Act
		err := monitor.checkOrders()

		// Assert
		assert.Error(t, err)
		mockStopOrder.AssertExpectations(t)
	})
}
