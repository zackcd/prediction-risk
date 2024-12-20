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
		mockStopLoss := new(mocks.MockStopLossService)
		mockExchange := new(mocks.MockExchangeService)

		threshold, _ := entities.NewContractPrice(60)
		order := entities.NewStopLossOrder("MARKET-1", entities.SideYes, threshold)

		market := &kalshi.Market{
			Ticker: "MARKET-1",
			YesBid: 55, // Below threshold
			YesAsk: 56,
		}

		mockStopLoss.On("GetActiveOrders").Return([]*entities.StopLossOrder{order}, nil)
		mockStopLoss.On("ExecuteOrder", order.ID(), true).Return(order, nil)
		mockExchange.On("GetMarket", "MARKET-1").Return(market, nil)

		monitor := NewStopLossMonitor(mockStopLoss, mockExchange, time.Second)

		// Act
		err := monitor.checkOrders(true)

		// Assert
		assert.NoError(t, err)
		mockStopLoss.AssertExpectations(t)
		mockExchange.AssertExpectations(t)
	})

	t.Run("should execute NO stop loss when bid drops below threshold", func(t *testing.T) {
		// Arrange
		mockStopLoss := new(mocks.MockStopLossService)
		mockExchange := new(mocks.MockExchangeService)

		threshold, _ := entities.NewContractPrice(60)
		order := entities.NewStopLossOrder("MARKET-1", entities.SideNo, threshold)

		market := &kalshi.Market{
			Ticker: "MARKET-1",
			NoBid:  55, // Below threshold
			NoAsk:  56,
		}

		mockStopLoss.On("GetActiveOrders").Return([]*entities.StopLossOrder{order}, nil)
		mockStopLoss.On("ExecuteOrder", order.ID(), true).Return(order, nil)
		mockExchange.On("GetMarket", "MARKET-1").Return(market, nil)

		monitor := NewStopLossMonitor(mockStopLoss, mockExchange, time.Second)

		// Act
		err := monitor.checkOrders(true)

		// Assert
		assert.NoError(t, err)
		mockStopLoss.AssertExpectations(t)
		mockExchange.AssertExpectations(t)
	})

	t.Run("should not execute when bid is above threshold", func(t *testing.T) {
		// Arrange
		mockStopLoss := new(mocks.MockStopLossService)
		mockExchange := new(mocks.MockExchangeService)

		threshold, _ := entities.NewContractPrice(60)
		order := entities.NewStopLossOrder("MARKET-1", entities.SideYes, threshold)

		market := &kalshi.Market{
			Ticker: "MARKET-1",
			YesBid: 65, // Above threshold
			YesAsk: 66,
		}

		mockStopLoss.On("GetActiveOrders").Return([]*entities.StopLossOrder{order}, nil)
		mockExchange.On("GetMarket", "MARKET-1").Return(market, nil)
		// Note: ExecuteOrder should not be called

		monitor := NewStopLossMonitor(mockStopLoss, mockExchange, time.Second)

		// Act
		err := monitor.checkOrders(true)

		// Assert
		assert.NoError(t, err)
		mockStopLoss.AssertExpectations(t)
		mockExchange.AssertExpectations(t)
	})

	t.Run("should handle market fetch error", func(t *testing.T) {
		// Arrange
		mockStopLoss := new(mocks.MockStopLossService)
		mockExchange := new(mocks.MockExchangeService)

		threshold, _ := entities.NewContractPrice(60)
		order := entities.NewStopLossOrder("MARKET-1", entities.SideYes, threshold)

		mockStopLoss.On("GetActiveOrders").Return([]*entities.StopLossOrder{order}, nil)
		mockExchange.On("GetMarket", "MARKET-1").Return(nil, assert.AnError)

		monitor := NewStopLossMonitor(mockStopLoss, mockExchange, time.Second)

		// Act
		err := monitor.checkOrders(true)

		// Assert
		assert.NoError(t, err) // Should not return error as it continues processing
		mockStopLoss.AssertExpectations(t)
		mockExchange.AssertExpectations(t)
	})

	t.Run("should handle GetActiveOrders error", func(t *testing.T) {
		// Arrange
		mockStopLoss := new(mocks.MockStopLossService)
		mockExchange := new(mocks.MockExchangeService)

		mockStopLoss.On("GetActiveOrders").Return(nil, assert.AnError)

		monitor := NewStopLossMonitor(mockStopLoss, mockExchange, time.Second)

		// Act
		err := monitor.checkOrders(true)

		// Assert
		assert.Error(t, err)
		mockStopLoss.AssertExpectations(t)
	})
}
