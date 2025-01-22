package weather_service

import (
	weather_domain "prediction-risk/internal/app/weather/domain"
	weather_mocks "prediction-risk/internal/app/weather/mocks"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestObservation(stationID string) *weather_domain.TemperatureObservation {
	return weather_domain.NewTemperatureObservation(
		stationID,
		weather_domain.Temperature{
			Value:           23.5,
			TemperatureUnit: weather_domain.Celsius,
		},
		time.Now(),
	)
}

func TestWeatherMonitor_Start(t *testing.T) {
	t.Run("starts monitoring and processes observations", func(t *testing.T) {
		mockService := &weather_mocks.MockWeatherObservationService{}
		stationID := "KNYC"
		interval := 100 * time.Millisecond

		observation := createTestObservation(stationID)

		// Expect at least two calls to RetrieveLatestObservation
		mockService.On("RetrieveLatestObservation", stationID).
			Return(observation, nil).
			Times(2)

		monitor := NewWeatherMonitor(
			stationID,
			mockService,
			interval,
		)

		// Start monitoring
		monitor.Start()

		// Wait for at least two intervals
		time.Sleep(250 * time.Millisecond)

		// Stop monitoring
		monitor.Stop()

		// Give it a moment to clean up
		time.Sleep(50 * time.Millisecond)

		mockService.AssertExpectations(t)
	})

	t.Run("handles service errors gracefully", func(t *testing.T) {
		mockService := &weather_mocks.MockWeatherObservationService{}
		stationID := "KNYC"
		interval := 100 * time.Millisecond

		// Mock service to return error
		mockService.On("RetrieveLatestObservation", stationID).
			Return(nil, assert.AnError).
			Times(2)

		monitor := NewWeatherMonitor(
			stationID,
			mockService,
			interval,
		)

		monitor.Start()
		time.Sleep(250 * time.Millisecond)
		monitor.Stop()
		time.Sleep(50 * time.Millisecond)

		mockService.AssertExpectations(t)
	})
}

func TestWeatherMonitor_Stop(t *testing.T) {
	t.Run("stops monitoring when requested", func(t *testing.T) {
		mockService := &weather_mocks.MockWeatherObservationService{}
		stationID := "KNYC"
		interval := 100 * time.Millisecond

		observation := createTestObservation(stationID)
		mockService.On("RetrieveLatestObservation", stationID).
			Return(observation, nil).
			Maybe() // Allow any number of calls before stopping

		monitor := NewWeatherMonitor(
			stationID,
			mockService,
			interval,
		)

		// Start and immediately stop
		monitor.Start()
		monitor.Stop()

		// Wait to ensure no more calls are made
		time.Sleep(250 * time.Millisecond)

		// Start again to verify monitor can be restarted
		monitor = NewWeatherMonitor(
			stationID,
			mockService,
			interval,
		)
		monitor.Start()
		time.Sleep(150 * time.Millisecond)
		monitor.Stop()
	})
}

func TestWeatherMonitor_CheckWeatherObservation(t *testing.T) {
	t.Run("processes observation successfully", func(t *testing.T) {
		mockService := &weather_mocks.MockWeatherObservationService{}
		stationID := "KNYC"
		interval := time.Second

		observation := createTestObservation(stationID)
		mockService.On("RetrieveLatestObservation", stationID).
			Return(observation, nil).
			Once()

		monitor := NewWeatherMonitor(
			stationID,
			mockService,
			interval,
		)

		err := monitor.checkWeatherObservation()
		require.NoError(t, err)
		mockService.AssertExpectations(t)
	})

	t.Run("handles retrieval error", func(t *testing.T) {
		mockService := &weather_mocks.MockWeatherObservationService{}
		stationID := "KNYC"
		interval := time.Second

		mockService.On("RetrieveLatestObservation", stationID).
			Return(nil, assert.AnError).
			Once()

		monitor := NewWeatherMonitor(
			stationID,
			mockService,
			interval,
		)

		err := monitor.checkWeatherObservation()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to retrieve latest weather observation")
		mockService.AssertExpectations(t)
	})
}

func TestWeatherMonitor_ProcessWeatherObservation(t *testing.T) {
	t.Run("processes observation", func(t *testing.T) {
		mockService := &weather_mocks.MockWeatherObservationService{}
		monitor := NewWeatherMonitor(
			"KNYC",
			mockService,
			time.Second,
		)

		observation := createTestObservation("KNYC")
		processed, err := monitor.processWeatherObservation(observation)

		require.NoError(t, err)
		assert.Equal(t, observation, processed)
	})
}
