package kalshi

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestClient(serverURL string) (*eventClient, error) {
	// Generate a test private key
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, err
	}

	baseClient := NewBaseClient(serverURL, "test-key", privateKey)
	return newEventClient(baseClient), nil
}

func TestEventClient(t *testing.T) {
	t.Run("GetEvent", func(t *testing.T) {
		t.Run("successfully gets single event", func(t *testing.T) {
			// Arrange
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/trade-api/v2/events/SHUTDOWNBY-24", r.URL.Path)
				assert.Equal(t, http.MethodGet, r.Method)

				response := EventResponse{
					Event: Event{
						EventTicker:  "SHUTDOWNBY-24",
						SeriesTicker: "KXSHUTDOWNBY",
						Title:        "Government shuts down in 2024?",
						SubTitle:     "In 2024",
					},
					Markets: []Market{
						{
							Ticker: "SHUTDOWNBY-24",
							Title:  "Government Shutdown",
							Status: "active",
						},
					},
				}
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			client, err := setupTestClient(server.URL)
			require.NoError(t, err)

			// Act
			result, err := client.GetEvent("SHUTDOWNBY-24")

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, "SHUTDOWNBY-24", result.Event.EventTicker)
			assert.Equal(t, "KXSHUTDOWNBY", result.Event.SeriesTicker)
			assert.Equal(t, 1, len(result.Markets))
		})

		t.Run("handles error response", func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				json.NewEncoder(w).Encode(map[string]string{"error": "Event not found"})
			}))
			defer server.Close()

			client, err := setupTestClient(server.URL)
			require.NoError(t, err)
			result, err := client.GetEvent("INVALID-EVENT")

			assert.Error(t, err)
			assert.Nil(t, result)
		})
	})

	t.Run("GetEvents", func(t *testing.T) {
		t.Run("successfully gets paginated events", func(t *testing.T) {
			// Mock server that returns two pages of results
			var callCount int
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/trade-api/v2/events", r.URL.Path)
				assert.Equal(t, http.MethodGet, r.Method)

				callCount++
				if callCount == 1 {
					// First page
					response := EventsResponse{
						Events: []Event{
							{
								EventTicker:  "EVENT1",
								SeriesTicker: "SERIES1",
							},
							{
								EventTicker:  "EVENT2",
								SeriesTicker: "SERIES1",
							},
						},
						Cursor: stringPtr("next-page"),
					}
					json.NewEncoder(w).Encode(response)
				} else {
					// Second/final page
					response := EventsResponse{
						Events: []Event{
							{
								EventTicker:  "EVENT3",
								SeriesTicker: "SERIES2",
							},
						},
						Cursor: nil, // No more pages
					}
					json.NewEncoder(w).Encode(response)
				}
			}))
			defer server.Close()

			client, err := setupTestClient(server.URL)
			require.NoError(t, err)

			// Test with options
			options := NewGetEventsOptions().
				WithLimit(5).
				WithSeriesTicker("SERIES1").
				WithStatuses([]string{"active"})

			result, err := client.GetEvents(options)

			fmt.Printf("RESULT: %d", len(result.Events))

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, 3, len(result.Events))
			assert.Equal(t, "EVENT1", result.Events[0].EventTicker)
			assert.Equal(t, 2, callCount) // Verify pagination worked
		})

		t.Run("respects limit parameter", func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				response := EventsResponse{
					Events: []Event{
						{EventTicker: "EVENT1"},
						{EventTicker: "EVENT2"},
					},
				}
				json.NewEncoder(w).Encode(response)
			}))
			defer server.Close()

			client, err := setupTestClient(server.URL)
			require.NoError(t, err)

			options := NewGetEventsOptions().WithLimit(2)
			result, err := client.GetEvents(options)

			assert.NoError(t, err)
			assert.Equal(t, 2, len(result.Events))
		})

		t.Run("handles error response", func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]string{"error": "Internal server error"})
			}))
			defer server.Close()

			client, err := setupTestClient(server.URL)
			require.NoError(t, err)

			options := NewGetEventsOptions().WithLimit(10)
			result, err := client.GetEvents(options)

			assert.Error(t, err)
			assert.Nil(t, result)
		})
	})
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
