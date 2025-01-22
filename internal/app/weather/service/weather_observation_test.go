package weather_service

import (
	"errors"
	weather_domain "prediction-risk/internal/app/weather/domain"
	"prediction-risk/internal/app/weather/infrastructure/nws"
	weather_mocks "prediction-risk/internal/app/weather/mocks"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newTestService() (
	*weatherObservationService,
	*weather_mocks.MockObservationGetter,
	*weather_mocks.MockTemperatureObservationRepo,
) {
	observations := &weather_mocks.MockObservationGetter{}
	repo := &weather_mocks.MockTemperatureObservationRepo{}

	service := &weatherObservationService{
		temperatureObservationRepo: repo,
		observationGetter:          observations,
	}

	return service, observations, repo
}

// Test cases
func TestRetrieveLatestObservation(t *testing.T) {
	t.Run("successful retrieval", func(t *testing.T) {
		service, mockNWSClient, mockRepo := newTestService()
		stationID := "KNYC"
		timestamp := time.Now()

		mockNWSClient.On("GetLatestObservations", stationID).Return(&nws.Observation{
			Properties: nws.ObservationProperties{
				Temperature: &nws.QuantitativeValue{
					Value:    20.5,
					UnitCode: "wmoUnit:degC",
				},
				Timestamp: timestamp,
			},
		}, nil)

		expectedObs := weather_domain.NewTemperatureObservation(
			stationID,
			weather_domain.Temperature{Value: 20.5, TemperatureUnit: weather_domain.Celsius},
			timestamp,
		)

		mockRepo.On("Persist", mock.MatchedBy(func(obs *weather_domain.TemperatureObservation) bool {
			return obs.StationID == expectedObs.StationID &&
				obs.Temperature.Value == expectedObs.Temperature.Value &&
				obs.Timestamp.Equal(expectedObs.Timestamp)
		})).Return(nil)

		observation, err := service.RetrieveLatestObservation(stationID)

		assert.NoError(t, err)
		assert.NotNil(t, observation)
		assert.Equal(t, expectedObs.StationID, observation.StationID)
		assert.Equal(t, expectedObs.Temperature.Value, observation.Temperature.Value)
		assert.True(t, expectedObs.Timestamp.Equal(observation.Timestamp))

		mockNWSClient.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("NWS client error", func(t *testing.T) {
		service, mockNWSClient, _ := newTestService()
		stationID := "KNYC"
		mockNWSClient.On("GetLatestObservations", stationID).Return(nil, errors.New("NWS API error"))

		observation, err := service.RetrieveLatestObservation(stationID)

		assert.Error(t, err)
		assert.Nil(t, observation)

		mockNWSClient.AssertExpectations(t)
	})

	t.Run("missing temperature data", func(t *testing.T) {
		service, mockNWSClient, _ := newTestService()
		stationID := "KNYC"
		mockNWSClient.On("GetLatestObservations", stationID).Return(&nws.Observation{
			Properties: nws.ObservationProperties{
				Temperature: nil,
				Timestamp:   time.Now(),
			},
		}, nil)

		observation, err := service.RetrieveLatestObservation(stationID)

		assert.Error(t, err)
		assert.Nil(t, observation)
		assert.Contains(t, err.Error(), "temperature observation not found")

		mockNWSClient.AssertExpectations(t)
	})
}

func TestRetrieveObservationsInRange(t *testing.T) {
	t.Run("successful retrieval", func(t *testing.T) {
		service, mockNWSClient, mockRepo := newTestService()
		stationID := "KNYC"
		startTime := time.Now().Add(-24 * time.Hour)
		endTime := time.Now()

		observations := &nws.ObservationCollection{
			Features: []nws.Observation{
				{
					Properties: nws.ObservationProperties{
						Temperature: &nws.QuantitativeValue{
							Value:    20.5,
							UnitCode: "wmoUnit:degC",
						},
						Timestamp: startTime.Add(time.Hour),
					},
				},
				{
					Properties: nws.ObservationProperties{
						Temperature: &nws.QuantitativeValue{
							Value:    21.5,
							UnitCode: "wmoUnit:degC",
						},
						Timestamp: startTime.Add(2 * time.Hour),
					},
				},
			},
		}

		mockNWSClient.On("GetObservations", stationID, mock.MatchedBy(func(params nws.ObservationQueryParams) bool {
			return params.Start.Equal(startTime) && params.End.Equal(endTime)
		})).Return(observations, nil)

		mockRepo.On("Persist", mock.Anything).Return(nil).Times(2)

		results, stats, err := service.RetrieveObservationsInRange(stationID, startTime, endTime)

		assert.NoError(t, err)
		assert.NotNil(t, results)
		assert.Len(t, results, 2)
		assert.Equal(t, 2, stats.TotalObservations)
		assert.Equal(t, 2, stats.StoredObservations)
		assert.Equal(t, 0, stats.MissingTemperature)
		assert.Empty(t, stats.ObservationsWithError)

		mockNWSClient.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})

	t.Run("partial success with missing temperature", func(t *testing.T) {
		service, mockNWSClient, mockRepo := newTestService()
		stationID := "KNYC"
		startTime := time.Now().Add(-24 * time.Hour)
		endTime := time.Now()

		observations := &nws.ObservationCollection{
			Features: []nws.Observation{
				{
					Properties: nws.ObservationProperties{
						Temperature: nil,
						Timestamp:   startTime.Add(time.Hour),
					},
				},
				{
					Properties: nws.ObservationProperties{
						Temperature: &nws.QuantitativeValue{
							Value:    21.5,
							UnitCode: "wmoUnit:degC",
						},
						Timestamp: startTime.Add(2 * time.Hour),
					},
				},
			},
		}

		mockNWSClient.On("GetObservations", stationID, mock.Anything).Return(observations, nil)
		mockRepo.On("Persist", mock.Anything).Return(nil)

		results, stats, err := service.RetrieveObservationsInRange(stationID, startTime, endTime)

		assert.NoError(t, err)
		assert.Len(t, results, 1)
		assert.Equal(t, 2, stats.TotalObservations)
		assert.Equal(t, 1, stats.StoredObservations)
		assert.Equal(t, 1, stats.MissingTemperature)

		mockNWSClient.AssertExpectations(t)
		mockRepo.AssertExpectations(t)
	})
}

func TestGetLatestTemperature(t *testing.T) {

	t.Run("successful retrieval", func(t *testing.T) {
		service, _, mockRepo := newTestService()

		stationID := "KNYC"
		timestamp := time.Now()
		expectedObs := weather_domain.NewTemperatureObservation(
			stationID,
			weather_domain.Temperature{Value: 20.5, TemperatureUnit: weather_domain.Celsius},
			timestamp,
		)

		filter := &weather_domain.TemperatureObservationFilter{
			StationID: &stationID,
		}
		// returns the expected observation in a slice
		mockRepo.On("Get", filter).Return([]*weather_domain.TemperatureObservation{expectedObs}, nil)

		observation, err := service.GetLatestTemperature(stationID)

		assert.NoError(t, err)
		assert.NotNil(t, observation)
		assert.Equal(t, expectedObs.StationID, observation.StationID)
		assert.Equal(t, expectedObs.Temperature.Value, observation.Temperature.Value)
		assert.True(t, expectedObs.Timestamp.Equal(observation.Timestamp))

		mockRepo.AssertExpectations(t)
	})

	t.Run("not found", func(t *testing.T) {
		service, _, mockRepo := newTestService()

		stationID := "KNYC"
		filter := &weather_domain.TemperatureObservationFilter{
			StationID: &stationID,
		}

		// Mock Get to return empty slice with no error
		mockRepo.On("Get", filter).Return([]*weather_domain.TemperatureObservation{}, nil)

		observation, err := service.GetLatestTemperature(stationID)
		assert.Error(t, err)
		assert.Nil(t, observation)
		mockRepo.AssertExpectations(t)
	})
}
