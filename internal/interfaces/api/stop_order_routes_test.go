package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"prediction-risk/internal/domain/contract"
	"prediction-risk/internal/domain/core"
	"prediction-risk/internal/domain/order"
	"prediction-risk/internal/domain/order/mocks"
	"testing"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
)

func setupTest() (*chi.Mux, *mocks.MockStopOrderService) {
	mockService := new(mocks.MockStopOrderService)
	routes := NewStopOrderRoutes(mockService)

	router := chi.NewRouter()
	routes.Register(router)

	return router, mockService
}

func TestStopOrderRoutes(t *testing.T) {
	t.Run("CreateStopOrder", func(t *testing.T) {
		t.Run("successfully creates order", func(t *testing.T) {
			// Arrange
			router, mockService := setupTest()

			threshold, err := contract.NewContractPrice(50)
			assert.NoError(t, err)

			// Create a properly initialized test order
			expectedOrder := order.NewStopOrder(
				"AAPL-2024",
				contract.SideYes,
				threshold,
				nil,
				nil,
			)

			mockService.On("CreateOrder",
				"AAPL-2024",
				contract.SideYes,
				threshold,
				(*contract.ContractPrice)(nil),
			).Return(expectedOrder, nil)

			request := CreateStopOrderRequest{
				Ticker:       "AAPL-2024",
				Side:         "YES",
				TriggerPrice: 50,
				LimitPrice:   nil,
			}
			body, _ := json.Marshal(request)

			// Act
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/stop-orders", bytes.NewReader(body))
			router.ServeHTTP(rec, req)

			// Assert
			assert.Equal(t, http.StatusOK, rec.Code)

			var response StopOrderResponse // Change to your API response type
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

			request := CreateStopOrderRequest{
				Ticker:       "AAPL-2024",
				Side:         "INVALID",
				TriggerPrice: 50,
			}
			body, _ := json.Marshal(request)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/stop-orders", bytes.NewReader(body))
			router.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusBadRequest, rec.Code)
		})
	})

	t.Run("GetStopOrder", func(t *testing.T) {
		t.Run("successfully gets order", func(t *testing.T) {
			// Arrange
			router, mockService := setupTest()

			threshold, err := contract.NewContractPrice(50)
			assert.NoError(t, err)
			// Create a properly initialized test order
			expectedOrder := order.NewStopOrder(
				"AAPL-2024",
				contract.SideYes,
				threshold,
				nil,
				nil,
			)
			mockService.On("GetOrder", expectedOrder.ID()).Return(expectedOrder, nil)

			// Act
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/stop-orders/%s", expectedOrder.ID().String()), nil)
			router.ServeHTTP(rec, req)

			// Assert
			assert.Equal(t, http.StatusOK, rec.Code)

			var response StopOrderResponse
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

			orderID := order.NewOrderID()
			mockService.On("GetOrder", orderID).Return(nil, nil)

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/stop-orders/%s", orderID), nil)
			router.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusNotFound, rec.Code)
		})
	})

	t.Run("ListStopOrder", func(t *testing.T) {
		t.Run("successfully lists active orders", func(t *testing.T) {
			// Arrange
			router, mockService := setupTest()

			threshold, err := contract.NewContractPrice(50)
			assert.NoError(t, err)

			expectedOrder1 := order.NewStopOrder("AAPL-2024", contract.SideYes, threshold, nil, nil)
			expectedOrder2 := order.NewStopOrder("MSFT-2024", contract.SideNo, threshold, nil, nil)

			expectedOrders := []*order.StopOrder{expectedOrder1, expectedOrder2}
			mockService.On("GetActiveOrders").Return(expectedOrders, nil)

			// Act
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/stop-orders", nil)
			router.ServeHTTP(rec, req)

			// Assert
			assert.Equal(t, http.StatusOK, rec.Code)

			var response []StopOrderResponse
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
			req := httptest.NewRequest(http.MethodGet, "/api/stop-orders", nil)
			router.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusInternalServerError, rec.Code)
		})
	})

	t.Run("UpdateStopOrder", func(t *testing.T) {
		t.Run("successfully updates order", func(t *testing.T) {
			// Arrange
			router, mockService := setupTest()

			initialThreshold, err := contract.NewContractPrice(50)
			assert.NoError(t, err)
			newThreshold, err := contract.NewContractPrice(60)
			assert.NoError(t, err)

			existingOrder := order.NewStopOrder("AAPL-2024", contract.SideYes, initialThreshold, nil, nil)
			updatedOrder := existingOrder
			updatedOrder.SetTriggerPrice(newThreshold)

			mockService.On("UpdateOrder", existingOrder.ID(), &newThreshold, (*contract.ContractPrice)(nil)).Return(updatedOrder, nil)

			price := 60
			request := UpdateStopOrderRequest{
				TriggerPrice: &price,
			}
			body, _ := json.Marshal(request)

			// Act
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(
				http.MethodPatch,
				fmt.Sprintf("/api/stop-orders/%s", existingOrder.ID().String()),
				bytes.NewReader(body),
			)
			router.ServeHTTP(rec, req)

			// Assert
			assert.Equal(t, http.StatusOK, rec.Code)

			var response StopOrderResponse
			err = json.NewDecoder(rec.Body).Decode(&response)
			assert.NoError(t, err)
			assert.Equal(t, updatedOrder.ID().String(), response.ID)
			assert.Equal(t, int(updatedOrder.TriggerPrice().Value()), response.TriggerPrice)

			mockService.AssertExpectations(t)
		})

		t.Run("handles invalid threshold", func(t *testing.T) {
			// Arrange
			router, mockService := setupTest()
			orderId := order.NewOrderID()

			price := 101
			request := UpdateStopOrderRequest{
				TriggerPrice: &price,
			}
			body, _ := json.Marshal(request)

			// Act
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(
				http.MethodPatch,
				fmt.Sprintf("/api/stop-orders/%s", orderId.String()),
				bytes.NewReader(body),
			)
			req.Header.Set("Content-Type", "application/json")
			router.ServeHTTP(rec, req)

			// Assert
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			mockService.AssertExpectations(t)
		})
	})

	t.Run("CancelStopOrder", func(t *testing.T) {
		t.Run("successfully cancels order", func(t *testing.T) {
			// Arrange
			router, mockService := setupTest()

			threshold, err := contract.NewContractPrice(50)
			assert.NoError(t, err)

			existingOrder := order.NewStopOrder("AAPL-2024", contract.SideYes, threshold, nil, nil)
			cancelledOrder := order.NewStopOrder("AAPL-2024", contract.SideYes, threshold, nil, nil)
			cancelledOrder.UpdateStatus(order.OrderStatusCancelled)

			mockService.On("CancelOrder", existingOrder.ID()).Return(cancelledOrder, nil)

			// Act
			rec := httptest.NewRecorder()
			req := httptest.NewRequest(
				http.MethodDelete,
				fmt.Sprintf("/api/stop-orders/%s", existingOrder.ID().String()),
				nil,
			)
			router.ServeHTTP(rec, req)

			// Assert
			assert.Equal(t, http.StatusOK, rec.Code)

			var response StopOrderResponse
			err = json.NewDecoder(rec.Body).Decode(&response)
			assert.NoError(t, err)
			assert.Equal(t, cancelledOrder.ID().String(), response.ID)
			assert.Equal(t, string(cancelledOrder.Status()), response.Status)

			mockService.AssertExpectations(t)
		})

		t.Run("handles non-existent order", func(t *testing.T) {
			router, mockService := setupTest()

			nonExistentID := order.NewOrderID()
			mockService.On("CancelOrder", nonExistentID).Return(nil, core.NewErrNotFound("StopOrder", nonExistentID.String()))

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(
				http.MethodDelete,
				fmt.Sprintf("/api/stop-orders/%s", nonExistentID.String()),
				nil,
			)
			router.ServeHTTP(rec, req)

			assert.Equal(t, http.StatusNotFound, rec.Code)
		})
	})
}
