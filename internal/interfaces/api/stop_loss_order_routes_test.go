package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"prediction-risk/internal/domain/entities"
	"testing"

	"github.com/go-chi/chi"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockStopLossService struct {
	mock.Mock
}

func (m *MockStopLossService) CreateOrder(
	ticker string,
	side entities.Side,
	threshold entities.ContractPrice,
) (*entities.StopLossOrder, error) {
	args := m.Called(ticker, side, threshold)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.StopLossOrder), args.Error(1)
}

func (m *MockStopLossService) GetOrder(
	stopLossOrderId uuid.UUID,
) (*entities.StopLossOrder, error) {
	args := m.Called(stopLossOrderId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.StopLossOrder), args.Error(1)
}

func (m *MockStopLossService) UpdateOrder(
	stopLossOrderId uuid.UUID,
	threshold entities.ContractPrice,
) (
	*entities.StopLossOrder,
	error,
) {
	args := m.Called(stopLossOrderId, threshold)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.StopLossOrder), args.Error(1)
}

func (m *MockStopLossService) CancelOrder(
	stopLossOrderId uuid.UUID,
) (
	*entities.StopLossOrder,
	error,
) {
	args := m.Called(stopLossOrderId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.StopLossOrder), args.Error(1)
}

func (m *MockStopLossService) GetActiveOrders() ([]*entities.StopLossOrder, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.StopLossOrder), args.Error(1)
}

func (m *MockStopLossService) ExecuteOrder(
	stopLossOrderId uuid.UUID,
) (
	*entities.StopLossOrder,
	error,
) {
	args := m.Called(stopLossOrderId)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.StopLossOrder), args.Error(1)
}

// Implement other service methods...

func setupTest() (*chi.Mux, *MockStopLossService) {
	mockService := new(MockStopLossService)
	routes := NewStopLossRoutes(mockService)

	router := chi.NewRouter()
	routes.Register(router)

	return router, mockService
}

func TestStopLossRoutes(t *testing.T) {
	t.Run("CreateStopLoss", func(t *testing.T) {
		t.Run("successfully creates order", func(t *testing.T) {
			// Arrange
			router, mockService := setupTest()

			expectedOrder := &entities.StopLossOrder{} // Create with appropriate test data
			threshold, err := entities.NewContractPrice(50)
			assert.NoError(t, err)
			mockService.On("CreateOrder",
				"AAPL-2024",
				entities.SideYes,
				threshold,
			).Return(expectedOrder, nil)

			request := CreateStopLossRequest{
				Ticker:    "AAPL-2024",
				Side:      "YES",
				Threshold: 50,
			}
			body, _ := json.Marshal(request)

			// Act
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/stop-loss", bytes.NewReader(body))
			router.ServeHTTP(rec, req)

			// Assert
			assert.Equal(t, http.StatusOK, rec.Code)

			var response entities.StopLossOrder
			err = json.NewDecoder(rec.Body).Decode(&response)
			assert.NoError(t, err)
			assert.Equal(t, expectedOrder, &response)

			mockService.AssertExpectations(t)
		})

		t.Run("returns error for invalid side", func(t *testing.T) {
			router, _ := setupTest()

			request := CreateStopLossRequest{
				Ticker:    "AAPL-2024",
				Side:      "INVALID",
				Threshold: 50,
			}
			body, _ := json.Marshal(request)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/stop-loss", bytes.NewReader(body))
			router.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusBadRequest, rec.Code)
		})
	})

	t.Run("GetStopLoss", func(t *testing.T) {
		t.Run("successfully gets order", func(t *testing.T) {
			// Arrange
			router, mockService := setupTest()

			orderID := uuid.New()
			expectedOrder := &entities.StopLossOrder{} // Create with appropriate test data
			mockService.On("GetOrder", orderID).Return(expectedOrder, nil)

			// Act
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/stop-loss/%s", orderID), nil)
			router.ServeHTTP(rec, req)

			// Assert
			assert.Equal(t, http.StatusOK, rec.Code)

			var response entities.StopLossOrder
			err := json.NewDecoder(rec.Body).Decode(&response)
			assert.NoError(t, err)
			assert.Equal(t, expectedOrder, &response)

			mockService.AssertExpectations(t)
		})

		t.Run("returns 404 for non-existent order", func(t *testing.T) {
			router, mockService := setupTest()

			orderID := uuid.New()
			mockService.On("GetOrder", orderID).Return(nil, nil)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/stop-loss/%s", orderID), nil)
			router.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusNotFound, rec.Code)
		})
	})

	// Add similar test groups for ListStopLoss, UpdateStopLoss, and CancelStopLoss
}
