package services

import (
	"prediction-risk/internal/domain/entities"
	"prediction-risk/internal/domain/services/mocks"
	"prediction-risk/internal/infrastructure/external/kalshi"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestStopLossMonitor(t *testing.T) {
	t.Run("should execute YES stop loss when bid drops below threshold", func(t *testing.T) {
		// Arrange
		mockStopOrder := new(mocks.MockStopOrderService)
		mockExchange := new(mocks.MockExchangeService)

		threshold, _ := entities.NewContractPrice(60)
		order := entities.NewStopOrder("MARKET-1", entities.SideYes, threshold, nil, nil)

		market := &kalshi.Market{
			Ticker: "MARKET-1",
			YesBid: 55, // Below threshold
			YesAsk: 56,
		}

		mockStopOrder.On("GetActiveOrders").Return([]*entities.StopOrder{order}, nil)
		mockStopOrder.On("ExecuteOrder", order.ID(), false).Return(order, nil)
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
		mockStopOrder := new(mocks.MockStopOrderService)
		mockExchange := new(mocks.MockExchangeService)

		threshold, _ := entities.NewContractPrice(60)
		order := entities.NewStopOrder("MARKET-1", entities.SideNo, threshold, nil, nil)

		market := &kalshi.Market{
			Ticker: "MARKET-1",
			NoBid:  55, // Below threshold
			NoAsk:  56,
		}

		mockStopOrder.On("GetActiveOrders").Return([]*entities.StopOrder{order}, nil)
		mockStopOrder.On("ExecuteOrder", order.ID(), false).Return(order, nil)
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
		mockStopOrder := new(mocks.MockStopOrderService)
		mockExchange := new(mocks.MockExchangeService)

		threshold, _ := entities.NewContractPrice(60)
		order := entities.NewStopOrder("MARKET-1", entities.SideYes, threshold, nil, nil)

		market := &kalshi.Market{
			Ticker: "MARKET-1",
			YesBid: 65, // Above threshold
			YesAsk: 66,
		}

		mockStopOrder.On("GetActiveOrders").Return([]*entities.StopOrder{order}, nil)
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
		mockStopOrder := new(mocks.MockStopOrderService)
		mockExchange := new(mocks.MockExchangeService)

		threshold, _ := entities.NewContractPrice(60)
		order := entities.NewStopOrder("MARKET-1", entities.SideYes, threshold, nil, nil)

		mockStopOrder.On("GetActiveOrders").Return([]*entities.StopOrder{order}, nil)
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
		mockStopOrder := new(mocks.MockStopOrderService)
		mockExchange := new(mocks.MockExchangeService)

		mockStopOrder.On("GetActiveOrders").Return(nil, assert.AnError)

		monitor := NewOrderMonitor(mockStopOrder, mockExchange, time.Second, false)

		// Act
		err := monitor.checkOrders()

		// Assert
		assert.Error(t, err)
		mockStopOrder.AssertExpectations(t)
	})
}
