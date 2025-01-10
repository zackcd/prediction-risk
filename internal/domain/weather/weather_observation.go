package weather

import (
	"fmt"
	"prediction-risk/internal/infrastructure/external/nws"
	"time"

	"github.com/samber/lo"
)

type TemperatureObservationRepo interface {
	Get(filter *TemperatureObservationFilter) ([]*TemperatureObservation, error)
	GetLatestByStation(stationID StationID) (*TemperatureObservation, error)
	Persist(observation *TemperatureObservation) error
}

type WeatherObservationService interface{}

type RetrievalStats struct {
	TotalObservations     int
	MissingTemperature    int
	StoredObservations    int
	ObservationsWithError []time.Time
}

type weatherObservationService struct {
	temperatureObservationRepo TemperatureObservationRepo
	nwsClient                  *nws.NWSClient
}

func NewWeatherObservationService(
	temperatureObservationRepo TemperatureObservationRepo,
	nwsClient *nws.NWSClient,
) *weatherObservationService {
	return &weatherObservationService{
		temperatureObservationRepo: temperatureObservationRepo,
		nwsClient:                  nwsClient,
	}
}

// RetrieveLatestObservation gets and stores the latest observation for a station
func (s *weatherObservationService) RetrieveLatestObservation(
	stationID StationID,
) (*TemperatureObservation, error) {
	observation, err := s.nwsClient.Station.GetLatestObservations(stationID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve latest observations from NWS: %w", err)
	}

	temperatureProperty := observation.Properties.Temperature

	if temperatureProperty == nil {
		return nil, fmt.Errorf("temperature observation not found for station %s", stationID)
	}

	if temperatureProperty.UnitCode != "wmoUnit:degC" {
		return nil, fmt.Errorf("unsupported temperature unit from NWS: %s", temperatureProperty.UnitCode)
	}
	temperature := Temperature{
		Value:           temperatureProperty.Value,
		TemperatureUnit: Celsius,
	}
	temperatureObservation := NewTemperatureObservation(stationID, temperature, observation.Properties.Timestamp)

	if err := s.temperatureObservationRepo.Persist(temperatureObservation); err != nil {
		return nil, fmt.Errorf("storing temperature observation: %w", err)
	}

	return temperatureObservation, nil
}

// RetrieveObservationsInRange gets and stores observations within a specific time range
func (s *weatherObservationService) RetrieveObservationsInRange(
	stationID StationID,
	startTime time.Time,
	endTime time.Time,
) ([]*TemperatureObservation, *RetrievalStats, error) {
	params := nws.ObservationQueryParams{
		Start: &startTime,
		End:   &endTime,
	}

	observations, err := s.nwsClient.Station.GetObservations(stationID.String(), params)
	if err != nil {
		return nil, nil, fmt.Errorf("fetching observations from NWS: %w", err)
	}

	stats := &RetrievalStats{
		TotalObservations: len(observations.Features),
	}

	// Process all observations and track missing data
	results := lo.FilterMap(observations.Features, func(obs nws.Observation, _ int) (*TemperatureObservation, bool) {
		if obs.Properties.Temperature == nil {
			stats.MissingTemperature++
			fmt.Printf("Missing temperature data for station %s at %v\n",
				stationID, obs.Properties.Timestamp)
			return nil, false
		}

		if obs.Properties.Temperature.UnitCode != "wmoUnit:degC" {
			fmt.Printf("Unsupported temperature unit from NWS: %s\n", obs.Properties.Temperature.UnitCode)
			return nil, false
		}

		temperature := Temperature{
			Value:           obs.Properties.Temperature.Value,
			TemperatureUnit: Celsius,
		}
		return NewTemperatureObservation(stationID, temperature, obs.Properties.Timestamp), true
	})

	// Store the observations and track any errors
	for _, tempObs := range results {
		if err := s.temperatureObservationRepo.Persist(tempObs); err != nil {
			stats.ObservationsWithError = append(stats.ObservationsWithError, tempObs.Timestamp)
			fmt.Printf("Failed to store observation for station %s at %v: %v\n",
				stationID, tempObs.Timestamp, err)
		} else {
			stats.StoredObservations++
		}
	}

	// Return error if no observations were successfully stored
	if stats.StoredObservations == 0 {
		return nil, stats, fmt.Errorf("failed to store any observations out of %d total (%d missing temperature, %d storage errors)",
			stats.TotalObservations, stats.MissingTemperature, len(stats.ObservationsWithError))
	}

	return results, stats, nil
}

// GetLatestTemperature retrieves the most recent stored observation for a station
func (s *weatherObservationService) GetLatestTemperature(
	stationID StationID,
) (*TemperatureObservation, error) {
	return s.temperatureObservationRepo.GetLatestByStation(stationID)
}
