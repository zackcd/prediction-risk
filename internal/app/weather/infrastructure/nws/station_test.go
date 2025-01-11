package nws

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestServer(handler http.HandlerFunc) (*httptest.Server, *NWSClient) {
	server := httptest.NewServer(handler)
	client := NewNWSClient(server.URL, "test-user-agent")
	return server, client
}

// TestStationGet tests the Station.Get method
func TestStationGet(t *testing.T) {
	testCases := []struct {
		name       string
		stationID  string
		setupMock  func(w http.ResponseWriter, r *http.Request)
		wantErr    bool
		wantStatus int
	}{
		{
			name:      "successfully gets a station",
			stationID: "KNYC",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/stations/KNYC", r.URL.Path)
				assert.Equal(t, "test-user-agent", r.Header.Get("User-Agent"))
				assert.Equal(t, "application/geo+json", r.Header.Get("Accept"))

				station := &Station{
					ID:   "https://api.weather.gov/stations/KNYC",
					Type: "Feature",
					Geometry: Geometry{
						Type:        "Point",
						Coordinates: []float64{-74.0, 40.7},
					},
					Properties: StationProperties{
						ID:                "https://api.weather.gov/stations/KNYC",
						Type:              "wx:ObservationStation",
						StationIdentifier: "KNYC",
						Name:              "NEW YORK CENTRAL PARK",
						TimeZone:          "America/New_York",
						Elevation: QuantitativeValue{
							Value:    42.7,
							UnitCode: "unit:m",
						},
					},
				}

				w.Header().Set("Content-Type", "application/geo+json")
				json.NewEncoder(w).Encode(station)
			},
		},
		{
			name:      "handles server error",
			stationID: "KNYC",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("internal server error"))
			},
			wantErr:    true,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:      "handles not found error",
			stationID: "INVALID",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("station not found"))
			},
			wantErr:    true,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server, client := setupTestServer(tc.setupMock)
			defer server.Close()

			result, err := client.Station.GetStation(tc.stationID)

			if tc.wantErr {
				require.Error(t, err)
				nwsErr, ok := err.(*NWSError)
				require.True(t, ok)
				assert.Equal(t, tc.wantStatus, nwsErr.StatusCode)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tc.stationID, result.Properties.StationIdentifier)
		})
	}
}

// TestStationListAllStations tests the Station.ListAllStations method
func TestStationListAllStations(t *testing.T) {
	t.Run("successfully lists stations with pagination", func(t *testing.T) {
		var pageCount int
		server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/stations", r.URL.Path)

			response := &StationsResponse{
				Type: "FeatureCollection",
				Features: []Station{{
					ID:   "https://api.weather.gov/stations/KNYC",
					Type: "Feature",
					Properties: StationProperties{
						StationIdentifier: "KNYC",
					},
				}},
				ObservationStations: []string{"https://api.weather.gov/stations/KNYC"},
			}

			// Simulate pagination for first page only
			if pageCount == 0 {
				response.Pagination = &Pagination{
					Next: "cursor123",
				}
			}

			pageCount++
			json.NewEncoder(w).Encode(response)
		})
		defer server.Close()

		result, err := client.Station.ListAllStations()
		require.NoError(t, err)
		assert.Equal(t, "FeatureCollection", result.Type)
		assert.Len(t, result.Features, 2) // Should have two stations (one from each page)
		assert.Equal(t, 2, pageCount)     // Should have made two API calls
	})

	t.Run("handles error during pagination", func(t *testing.T) {
		pageCount := 0
		server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
			pageCount++
			if pageCount > 1 {
				// Second request should error
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("internal server error"))
				return
			}

			// First request succeeds with pagination token
			response := &StationsResponse{
				Type: "FeatureCollection",
				Features: []Station{{
					ID:   "https://api.weather.gov/stations/KNYC",
					Type: "Feature",
					Properties: StationProperties{
						StationIdentifier: "KNYC",
					},
				}},
				ObservationStations: []string{"https://api.weather.gov/stations/KNYC"},
				Pagination: &Pagination{
					Next: "cursor123",
				},
			}
			json.NewEncoder(w).Encode(response)
		})
		defer server.Close()

		result, err := client.Station.ListAllStations()
		require.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, 2, pageCount, "should have attempted two page requests")
	})
}

// TestStationGetObservations tests the Station.GetObservations method
func TestStationGetObservations(t *testing.T) {
	testCases := []struct {
		name      string
		stationID string
		params    ObservationQueryParams
		wantErr   bool
	}{
		{
			name:      "successfully gets observations",
			stationID: "KNYC",
			params:    ObservationQueryParams{},
		},
		{
			name:      "successfully gets observations with time range",
			stationID: "KNYC",
			params: ObservationQueryParams{
				Start: &time.Time{},
				End:   &time.Time{},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/stations/"+tc.stationID+"/observations", r.URL.Path)

				// Verify query parameters if provided
				if tc.params.Start != nil {
					startParam := r.URL.Query().Get("start")
					assert.NotEmpty(t, startParam)
					_, err := time.Parse(time.RFC3339, startParam)
					assert.NoError(t, err, "start time should be in RFC3339 format")
				}
				if tc.params.End != nil {
					endParam := r.URL.Query().Get("end")
					assert.NotEmpty(t, endParam)
					_, err := time.Parse(time.RFC3339, endParam)
					assert.NoError(t, err, "end time should be in RFC3339 format")
				}

				observations := &ObservationCollection{
					Type: "FeatureCollection",
					Features: []Observation{{
						ID:   "https://api.weather.gov/stations/KNYC/observations/2024-01-07T00:00:00Z",
						Type: "Feature",
						Properties: ObservationProperties{
							Timestamp: time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC),
							Temperature: &QuantitativeValue{
								Value:    20.0,
								UnitCode: "unit:degC",
							},
						},
					}},
				}
				json.NewEncoder(w).Encode(observations)
			})
			defer server.Close()

			result, err := client.Station.GetObservations(tc.stationID, tc.params)
			require.NoError(t, err)
			assert.NotNil(t, result)
			assert.NotEmpty(t, result.Features)
		})
	}
}

// TestStationGetLatestObservations tests the Station.GetLatestObservations method
func TestStationGetLatestObservations(t *testing.T) {
	testCases := []struct {
		name      string
		stationID string
		setupMock func(w http.ResponseWriter, r *http.Request)
		wantErr   bool
	}{
		{
			name:      "successfully gets latest observation",
			stationID: "KNYC",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "/stations/KNYC/observations/latest", r.URL.Path)

				observation := &Observation{
					ID:   "https://api.weather.gov/stations/KNYC/observations/latest",
					Type: "Feature",
					Properties: ObservationProperties{
						Timestamp: time.Now().UTC(),
						Temperature: &QuantitativeValue{
							Value:    20.0,
							UnitCode: "unit:degC",
						},
					},
				}
				json.NewEncoder(w).Encode(observation)
			},
		},
		{
			name:      "handles no data available",
			stationID: "KNYC",
			setupMock: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			},
			wantErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			server, client := setupTestServer(tc.setupMock)
			defer server.Close()

			result, err := client.Station.GetLatestObservations(tc.stationID)

			if tc.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.NotZero(t, result.Properties.Timestamp)
		})
	}
}
