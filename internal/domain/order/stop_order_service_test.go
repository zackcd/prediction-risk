package order

import (
	"prediction-risk/internal/infrastructure/external/kalshi"
	"testing"

	"prediction-risk/internal/domain/contract"
	"prediction-risk/internal/domain/core"
	"prediction-risk/internal/domain/order/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Test suite
func TestStopOrderService(t *testing.T) {
	t.Run("GetOrder", func(t *testing.T) {
		t.Run("returns order when found", func(t *testing.T) {
			// Arrange
			mockRepo := new(mocks.MockStopOrderRepo)
			mockExchange := new(mocks.MockExchangeProvider)
			service := NewStopOrderService(mockRepo, mockExchange)

			threshold, err := contract.NewContractPrice(100)
			assert.NoError(t, err)
			expectedOrder := order.NewStopOrder("AAPL", contract.SideYes, threshold, nil, nil)

			mockRepo.On("GetByID", expectedOrder.ID()).Return(expectedOrder, nil)

			// Act
			order, err := service.GetOrder(expectedOrder.ID())

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, expectedOrder, order)
			mockRepo.AssertExpectations(t)
		})

		t.Run("returns nil when not found", func(t *testing.T) {
			mockRepo := new(mocks.MockStopOrderRepo)
			mockExchange := new(mocks.MockExchangeProvider)
			service := NewStopOrderService(mockRepo, mockExchange)

			orderID := order.NewOrderID()
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
			mockExchange := new(mocks.MockExchangeProvider)
			service := NewStopOrderService(mockRepo, mockExchange)

			ticker := "AAPL"
			threshold, err := contract.NewContractPrice(100)
			assert.NoError(t, err)

			mockRepo.On("Persist", mock.MatchedBy(func(order *order.StopOrder) bool {
				return order.Ticker() == ticker &&
					order.TriggerPrice() == threshold &&
					order.Status() == order.OrderStatusActive
			})).Return(nil)

			newOrder, err := service.CreateOrder(ticker, contract.SideYes, threshold, nil)

			assert.NoError(t, err)
			assert.Equal(t, ticker, newOrder.Ticker())
			assert.Equal(t, threshold, newOrder.TriggerPrice())
			assert.Equal(t, order.OrderStatusActive, newOrder.Status())
			mockRepo.AssertExpectations(t)
		})

		t.Run("creates order successfully with limit price", func(t *testing.T) {
			mockRepo := new(mocks.MockStopOrderRepo)
			mockExchange := new(mocks.MockExchangeProvider)
			service := NewStopOrderService(mockRepo, mockExchange)

			ticker := "AAPL"
			threshold, err := contract.NewContractPrice(100)
			assert.NoError(t, err)
			limitPrice, err := contract.NewContractPrice(95)
			assert.NoError(t, err)

			mockRepo.On("Persist", mock.MatchedBy(func(order *order.StopOrder) bool {
				return order.Ticker() == ticker &&
					order.TriggerPrice() == threshold &&
					*order.LimitPrice() == limitPrice &&
					order.Status() == order.OrderStatusActive
			})).Return(nil)

			newOrder, err := service.CreateOrder(ticker, contract.SideYes, threshold, &limitPrice)

			assert.NoError(t, err)
			assert.Equal(t, ticker, newOrder.Ticker())
			assert.Equal(t, threshold, newOrder.TriggerPrice())
			assert.Equal(t, &limitPrice, newOrder.LimitPrice())
			assert.Equal(t, order.OrderStatusActive, newOrder.Status())
			mockRepo.AssertExpectations(t)
		})
	})

	t.Run("UpdateOrder", func(t *testing.T) {
		t.Run("updates order successfully", func(t *testing.T) {
			mockRepo := new(mocks.MockStopOrderRepo)
			mockExchange := new(mocks.MockExchangeProvider)
			service := NewStopOrderService(mockRepo, mockExchange)

			threshold, err := contract.NewContractPrice(80)
			assert.NoError(t, err)

			newThreshold, err := contract.NewContractPrice(100)
			assert.NoError(t, err)
			existingOrder := order.NewStopOrder("AAPL", contract.SideYes, threshold, nil, nil)

			mockRepo.On("GetByID", existingOrder.ID()).Return(existingOrder, nil)
			mockRepo.On("Persist", mock.MatchedBy(func(order *order.StopOrder) bool {
				return order.TriggerPrice() == newThreshold
			})).Return(nil)

			order, err := service.UpdateOrder(existingOrder.ID(), &newThreshold, nil)

			assert.NoError(t, err)
			assert.Equal(t, newThreshold, order.TriggerPrice())
			mockRepo.AssertExpectations(t)
		})
	})

	t.Run("CancelOrder", func(t *testing.T) {
		t.Run("cancels active order successfully", func(t *testing.T) {
			mockRepo := new(mocks.MockStopOrderRepo)
			mockExchange := new(mocks.MockExchangeProvider)
			service := NewStopOrderService(mockRepo, mockExchange)

			orderID := order.NewOrderID()
			threshold, err := contract.NewContractPrice(100)
			assert.NoError(t, err)
			existingOrder := order.NewStopOrder("AAPL", contract.SideYes, threshold, nil, &orderID)

			mockRepo.On("GetByID", orderID).Return(existingOrder, nil)
			mockRepo.On("Persist", mock.MatchedBy(func(order *order.StopOrder) bool {
				return order.Status() == order.OrderStatusCancelled
			})).Return(nil)

			cancelledOrder, err := service.CancelOrder(orderID)

			assert.NoError(t, err)
			assert.Equal(t, order.OrderStatusCancelled, cancelledOrder.Status())
			mockRepo.AssertExpectations(t)
		})

		t.Run("fails when order not found", func(t *testing.T) {
			mockRepo := new(mocks.MockStopOrderRepo)
			mockExchange := new(mocks.MockExchangeProvider)
			service := NewStopOrderService(mockRepo, mockExchange)

			orderID := order.NewOrderID()
			errResult := core.NewErrNotFound("OrderId", orderID.String())
			mockRepo.On("GetByID", orderID).Return(nil, errResult)

			order, err := service.CancelOrder(orderID)

			assert.Error(t, err)
			assert.Nil(t, order)
			assert.IsType(t, errResult, err)
			mockRepo.AssertExpectations(t)
		})

		t.Run("fails when order already cancelled", func(t *testing.T) {
			mockRepo := new(mocks.MockStopOrderRepo)
			mockExchange := new(mocks.MockExchangeProvider)
			service := NewStopOrderService(mockRepo, mockExchange)

			orderID := order.NewOrderID()
			threshold, err := contract.NewContractPrice(100)
			assert.NoError(t, err)
			existingOrder := order.NewStopOrder("AAPL", contract.SideYes, threshold, nil, &orderID)
			existingOrder.UpdateStatus(order.OrderStatusCancelled)

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
			mockExchange := new(mocks.MockExchangeProvider)
			service := NewStopOrderService(mockRepo, mockExchange)

			threshold, _ := contract.NewContractPrice(100)
			activeOrder := order.NewStopOrder("AAPL", contract.SideYes, threshold, nil, nil)
			cancelledOrder := order.NewStopOrder("MSFT", contract.SideYes, threshold, nil, nil)
			cancelledOrder.UpdateStatus(order.OrderStatusCancelled)

			allOrders := []*order.StopOrder{activeOrder, cancelledOrder}
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
			mockExchange := new(mocks.MockExchangeProvider)
			service := NewStopOrderService(mockRepo, mockExchange)

			threshold, _ := contract.NewContractPrice(100)
			existingOrder := order.NewStopOrder("AAPL", contract.SideYes, threshold, nil, nil)

			mockRepo.On("GetByID", existingOrder.ID()).Return(existingOrder, nil)
			mockExchange.On("GetPositions").Return(&kalshi.PositionsResult{
				MarketPositions: []kalshi.MarketPosition{
					{Ticker: "AAPL", Position: 10},
				},
			}, nil)
			mockExchange.On("CreateSellOrder",
				"AAPL",
				10,
				contract.SideYes,
				existingOrder.ID().String(),
				(*contract.ContractPrice)(nil),
			).Return(&order.ExchangeOrder{}, nil)
			mockRepo.On("Persist", mock.MatchedBy(func(order *order.StopOrder) bool {
				return order.Status() == order.OrderStatusTriggered
			})).Return(nil)

			executedOrder, err := service.ExecuteOrder(existingOrder.ID(), false)

			assert.NoError(t, err)
			assert.Equal(t, order.OrderStatusTriggered, executedOrder.Status())
			mockRepo.AssertExpectations(t)
			mockExchange.AssertExpectations(t)
		})

		t.Run("executes order successfully with limit price", func(t *testing.T) {
			mockRepo := new(mocks.MockStopOrderRepo)
			mockExchange := new(mocks.MockExchangeProvider)
			service := NewStopOrderService(mockRepo, mockExchange)

			threshold, _ := contract.NewContractPrice(100)
			limitPrice, _ := contract.NewContractPrice(95)
			existingOrder := order.NewStopOrder("AAPL", contract.SideYes, threshold, &limitPrice, nil)

			mockRepo.On("GetByID", existingOrder.ID()).Return(existingOrder, nil)
			mockExchange.On("GetPositions").Return(&kalshi.PositionsResult{
				MarketPositions: []kalshi.MarketPosition{
					{Ticker: "AAPL", Position: 10},
				},
			}, nil)
			mockExchange.On("CreateSellOrder",
				"AAPL",
				10,
				contract.SideYes,
				existingOrder.ID().String(),
				existingOrder.LimitPrice(),
			).Return(&order.ExchangeOrder{}, nil)
			mockRepo.On("Persist", mock.MatchedBy(func(order *order.StopOrder) bool {
				return order.Status() == order.OrderStatusTriggered
			})).Return(nil)

			executedOrder, err := service.ExecuteOrder(existingOrder.ID(), false)

			assert.NoError(t, err)
			assert.Equal(t, order.OrderStatusTriggered, executedOrder.Status())
			mockRepo.AssertExpectations(t)
			mockExchange.AssertExpectations(t)
		})

		t.Run("fails when no position found", func(t *testing.T) {
			mockRepo := new(mocks.MockStopOrderRepo)
			mockExchange := new(mocks.MockExchangeProvider)
			service := NewStopOrderService(mockRepo, mockExchange)

			threshold, _ := contract.NewContractPrice(100)
			existingOrder := order.NewStopOrder("AAPL", contract.SideYes, threshold, nil, nil)

			mockRepo.On("GetByID", existingOrder.ID()).Return(existingOrder, nil)
			mockExchange.On("GetPositions").Return(&kalshi.PositionsResult{
				EventPositions:  []kalshi.EventPosition{},
				MarketPositions: []kalshi.MarketPosition{},
			}, nil)

			order, err := service.ExecuteOrder(existingOrder.ID(), false)

			assert.Error(t, err)
			assert.Nil(t, order)
			assert.Contains(t, err.Error(), "no position found")
			mockRepo.AssertExpectations(t)
			mockExchange.AssertExpectations(t)
		})
	})
}
