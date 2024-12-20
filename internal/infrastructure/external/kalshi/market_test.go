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

func setupTestMarketClient(serverURL string) (*marketClient, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}
	baseClient := NewBaseClient(serverURL, "test-key", privateKey)
	return NewMarketClient(baseClient), nil
}

func TestMarketClient(t *testing.T) {
	t.Run("GetMarket", func(t *testing.T) {
		t.Run("successfully gets single market", func(t *testing.T) {
			// Arrange
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/trade-api/v2/markets/SHUTDOWNBY-24", r.URL.Path)
				assert.Equal(t, http.MethodGet, r.Method)

				response := MarketResponse{
					Market: Market{
						Ticker:      "SHUTDOWNBY-24",
						EventTicker: "SHUTDOWNBY-24",
						Title:       "Government Shutdown",
						Status:      "active",
						YesBid:      60,
						NoBid:       40,
						MarketType:  "binary",
					},
				}
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			client, err := setupTestMarketClient(server.URL)
			require.NoError(t, err)

			// Act
			result, err := client.GetMarket("SHUTDOWNBY-24")

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, "SHUTDOWNBY-24", result.Market.Ticker)
			assert.Equal(t, "active", result.Market.Status)
			assert.Equal(t, 60, result.Market.YesBid)
		})

		t.Run("handles not found error", func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{"error": "Market not found"})
			}))
			defer server.Close()

			client, err := setupTestMarketClient(server.URL)
			require.NoError(t, err)

			result, err := client.GetMarket("INVALID-MARKET")
			assert.Error(t, err)
			assert.Nil(t, result)
		})
	})

	t.Run("GetMarkets", func(t *testing.T) {
		t.Run("successfully gets paginated markets with filters", func(t *testing.T) {
			var callCount int
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/trade-api/v2/markets", r.URL.Path)
				assert.Equal(t, http.MethodGet, r.Method)
				assert.Equal(t, "active", r.URL.Query().Get("status"))
				assert.Equal(t, "SERIES1", r.URL.Query().Get("series_ticker"))

				callCount++
				if callCount == 1 {
					response := MarketsResponse{
						Markets: []Market{
							{
								Ticker:      "MARKET1",
								EventTicker: "EVENT1",
								Status:      "active",
							},
							{
								Ticker:      "MARKET2",
								EventTicker: "EVENT1",
								Status:      "active",
							},
						},
						Cursor: stringPtr("next-page"),
					}
					json.NewEncoder(w).Encode(response)
				} else {
					response := MarketsResponse{
						Markets: []Market{
							{
								Ticker:      "MARKET3",
								EventTicker: "EVENT2",
								Status:      "active",
							},
						},
					}
					json.NewEncoder(w).Encode(response)
				}
			}))
			defer server.Close()

			client, err := setupTestMarketClient(server.URL)
			require.NoError(t, err)

			now := time.Now()
			options := NewGetMarketsOptions().
				WithStatus([]string{"active"}).
				WithSeriesTicker("SERIES1").
				WithMaxCloseTime(now.Add(24 * time.Hour))

			result, err := client.GetMarkets(options)

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, 3, len(result.Markets))
			assert.Equal(t, "MARKET1", result.Markets[0].Ticker)
			assert.Equal(t, 2, callCount)
		})

		t.Run("handles market query parameters", func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "MARKET1,MARKET2", r.URL.Query().Get("ticker"))
				assert.Equal(t, "EVENT1", r.URL.Query().Get("event_ticker"))
				assert.Equal(t, "active,settled", r.URL.Query().Get("status"))

				response := MarketsResponse{
					Markets: []Market{
						{Ticker: "MARKET1"},
						{Ticker: "MARKET2"},
					},
				}
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			client, err := setupTestMarketClient(server.URL)
			require.NoError(t, err)

			options := NewGetMarketsOptions().
				WithTickers([]string{"MARKET1", "MARKET2"}).
				WithEventTicker("EVENT1").
				WithStatus([]string{"active", "settled"})

			result, err := client.GetMarkets(options)
			assert.NoError(t, err)
			assert.Equal(t, 2, len(result.Markets))
		})

		t.Run("respects pagination limit", func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "2", r.URL.Query().Get("limit"))

				response := MarketsResponse{
					Markets: []Market{
						{Ticker: "MARKET1"},
						{Ticker: "MARKET2"},
					},
				}
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			client, err := setupTestMarketClient(server.URL)
			require.NoError(t, err)

			limit := 2
			options := NewGetMarketsOptions()
			options.Limit = &limit

			result, err := client.GetMarkets(options)
			assert.NoError(t, err)
			assert.Equal(t, 2, len(result.Markets))
		})
	})
}
