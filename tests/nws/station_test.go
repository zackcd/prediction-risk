package nws_test // Use _test package to ensure we're testing the public API

import (
	"prediction-risk/internal/infrastructure/external/nws"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

const (
	testUserAgent = "go-nws-test/1.0 (github.com/yourusername/nws-go; your@email.com)"
	baseURL       = "https://api.weather.gov"
)

func setupClient(t *testing.T) *nws.NWSClient {
	client := nws.NewNWSClient(baseURL, testUserAgent)
	require.NotNil(t, client)
	return client
}

func TestStationIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := setupClient(t)

	// Use KNYC (Central Park, NYC) as a reliable test station
	const testStation = "KNYC"

	t.Run("get station details", func(t *testing.T) {
		station, err := client.Station.Get(testStation)
		require.NoError(t, err)
		require.NotNil(t, station)

		// Verify station details
		require.Equal(t, "Feature", station.Type)
		require.Equal(t, testStation, station.Properties.StationIdentifier)
		require.Equal(t, "wx:ObservationStation", station.Properties.Type)

		// Verify coordinates exist
		require.Len(t, station.Geometry.Coordinates, 2)
		require.NotZero(t, station.Geometry.Coordinates[0]) // longitude
		require.NotZero(t, station.Geometry.Coordinates[1]) // latitude
	})

	t.Run("list stations", func(t *testing.T) {
		stations, err := client.Station.ListAllStations()
		require.NoError(t, err)
		require.NotNil(t, stations)

		require.Equal(t, "FeatureCollection", stations.Type)
		require.NotEmpty(t, stations.Features)
		require.NotEmpty(t, stations.ObservationStations)
	})

	t.Run("get station observations", func(t *testing.T) {
		observations, err := client.Station.GetObservations(testStation, nws.ObservationQueryParams{})
		require.NoError(t, err)
		require.NotNil(t, observations)
		require.NotEmpty(t, observations.Features)

		// Test observation fields
		obs := observations.Features[0]
		require.NotEmpty(t, obs.Properties.Timestamp)
		require.NotNil(t, obs.Properties.Temperature)
		require.NotEmpty(t, obs.Properties.TextDescription)
	})

	t.Run("get latest observation", func(t *testing.T) {
		observation, err := client.Station.GetLatestObservations(testStation)
		require.NoError(t, err)
		require.NotNil(t, observation)

		// Verify it's recent
		require.True(t, time.Since(observation.Properties.Timestamp) < 2*time.Hour)

		// Check common fields
		require.NotNil(t, observation.Properties.Temperature)
		require.NotNil(t, observation.Properties.Dewpoint)
		require.NotNil(t, observation.Properties.WindSpeed)
		require.NotEmpty(t, observation.Properties.TextDescription)
	})

	t.Run("handle invalid station ID", func(t *testing.T) {
		_, err := client.Station.Get("INVALID")
		require.Error(t, err)
	})

	// Add a small delay between tests to respect rate limits
	time.Sleep(500 * time.Millisecond)
}

func TestStationObservationValues(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	client := setupClient(t)

	observation, err := client.Station.GetLatestObservations("KNYC")
	require.NoError(t, err)

	t.Run("verify quantitative values", func(t *testing.T) {
		// Temperature should have a reasonable value
		temp := observation.Properties.Temperature
		require.NotNil(t, temp)
		require.True(t, temp.Value > -50 && temp.Value < 50, "Temperature should be reasonable")
		require.Contains(t, temp.UnitCode, "deg")

		// Wind speed should be non-negative
		if wind := observation.Properties.WindSpeed; wind != nil {
			require.GreaterOrEqual(t, wind.Value, 0.0)
		}

		// Relative humidity should be between 0-100 if present
		if rh := observation.Properties.RelativeHumidity; rh != nil {
			require.True(t, rh.Value >= 0 && rh.Value <= 100)
		}
	})

	t.Run("verify text fields", func(t *testing.T) {
		require.NotEmpty(t, observation.Properties.TextDescription)
		require.NotEmpty(t, observation.Properties.RawMessage)
	})

	// Add a small delay after tests
	time.Sleep(500 * time.Millisecond)
}
