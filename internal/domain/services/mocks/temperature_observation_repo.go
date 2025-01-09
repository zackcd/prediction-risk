package mocks

import (
	"prediction-risk/internal/domain/entities/weather"
	"testing"

	"github.com/stretchr/testify/mock"
)

type MockTemperatureObservationRepo struct {
	mock.Mock
}

func NewMockTemperatureObservationRepo(t *testing.T) *MockTemperatureObservationRepo {
	mock := &MockTemperatureObservationRepo{}
	mock.Test(t)
	return mock
}

func (m *MockTemperatureObservationRepo) Get(filter *weather.TemperatureObservationFilter) ([]*weather.TemperatureObservation, error) {
	args := m.Called(filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*weather.TemperatureObservation), args.Error(1)
}

func (m *MockTemperatureObservationRepo) GetLatestByStation(stationID weather.StationID) (*weather.TemperatureObservation, error) {
	args := m.Called(stationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*weather.TemperatureObservation), args.Error(1)
}

func (m *MockTemperatureObservationRepo) Persist(observation *weather.TemperatureObservation) error {
	args := m.Called(observation)
	return args.Error(0)
}
