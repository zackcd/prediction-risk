package weather_mocks

import (
	"prediction-risk/internal/app/weather/infrastructure/nws"

	"github.com/stretchr/testify/mock"
)

type MockObservationGetter struct {
	mock.Mock
}

func (m *MockObservationGetter) GetObservations(stationID string, params nws.ObservationQueryParams) (*nws.ObservationCollection, error) {
	args := m.Called(stationID, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*nws.ObservationCollection), args.Error(1)
}

func (m *MockObservationGetter) GetLatestObservations(stationID string) (*nws.Observation, error) {
	args := m.Called(stationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*nws.Observation), args.Error(1)
}
