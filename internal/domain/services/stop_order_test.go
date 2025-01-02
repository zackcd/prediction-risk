package services

import (
	"prediction-risk/internal/domain/entities"
	"prediction-risk/internal/domain/services/mocks"
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
			mockExecutor := new(mocks.MockOrderExecutor)
			service := NewStopOrderService(mockRepo, mockExecutor)

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
			mockExecutor := new(mocks.MockOrderExecutor)
			service := NewStopOrderService(mockRepo, mockExecutor)

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
			mockExecutor := new(mocks.MockOrderExecutor)
			service := NewStopOrderService(mockRepo, mockExecutor)

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
	})

	t.Run("UpdateOrder", func(t *testing.T) {
		t.Run("updates order successfully", func(t *testing.T) {
			mockRepo := new(mocks.MockStopOrderRepo)
			mockExecutor := new(mocks.MockOrderExecutor)
			service := NewStopOrderService(mockRepo, mockExecutor)

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
			mockExecutor := new(mocks.MockOrderExecutor)
			service := NewStopOrderService(mockRepo, mockExecutor)

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
			mockExecutor := new(mocks.MockOrderExecutor)
			service := NewStopOrderService(mockRepo, mockExecutor)

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
			mockExecutor := new(mocks.MockOrderExecutor)
			service := NewStopOrderService(mockRepo, mockExecutor)

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
}
