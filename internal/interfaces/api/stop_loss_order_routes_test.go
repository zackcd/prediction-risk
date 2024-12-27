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
	isDryRun bool,
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

			threshold, err := entities.NewContractPrice(50)
			assert.NoError(t, err)

			// Create a properly initialized test order
			expectedOrder := entities.NewStopLossOrder(
				"AAPL-2024",
				entities.SideYes,
				threshold,
			)

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

			var response StopLossOrderResponse // Change to your API response type
			err = json.NewDecoder(rec.Body).Decode(&response)
			assert.NoError(t, err)
			assert.Equal(t, expectedOrder.ID().String(), response.ID)
			assert.Equal(t, "AAPL-2024", response.Ticker)
			assert.Equal(t, "YES", response.Side)
			assert.Equal(t, 50, response.TriggerPrice)
			assert.Equal(t, "ACTIVE", response.Status)

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

			threshold, err := entities.NewContractPrice(50)
			assert.NoError(t, err)
			// Create a properly initialized test order
			expectedOrder := entities.NewStopLossOrder(
				"AAPL-2024",
				entities.SideYes,
				threshold,
			)
			mockService.On("GetOrder", expectedOrder.ID()).Return(expectedOrder, nil)

			// Act
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/stop-loss/%s", expectedOrder.ID().String()), nil)
			router.ServeHTTP(rec, req)

			// Assert
			assert.Equal(t, http.StatusOK, rec.Code)

			var response StopLossOrderResponse
			err = json.NewDecoder(rec.Body).Decode(&response)
			assert.NoError(t, err)

			// Compare individual fields
			assert.Equal(t, expectedOrder.ID().String(), response.ID)
			assert.Equal(t, expectedOrder.Ticker(), response.Ticker)
			assert.Equal(t, expectedOrder.Side().String(), response.Side)
			assert.Equal(t, int(expectedOrder.TriggerPrice().Value()), response.TriggerPrice)
			assert.Equal(t, string(expectedOrder.Status()), response.Status)

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

	t.Run("ListStopLoss", func(t *testing.T) {
		t.Run("successfully lists active orders", func(t *testing.T) {
			// Arrange
			router, mockService := setupTest()

			threshold, err := entities.NewContractPrice(50)
			assert.NoError(t, err)

			expectedOrder1 := entities.NewStopLossOrder("AAPL-2024", entities.SideYes, threshold)
			expectedOrder2 := entities.NewStopLossOrder("MSFT-2024", entities.SideNo, threshold)

			expectedOrders := []*entities.StopLossOrder{expectedOrder1, expectedOrder2}
			mockService.On("GetActiveOrders").Return(expectedOrders, nil)

			// Act
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/stop-loss", nil)
			router.ServeHTTP(rec, req)

			// Assert
			assert.Equal(t, http.StatusOK, rec.Code)

			var response []StopLossOrderResponse
			err = json.NewDecoder(rec.Body).Decode(&response)
			assert.NoError(t, err)
			assert.Len(t, response, 2)

			// Verify first order
			assert.Equal(t, expectedOrder1.ID().String(), response[0].ID)
			assert.Equal(t, expectedOrder1.Ticker(), response[0].Ticker)
			assert.Equal(t, expectedOrder1.Side().String(), response[0].Side)
			assert.Equal(t, int(expectedOrder1.TriggerPrice().Value()), response[0].TriggerPrice)

			// Verify second order
			assert.Equal(t, expectedOrder2.ID().String(), response[1].ID)
			assert.Equal(t, expectedOrder2.Ticker(), response[1].Ticker)
			assert.Equal(t, expectedOrder2.Side().String(), response[1].Side)
			assert.Equal(t, int(expectedOrder2.TriggerPrice().Value()), response[1].TriggerPrice)

			mockService.AssertExpectations(t)
		})

		t.Run("handles error from service", func(t *testing.T) {
			router, mockService := setupTest()
			mockService.On("GetActiveOrders").Return(nil, fmt.Errorf("database error"))

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/stop-loss", nil)
			router.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusInternalServerError, rec.Code)
		})
	})

	t.Run("UpdateStopLoss", func(t *testing.T) {
		t.Run("successfully updates order", func(t *testing.T) {
			// Arrange
			router, mockService := setupTest()

			initialThreshold, err := entities.NewContractPrice(50)
			assert.NoError(t, err)
			newThreshold, err := entities.NewContractPrice(60)
			assert.NoError(t, err)

			existingOrder := entities.NewStopLossOrder("AAPL-2024", entities.SideYes, initialThreshold)
			updatedOrder := existingOrder
			updatedOrder.UpdateTriggerPrice(newThreshold)

			mockService.On("UpdateOrder", existingOrder.ID(), newThreshold).Return(updatedOrder, nil)

			request := UpdateStopLossRequest{
				Threshold: 60,
			}
			body, _ := json.Marshal(request)

			// Act
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(
				http.MethodPatch,
				fmt.Sprintf("/api/stop-loss/%s", existingOrder.ID().String()),
				bytes.NewReader(body),
			)
			router.ServeHTTP(rec, req)

			// Assert
			assert.Equal(t, http.StatusOK, rec.Code)

			var response StopLossOrderResponse
			err = json.NewDecoder(rec.Body).Decode(&response)
			assert.NoError(t, err)
			assert.Equal(t, updatedOrder.ID().String(), response.ID)
			assert.Equal(t, int(updatedOrder.TriggerPrice().Value()), response.TriggerPrice)

			mockService.AssertExpectations(t)
		})

		t.Run("handles invalid threshold", func(t *testing.T) {
			// Arrange
			router, mockService := setupTest()
			orderId := uuid.New()

			// Mock the service call - it should return an error for invalid threshold
			invalidThreshold, _ := entities.NewContractPrice(101) // This might fail, depending on your validation
			mockService.On("UpdateOrder", orderId, invalidThreshold).
				Return(nil, fmt.Errorf("threshold must be between 0 and 100"))

			request := UpdateStopLossRequest{
				Threshold: 101,
			}
			body, _ := json.Marshal(request)

			// Act
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(
				http.MethodPatch,
				fmt.Sprintf("/api/stop-loss/%s", orderId.String()),
				bytes.NewReader(body),
			)
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(rec, req)

			// Assert
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			mockService.AssertExpectations(t)
		})
	})

	t.Run("CancelStopLoss", func(t *testing.T) {
		t.Run("successfully cancels order", func(t *testing.T) {
			// Arrange
			router, mockService := setupTest()

			threshold, err := entities.NewContractPrice(50)
			assert.NoError(t, err)

			existingOrder := entities.NewStopLossOrder("AAPL-2024", entities.SideYes, threshold)
			cancelledOrder := entities.NewStopLossOrder("AAPL-2024", entities.SideYes, threshold)
			cancelledOrder.UpdateStatus(entities.OrderStatusCancelled)

			mockService.On("CancelOrder", existingOrder.ID()).Return(cancelledOrder, nil)

			// Act
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(
				http.MethodDelete,
				fmt.Sprintf("/api/stop-loss/%s", existingOrder.ID().String()),
				nil,
			)
			router.ServeHTTP(rec, req)

			// Assert
			assert.Equal(t, http.StatusOK, rec.Code)

			var response StopLossOrderResponse
			err = json.NewDecoder(rec.Body).Decode(&response)
			assert.NoError(t, err)
			assert.Equal(t, cancelledOrder.ID().String(), response.ID)
			assert.Equal(t, string(cancelledOrder.Status()), response.Status)

			mockService.AssertExpectations(t)
		})

		t.Run("handles non-existent order", func(t *testing.T) {
			router, mockService := setupTest()

			nonExistentID := uuid.New()
			mockService.On("CancelOrder", nonExistentID).Return(nil, entities.NewErrNotFound("StopLossOrder", nonExistentID.String()))

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(
				http.MethodDelete,
				fmt.Sprintf("/api/stop-loss/%s", nonExistentID.String()),
				nil,
			)
			router.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusNotFound, rec.Code)
		})
	})
}
