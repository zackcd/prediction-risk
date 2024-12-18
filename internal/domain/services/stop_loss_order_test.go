package services

import (
	"prediction-risk/internal/domain/entities"
	"prediction-risk/internal/domain/services/mocks"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Test suite
func TestStopLossService(t *testing.T) {
	t.Run("GetOrder", func(t *testing.T) {
		t.Run("returns order when found", func(t *testing.T) {
			// Arrange
			mockRepo := new(mocks.MockStopLossOrderRepo)
			mockExchange := new(mocks.MockExchangeService)
			service := NewStopLossService(mockRepo, mockExchange)

			orderID := uuid.New()
			threshold, err := entities.NewContractPrice(100)
			assert.NoError(t, err)
			expectedOrder := entities.NewStopLossOrder("AAPL", entities.SideYes, threshold)

			mockRepo.On("GetByID", orderID).Return(expectedOrder, nil)

			// Act
			order, err := service.GetOrder(orderID)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, expectedOrder, order)
			mockRepo.AssertExpectations(t)
		})

		t.Run("returns nil when not found", func(t *testing.T) {
			mockRepo := new(mocks.MockStopLossOrderRepo)
			mockExchange := new(mocks.MockExchangeService)
			service := NewStopLossService(mockRepo, mockExchange)

			orderID := uuid.New()
			mockRepo.On("GetByID", orderID).Return(nil, nil)

			order, err := service.GetOrder(orderID)

			assert.NoError(t, err)
			assert.Nil(t, order)
			mockRepo.AssertExpectations(t)
		})
	})

	t.Run("CreateOrder", func(t *testing.T) {
		t.Run("creates order successfully", func(t *testing.T) {
			mockRepo := new(mocks.MockStopLossOrderRepo)
			mockExchange := new(mocks.MockExchangeService)
			service := NewStopLossService(mockRepo, mockExchange)

			ticker := "AAPL"
			threshold, err := entities.NewContractPrice(100)
			assert.NoError(t, err)

			mockRepo.On("Persist", mock.MatchedBy(func(order *entities.StopLossOrder) bool {
				return order.Ticker() == ticker &&
					order.Threshold() == threshold &&
					order.Status() == entities.StatusActive
			})).Return(nil)

			order, err := service.CreateOrder(ticker, entities.SideYes, threshold)

			assert.NoError(t, err)
			assert.Equal(t, ticker, order.Ticker())
			assert.Equal(t, threshold, order.Threshold())
			assert.Equal(t, entities.StatusActive, order.Status())
			mockRepo.AssertExpectations(t)
		})
	})

	t.Run("UpdateOrder", func(t *testing.T) {
		t.Run("updates order successfully", func(t *testing.T) {
			mockRepo := new(mocks.MockStopLossOrderRepo)
			mockExchange := new(mocks.MockExchangeService)
			service := NewStopLossService(mockRepo, mockExchange)

			orderID := uuid.New()
			threshold, err := entities.NewContractPrice(80)
			assert.NoError(t, err)

			newThreshold, err := entities.NewContractPrice(100)
			assert.NoError(t, err)
			existingOrder := entities.NewStopLossOrder("AAPL", entities.SideYes, threshold)

			mockRepo.On("GetByID", orderID).Return(existingOrder, nil)
			mockRepo.On("Persist", mock.MatchedBy(func(order *entities.StopLossOrder) bool {
				return order.Threshold() == newThreshold
			})).Return(nil)

			order, err := service.UpdateOrder(orderID, newThreshold)

			assert.NoError(t, err)
			assert.Equal(t, newThreshold, order.Threshold())
			mockRepo.AssertExpectations(t)
		})
	})

	t.Run("CancelOrder", func(t *testing.T) {
		t.Run("cancels active order successfully", func(t *testing.T) {
			mockRepo := new(mocks.MockStopLossOrderRepo)
			mockExchange := new(mocks.MockExchangeService)
			service := NewStopLossService(mockRepo, mockExchange)

			orderID := uuid.New()
			threshold, err := entities.NewContractPrice(100)
			assert.NoError(t, err)
			existingOrder := entities.NewStopLossOrder("AAPL", entities.SideYes, threshold)

			mockRepo.On("GetByID", orderID).Return(existingOrder, nil)
			mockRepo.On("Persist", mock.MatchedBy(func(order *entities.StopLossOrder) bool {
				return order.Status() == entities.StatusCanceled
			})).Return(nil)

			order, err := service.CancelOrder(orderID)

			assert.NoError(t, err)
			assert.Equal(t, entities.StatusCanceled, order.Status())
			mockRepo.AssertExpectations(t)
		})

		t.Run("fails when order not found", func(t *testing.T) {
			mockRepo := new(mocks.MockStopLossOrderRepo)
			mockExchange := new(mocks.MockExchangeService)
			service := NewStopLossService(mockRepo, mockExchange)

			orderID := uuid.New()
			mockRepo.On("GetByID", orderID).Return(nil, nil)

			order, err := service.CancelOrder(orderID)

			assert.Error(t, err)
			assert.Nil(t, order)
			assert.IsType(t, &entities.ErrNotFound{}, err)
			mockRepo.AssertExpectations(t)
		})

		t.Run("fails when order already canceled", func(t *testing.T) {
			mockRepo := new(mocks.MockStopLossOrderRepo)
			mockExchange := new(mocks.MockExchangeService)
			service := NewStopLossService(mockRepo, mockExchange)

			orderID := uuid.New()
			threshold, err := entities.NewContractPrice(100)
			assert.NoError(t, err)
			existingOrder := entities.NewStopLossOrder("AAPL", entities.SideYes, threshold)
			existingOrder.SetStatus(entities.StatusCanceled)

			mockRepo.On("GetByID", orderID).Return(existingOrder, nil)

			order, err := service.CancelOrder(orderID)

			assert.Error(t, err)
			assert.Nil(t, order)
			assert.Contains(t, err.Error(), "invalid status")
			mockRepo.AssertExpectations(t)
		})
	})
}
