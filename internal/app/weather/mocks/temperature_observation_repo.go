package weather_mocks

import (
	weather_domain "prediction-risk/internal/app/weather/domain"
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

func (m *MockTemperatureObservationRepo) Get(filter *weather_domain.TemperatureObservationFilter) ([]*weather_domain.TemperatureObservation, error) {
	args := m.Called(filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*weather_domain.TemperatureObservation), args.Error(1)
}

func (m *MockTemperatureObservationRepo) Persist(observation *weather_domain.TemperatureObservation) error {
	args := m.Called(observation)
	return args.Error(0)
}
