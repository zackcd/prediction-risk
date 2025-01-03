package services

import (
	"prediction-risk/internal/domain/entities"
	"prediction-risk/internal/domain/services/mocks"
	"prediction-risk/internal/infrastructure/external/kalshi"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Test suite
func TestStopOrderService(t *testing.T) {
	t.Run("GetOrder", func(t *testing.T) {
		t.Run("returns order when found", func(t *testing.T) {
			// Arrange
			mockRepo := new(mocks.MockStopOrderRepo)
			mockExchange := new(mocks.MockExchangeService)
			service := NewStopOrderService(mockRepo, mockExchange)

			orderID := entities.NewOrderID()
			threshold, err := entities.NewContractPrice(100)
			assert.NoError(t, err)
			expectedOrder := entities.NewStopOrder("AAPL", entities.SideYes, threshold, nil, &orderID)

			mockRepo.On("GetByID", orderID).Return(expectedOrder, nil)

			// Act
			order, err := service.GetOrder(orderID)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, expectedOrder, order)
			mockRepo.AssertExpectations(t)
		})

		t.Run("returns nil when not found", func(t *testing.T) {
			mockRepo := new(mocks.MockStopOrderRepo)
			mockExchange := new(mocks.MockExchangeService)
			service := NewStopOrderService(mockRepo, mockExchange)

			orderID := entities.NewOrderID()
			mockRepo.On("GetByID", orderID).Return(nil, nil)

			order, err := service.GetOrder(orderID)

			assert.NoError(t, err)
			assert.Nil(t, order)
			mockRepo.AssertExpectations(t)
		})
	})

	t.Run("CreateOrder", func(t *testing.T) {
		t.Run("creates order successfully", func(t *testing.T) {
			mockRepo := new(mocks.MockStopOrderRepo)
			mockExchange := new(mocks.MockExchangeService)
			service := NewStopOrderService(mockRepo, mockExchange)

			ticker := "AAPL"
			threshold, err := entities.NewContractPrice(100)
			assert.NoError(t, err)

			mockRepo.On("Persist", mock.MatchedBy(func(order *entities.StopOrder) bool {
				return order.Ticker() == ticker &&
					order.TriggerPrice() == threshold &&
					order.Status() == entities.OrderStatusActive
			})).Return(nil)

			order, err := service.CreateOrder(ticker, entities.SideYes, threshold, nil)

			assert.NoError(t, err)
			assert.Equal(t, ticker, order.Ticker())
			assert.Equal(t, threshold, order.TriggerPrice())
			assert.Equal(t, entities.OrderStatusActive, order.Status())
			mockRepo.AssertExpectations(t)
		})

		t.Run("creates order successfully with limit price", func(t *testing.T) {
			mockRepo := new(mocks.MockStopOrderRepo)
			mockExchange := new(mocks.MockExchangeService)
			service := NewStopOrderService(mockRepo, mockExchange)

			ticker := "AAPL"
			threshold, err := entities.NewContractPrice(100)
			assert.NoError(t, err)
			limitPrice, err := entities.NewContractPrice(95)
			assert.NoError(t, err)

			mockRepo.On("Persist", mock.MatchedBy(func(order *entities.StopOrder) bool {
				return order.Ticker() == ticker &&
					order.TriggerPrice() == threshold &&
					*order.LimitPrice() == limitPrice &&
					order.Status() == entities.OrderStatusActive
			})).Return(nil)

			order, err := service.CreateOrder(ticker, entities.SideYes, threshold, &limitPrice)

			assert.NoError(t, err)
			assert.Equal(t, ticker, order.Ticker())
			assert.Equal(t, threshold, order.TriggerPrice())
			assert.Equal(t, &limitPrice, order.LimitPrice())
			assert.Equal(t, entities.OrderStatusActive, order.Status())
			mockRepo.AssertExpectations(t)
		})
	})

	t.Run("UpdateOrder", func(t *testing.T) {
		t.Run("updates order successfully", func(t *testing.T) {
			mockRepo := new(mocks.MockStopOrderRepo)
			mockExchange := new(mocks.MockExchangeService)
			service := NewStopOrderService(mockRepo, mockExchange)

			orderID := entities.NewOrderID()
			threshold, err := entities.NewContractPrice(80)
			assert.NoError(t, err)

			newThreshold, err := entities.NewContractPrice(100)
			assert.NoError(t, err)
			existingOrder := entities.NewStopOrder("AAPL", entities.SideYes, threshold, nil, &orderID)

			mockRepo.On("GetByID", orderID).Return(existingOrder, nil)
			mockRepo.On("Persist", mock.MatchedBy(func(order *entities.StopOrder) bool {
				return order.TriggerPrice() == newThreshold
			})).Return(nil)

			order, err := service.UpdateOrder(orderID, &newThreshold, nil)

			assert.NoError(t, err)
			assert.Equal(t, newThreshold, order.TriggerPrice())
			mockRepo.AssertExpectations(t)
		})
	})

	t.Run("CancelOrder", func(t *testing.T) {
		t.Run("cancels active order successfully", func(t *testing.T) {
			mockRepo := new(mocks.MockStopOrderRepo)
			mockExchange := new(mocks.MockExchangeService)
			service := NewStopOrderService(mockRepo, mockExchange)

			orderID := entities.NewOrderID()
			threshold, err := entities.NewContractPrice(100)
			assert.NoError(t, err)
			existingOrder := entities.NewStopOrder("AAPL", entities.SideYes, threshold, nil, &orderID)

			mockRepo.On("GetByID", orderID).Return(existingOrder, nil)
			mockRepo.On("Persist", mock.MatchedBy(func(order *entities.StopOrder) bool {
				return order.Status() == entities.OrderStatusCancelled
			})).Return(nil)

			order, err := service.CancelOrder(orderID)

			assert.NoError(t, err)
			assert.Equal(t, entities.OrderStatusCancelled, order.Status())
			mockRepo.AssertExpectations(t)
		})

		t.Run("fails when order not found", func(t *testing.T) {
			mockRepo := new(mocks.MockStopOrderRepo)
			mockExchange := new(mocks.MockExchangeService)
			service := NewStopOrderService(mockRepo, mockExchange)

			orderID := entities.NewOrderID()
			errResult := entities.NewErrNotFound("OrderId", orderID.String())
			mockRepo.On("GetByID", orderID).Return(nil, errResult)

			order, err := service.CancelOrder(orderID)

			assert.Error(t, err)
			assert.Nil(t, order)
			assert.IsType(t, errResult, err)
			mockRepo.AssertExpectations(t)
		})

		t.Run("fails when order already cancelled", func(t *testing.T) {
			mockRepo := new(mocks.MockStopOrderRepo)
			mockExchange := new(mocks.MockExchangeService)
			service := NewStopOrderService(mockRepo, mockExchange)

			orderID := entities.NewOrderID()
			threshold, err := entities.NewContractPrice(100)
			assert.NoError(t, err)
			existingOrder := entities.NewStopOrder("AAPL", entities.SideYes, threshold, nil, &orderID)
			existingOrder.UpdateStatus(entities.OrderStatusCancelled)

			mockRepo.On("GetByID", orderID).Return(existingOrder, nil)

			order, err := service.CancelOrder(orderID)

			assert.Error(t, err)
			assert.Nil(t, order)
			assert.Contains(t, err.Error(), "invalid status")
			mockRepo.AssertExpectations(t)
		})
	})

	t.Run("GetActiveOrders", func(t *testing.T) {
		t.Run("returns only active orders", func(t *testing.T) {
			mockRepo := new(mocks.MockStopOrderRepo)
			mockExchange := new(mocks.MockExchangeService)
			service := NewStopOrderService(mockRepo, mockExchange)

			threshold, _ := entities.NewContractPrice(100)
			activeOrder := entities.NewStopOrder("AAPL", entities.SideYes, threshold, nil, nil)
			cancelledOrder := entities.NewStopOrder("MSFT", entities.SideYes, threshold, nil, nil)
			cancelledOrder.UpdateStatus(entities.OrderStatusCancelled)

			allOrders := []*entities.StopOrder{activeOrder, cancelledOrder}
			mockRepo.On("GetAll").Return(allOrders, nil)

			orders, err := service.GetActiveOrders()

			assert.NoError(t, err)
			assert.Len(t, orders, 1)
			assert.Equal(t, activeOrder, orders[0])
			mockRepo.AssertExpectations(t)
		})
	})

	t.Run("ExecuteOrder", func(t *testing.T) {
		t.Run("executes order successfully with market price", func(t *testing.T) {
			mockRepo := new(mocks.MockStopOrderRepo)
			mockExchange := new(mocks.MockExchangeService)
			service := NewStopOrderService(mockRepo, mockExchange)

			orderID := entities.NewOrderID()
			threshold, _ := entities.NewContractPrice(100)
			existingOrder := entities.NewStopOrder("AAPL", entities.SideYes, threshold, nil, &orderID)

			mockRepo.On("GetByID", orderID).Return(existingOrder, nil)
			mockExchange.On("GetPositions").Return(&kalshi.PositionsResult{
				MarketPositions: []kalshi.MarketPosition{
					{Ticker: "AAPL", Position: 10},
				},
			}, nil)
			mockExchange.On("CreateSellOrder",
				"AAPL",
				10,
				entities.SideYes,
				orderID.String(),
				(*entities.ContractPrice)(nil),
			).Return(&entities.ExchangeOrder{}, nil)
			mockRepo.On("Persist", mock.MatchedBy(func(order *entities.StopOrder) bool {
				return order.Status() == entities.OrderStatusTriggered
			})).Return(nil)

			order, err := service.ExecuteOrder(orderID, false)

			assert.NoError(t, err)
			assert.Equal(t, entities.OrderStatusTriggered, order.Status())
			mockRepo.AssertExpectations(t)
			mockExchange.AssertExpectations(t)
		})

		t.Run("executes order successfully with limit price", func(t *testing.T) {
			mockRepo := new(mocks.MockStopOrderRepo)
			mockExchange := new(mocks.MockExchangeService)
			service := NewStopOrderService(mockRepo, mockExchange)

			orderID := entities.NewOrderID()
			threshold, _ := entities.NewContractPrice(100)
			limitPrice, _ := entities.NewContractPrice(95)
			existingOrder := entities.NewStopOrder("AAPL", entities.SideYes, threshold, &limitPrice, &orderID)

			mockRepo.On("GetByID", orderID).Return(existingOrder, nil)
			mockExchange.On("GetPositions").Return(&kalshi.PositionsResult{
				MarketPositions: []kalshi.MarketPosition{
					{Ticker: "AAPL", Position: 10},
				},
			}, nil)
			mockExchange.On("CreateSellOrder",
				"AAPL",
				10,
				entities.SideYes,
				orderID.String(),
				existingOrder.LimitPrice(),
			).Return(&entities.ExchangeOrder{}, nil)
			mockRepo.On("Persist", mock.MatchedBy(func(order *entities.StopOrder) bool {
				return order.Status() == entities.OrderStatusTriggered
			})).Return(nil)

			order, err := service.ExecuteOrder(orderID, false)

			assert.NoError(t, err)
			assert.Equal(t, entities.OrderStatusTriggered, order.Status())
			mockRepo.AssertExpectations(t)
			mockExchange.AssertExpectations(t)
		})

		t.Run("fails when no position found", func(t *testing.T) {
			mockRepo := new(mocks.MockStopOrderRepo)
			mockExchange := new(mocks.MockExchangeService)
			service := NewStopOrderService(mockRepo, mockExchange)

			orderID := entities.NewOrderID()
			threshold, _ := entities.NewContractPrice(100)
			existingOrder := entities.NewStopOrder("AAPL", entities.SideYes, threshold, nil, &orderID)

			mockRepo.On("GetByID", orderID).Return(existingOrder, nil)
			mockExchange.On("GetPositions").Return(&kalshi.PositionsResult{
				EventPositions:  []kalshi.EventPosition{},
				MarketPositions: []kalshi.MarketPosition{},
			}, nil)

			order, err := service.ExecuteOrder(orderID, false)

			assert.Error(t, err)
			assert.Nil(t, order)
			assert.Contains(t, err.Error(), "no position found")
			mockRepo.AssertExpectations(t)
			mockExchange.AssertExpectations(t)
		})
	})
}
