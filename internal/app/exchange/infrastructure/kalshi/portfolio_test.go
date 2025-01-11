package kalshi

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestPortfolioClient(serverURL string) (*portfolioClient, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	baseClient := newClient(serverURL, "test-key", privateKey)
	return NewPortfolioClient(baseClient), nil
}

func TestPortfolioClient(t *testing.T) {
	t.Run("CreateOrder", func(t *testing.T) {
		t.Run("successfully creates limit order", func(t *testing.T) {
			// Arrange
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/trade-api/v2/portfolio/order", r.URL.Path)
				assert.Equal(t, http.MethodPost, r.Method)

				// Verify request body
				var reqBody CreateOrderRequest
				err := json.NewDecoder(r.Body).Decode(&reqBody)
				require.NoError(t, err)
				assert.Equal(t, "SHUTDOWNBY-24", reqBody.Ticker)
				assert.Equal(t, OrderSideYes, reqBody.Side)
				assert.Equal(t, OrderActionBuy, reqBody.Action)
				assert.Equal(t, OrderTypeLimit, reqBody.Type)
				assert.Equal(t, 10, reqBody.Count)
				assert.Equal(t, 60, *reqBody.YesPrice)

				// Return mock response
				response := CreateOrderResponse{
					Order: Order{
						ID:            "test-order-id",
						Ticker:        reqBody.Ticker,
						Side:          reqBody.Side,
						Action:        string(reqBody.Action),
						Type:          reqBody.Type,
						YesPrice:      *reqBody.YesPrice,
						Status:        OrderStatusOpen,
						CreatedTime:   time.Now(),
						ClientOrderID: reqBody.ClientOrderID,
					},
				}
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			client, err := setupTestPortfolioClient(server.URL)
			require.NoError(t, err)

			// Act
			yesPrice := 60
			request := CreateOrderRequest{
				Ticker:        "SHUTDOWNBY-24",
				ClientOrderID: "test-client-id",
				Side:          OrderSideYes,
				Action:        OrderActionBuy,
				Count:         10,
				Type:          OrderTypeLimit,
				YesPrice:      &yesPrice,
			}
			result, err := client.CreateOrder(request)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, request.Ticker, result.Order.Ticker)
			assert.Equal(t, request.Side, result.Order.Side)
			assert.Equal(t, OrderStatusOpen, result.Order.Status)
		})

		t.Run("handles error response", func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "Invalid order parameters"})
			}))
			defer server.Close()

			client, err := setupTestPortfolioClient(server.URL)
			require.NoError(t, err)

			request := CreateOrderRequest{
				Ticker: "INVALID",
				Side:   OrderSideYes,
				Action: OrderActionBuy,
				Count:  -1, // Invalid count
			}

			result, err := client.CreateOrder(request)
			assert.Error(t, err)
			assert.Nil(t, result)
		})
	})

	t.Run("GetPositions", func(t *testing.T) {
		t.Run("successfully gets positions with pagination", func(t *testing.T) {
			var callCount int
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/trade-api/v2/portfolio/positions", r.URL.Path)
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "open", r.URL.Query().Get("settlement_status"))

				callCount++
				if callCount == 1 {
					response := PositionsResponse{
						MarketPositions: []MarketPosition{
							{
								Ticker:         "MARKET1",
								Position:       10,
								MarketExposure: 1000,
							},
						},
						EventPositions: []EventPosition{
							{
								EventTicker:   "EVENT1",
								EventExposure: 1000,
							},
						},
						Cursor: stringPtr("next-page"),
					}
					json.NewEncoder(w).Encode(response)
				} else {
					response := PositionsResponse{
						MarketPositions: []MarketPosition{
							{
								Ticker:         "MARKET2",
								Position:       -5,
								MarketExposure: -500,
							},
						},
						EventPositions: []EventPosition{
							{
								EventTicker:   "EVENT2",
								EventExposure: -500,
							},
						},
					}
					json.NewEncoder(w).Encode(response)
				}
			}))
			defer server.Close()

			client, err := setupTestPortfolioClient(server.URL)
			require.NoError(t, err)

			options := NewGetPositionsOptions().
				WithSettlementStatus(SettlementStatusOpen)

			result, err := client.GetPositions(options)

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, 2, len(result.MarketPositions))
			assert.Equal(t, 2, len(result.EventPositions))
			assert.Equal(t, "MARKET1", result.MarketPositions[0].Ticker)
			assert.Equal(t, "EVENT1", result.EventPositions[0].EventTicker)
			assert.Equal(t, 2, callCount) // Verify pagination worked
		})
	})
}
