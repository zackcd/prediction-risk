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

func TestStationGet(t *testing.T) {
	t.Run("successfully gets a station", func(t *testing.T) {
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

		server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/stations/KNYC", r.URL.Path)
			assert.Equal(t, "test-user-agent", r.Header.Get("User-Agent"))
			assert.Equal(t, "application/geo+json", r.Header.Get("Accept"))

			w.Header().Set("Content-Type", "application/geo+json")
			json.NewEncoder(w).Encode(station)
		})
		defer server.Close()

		result, err := client.Station.Get("KNYC")
		require.NoError(t, err)
		assert.Equal(t, station, result)
	})

	t.Run("handles server error", func(t *testing.T) {
		server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("internal server error"))
		})
		defer server.Close()

		result, err := client.Station.Get("KNYC")
		require.Error(t, err)
		assert.Nil(t, result)

		nwsErr, ok := err.(*NWSError)
		require.True(t, ok)
		assert.Equal(t, http.StatusInternalServerError, nwsErr.StatusCode)
	})
}

func TestStationList(t *testing.T) {
	t.Run("successfully lists stations", func(t *testing.T) {
		collection := &StationCollection{
			Type: "FeatureCollection",
			Features: []Station{{
				ID:   "https://api.weather.gov/stations/KNYC",
				Type: "Feature",
				Properties: StationProperties{
					StationIdentifier: "KNYC",
				},
			}},
			ObservationStations: []string{"https://api.weather.gov/stations/KNYC"},
			Pagination: Pagination{
				Next: "https://api.weather.gov/stations?cursor=abc123",
			},
		}

		server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/stations", r.URL.Path)
			json.NewEncoder(w).Encode(collection)
		})
		defer server.Close()

		result, err := client.Station.List(nil)
		require.NoError(t, err)
		assert.Equal(t, collection, result)
	})
}

func TestStationGetObservations(t *testing.T) {
	t.Run("successfully gets observations", func(t *testing.T) {
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

		server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/stations/KNYC/observations", r.URL.Path)
			json.NewEncoder(w).Encode(observations)
		})
		defer server.Close()

		result, err := client.Station.GetObservations("KNYC")
		require.NoError(t, err)
		assert.Equal(t, observations, result)
	})
}

func TestStationGetLatestObservations(t *testing.T) {
	t.Run("successfully gets latest observation", func(t *testing.T) {
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

		server, client := setupTestServer(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "/stations/KNYC/observations/latest", r.URL.Path)
			json.NewEncoder(w).Encode(observation)
		})
		defer server.Close()

		result, err := client.Station.GetLatestObservations("KNYC")
		require.NoError(t, err)
		assert.Equal(t, observation, result)
	})
}
