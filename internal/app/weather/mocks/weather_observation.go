package weather_mocks

import (
	weather_domain "prediction-risk/internal/app/weather/domain"
	"time"

	"github.com/stretchr/testify/mock"
)

type MockWeatherObservationService struct {
	mock.Mock
}

func (m *MockWeatherObservationService) RetrieveLatestObservation(stationID string) (*weather_domain.TemperatureObservation, error) {
	args := m.Called(stationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*weather_domain.TemperatureObservation), args.Error(1)
}

func (m *MockWeatherObservationService) RetrieveObservationsInRange(stationID string, startTime time.Time, endTime time.Time) ([]*weather_domain.TemperatureObservation, *weather_domain.RetrievalStats, error) {
	args := m.Called(stationID, startTime, endTime)
	return args.Get(0).([]*weather_domain.TemperatureObservation), args.Get(1).(*weather_domain.RetrievalStats), args.Error(2)
}

func (m *MockWeatherObservationService) GetLatestTemperature(stationID string) (*weather_domain.TemperatureObservation, error) {
	args := m.Called(stationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*weather_domain.TemperatureObservation), args.Error(1)
}
