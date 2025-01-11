package weather_domain

import (
	"time"

	"github.com/google/uuid"
)

type ObservationID uuid.UUID

func NewObservationID() ObservationID {
	return ObservationID(uuid.New())
}

func (o ObservationID) String() string {
	return uuid.UUID(o).String()
}

type TemperatureUnit string

const (
	Celsius    TemperatureUnit = "CELSIUS"
	Fahrenheit TemperatureUnit = "FAHRENHEIT"
)

type Temperature struct {
	Value           float64
	TemperatureUnit TemperatureUnit
}

type TemperatureObservation struct {
	ObservationID ObservationID
	StationID     StationID
	Temperature   Temperature
	Timestamp     time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func NewTemperatureObservation(
	stationID StationID,
	temperature Temperature,
	timestamp time.Time,
) *TemperatureObservation {
	currentTime := time.Now()
	return &TemperatureObservation{
		ObservationID: NewObservationID(),
		StationID:     stationID,
		Temperature:   temperature,
		Timestamp:     timestamp,
		CreatedAt:     currentTime,
		UpdatedAt:     currentTime,
	}
}

type TemperatureObservationFilter struct{}
