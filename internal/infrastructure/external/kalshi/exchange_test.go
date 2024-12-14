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

// TestExchangeService groups all exchange service tests
func TestExchangeService(t *testing.T) {
	// Setup helper to create test server and client
	setup := func(handler http.HandlerFunc) (*exchangeClient, *httptest.Server) {
		server := httptest.NewServer(handler)

		privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
		require.NoError(t, err, "Failed to generate test private key")

		client := NewBaseClient(server.URL, "test-key", privateKey)
		return NewExchangeClient(client), server
	}

	t.Run("GetSchedule", func(t *testing.T) {
		t.Run("successfully returns schedule", func(t *testing.T) {
			// Arrange
			expectedResp := ExchangeScheduleResponse{
				Schedule: Schedule{
					MaintenanceWindows: []MaintenanceWindow{
						{
							StartDatetime: time.Date(2024, 12, 12, 18, 56, 12, 719000000, time.UTC),
							EndDatetime:   time.Date(2024, 12, 12, 19, 56, 12, 719000000, time.UTC),
						},
					},
					StandardHours: []StandardHours{
						{
							StartTime: time.Date(2024, 12, 12, 18, 56, 12, 719000000, time.UTC),
							EndTime:   time.Date(2024, 12, 12, 19, 56, 12, 719000000, time.UTC),
							Monday: []TradingHours{
								{OpenTime: "09:00", CloseTime: "17:00"},
							},
						},
					},
				},
			}

			service, server := setup(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/trade-api/v2/exchange/schedule", r.URL.Path)
				assert.Equal(t, "GET", r.Method)

				json.NewEncoder(w).Encode(expectedResp)
			})
			defer server.Close()

			// Act
			schedule, err := service.GetExchangeSchedule()

			// Assert
			require.NoError(t, err)
			assert.Equal(t, expectedResp, *schedule)
		})

		t.Run("handles error response", func(t *testing.T) {
			service, server := setup(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(`{"error": "internal error"}`))
			})
			defer server.Close()

			schedule, err := service.GetExchangeSchedule()

			assert.Error(t, err)
			assert.Nil(t, schedule)
			assert.Contains(t, err.Error(), "500")
		})
	})

	t.Run("GetAnnouncements", func(t *testing.T) {
		t.Run("successfully returns announcements", func(t *testing.T) {
			// Arrange
			expectedResp := ExchangeAnnouncementsResponse{
				Announcements: []Announcement{
					{
						DeliveryTime: time.Date(2024, 12, 12, 18, 56, 12, 719000000, time.UTC),
						Message:      "Test announcement",
						Status:       "info",
						Type:         "info",
					},
				},
			}

			service, server := setup(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/trade-api/v2/exchange/announcements", r.URL.Path)
				assert.Equal(t, "GET", r.Method)

				json.NewEncoder(w).Encode(expectedResp)
			})
			defer server.Close()

			// Act
			announcements, err := service.GetExchangeAnnouncements()

			// Assert
			require.NoError(t, err)
			assert.Equal(t, expectedResp, *announcements)
		})

		t.Run("handles empty announcements", func(t *testing.T) {
			service, server := setup(func(w http.ResponseWriter, r *http.Request) {
				json.NewEncoder(w).Encode(ExchangeAnnouncementsResponse{
					Announcements: []Announcement{},
				})
			})
			defer server.Close()

			announcements, err := service.GetExchangeAnnouncements()

			require.NoError(t, err)
			assert.Empty(t, announcements.Announcements)
		})
	})
}
