package services

import (
	"prediction-risk/internal/domain"
	"prediction-risk/internal/domain/entities"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock repository
type MockStopLossOrderRepo struct {
	mock.Mock
}

func (m *MockStopLossOrderRepo) GetByID(id uuid.UUID) (*entities.StopLossOrder, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.StopLossOrder), args.Error(1)
}

func (m *MockStopLossOrderRepo) Persist(order *entities.StopLossOrder) error {
	args := m.Called(order)
	return args.Error(0)
}

func (m *MockStopLossOrderRepo) GetAll() ([]*entities.StopLossOrder, error) {
	args := m.Called()
	return args.Get(0).([]*entities.StopLossOrder), args.Error(1)
}

// Test suite
func TestStopLossService(t *testing.T) {
	t.Run("GetOrder", func(t *testing.T) {
		t.Run("returns order when found", func(t *testing.T) {
			// Arrange
			mockRepo := new(MockStopLossOrderRepo)
			service := NewStopLossService(mockRepo)

			orderID := uuid.New()
			expectedOrder := entities.NewStopLossOrder("AAPL", entities.SideYes, entities.ContractPrice(100.0))

			mockRepo.On("GetByID", orderID).Return(expectedOrder, nil)

			// Act
			order, err := service.GetOrder(orderID)

			// Assert
			assert.NoError(t, err)
			assert.Equal(t, expectedOrder, order)
			mockRepo.AssertExpectations(t)
		})

		t.Run("returns nil when not found", func(t *testing.T) {
			mockRepo := new(MockStopLossOrderRepo)
			service := NewStopLossService(mockRepo)

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
			mockRepo := new(MockStopLossOrderRepo)
			service := NewStopLossService(mockRepo)

			ticker := "AAPL"
			threshold := entities.ContractPrice(100.0)

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
			mockRepo := new(MockStopLossOrderRepo)
			service := NewStopLossService(mockRepo)

			orderID := uuid.New()
			existingOrder := entities.NewStopLossOrder("AAPL", entities.SideYes, entities.ContractPrice(100.0))
			newThreshold := entities.ContractPrice(120.0)

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
			mockRepo := new(MockStopLossOrderRepo)
			service := NewStopLossService(mockRepo)

			orderID := uuid.New()
			existingOrder := entities.NewStopLossOrder("AAPL", entities.SideYes, entities.ContractPrice(100.0))

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
			mockRepo := new(MockStopLossOrderRepo)
			service := NewStopLossService(mockRepo)

			orderID := uuid.New()
			mockRepo.On("GetByID", orderID).Return(nil, nil)

			order, err := service.CancelOrder(orderID)

			assert.Error(t, err)
			assert.Nil(t, order)
			assert.IsType(t, &domain.ErrNotFound{}, err)
			mockRepo.AssertExpectations(t)
		})

		t.Run("fails when order already canceled", func(t *testing.T) {
			mockRepo := new(MockStopLossOrderRepo)
			service := NewStopLossService(mockRepo)

			orderID := uuid.New()
			existingOrder := entities.NewStopLossOrder("AAPL", entities.SideYes, entities.ContractPrice(100.0))
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
