package weather_service

import (
	"fmt"
	weather_domain "prediction-risk/internal/app/weather/domain"
	"prediction-risk/internal/app/weather/infrastructure/nws"
	weather_repository "prediction-risk/internal/app/weather/repository"
	"time"

	"github.com/samber/lo"
)

type WeatherObservationService interface {
	RetrieveLatestObservation(stationID string) (*weather_domain.TemperatureObservation, error)
	RetrieveObservationsInRange(stationID string, startTime time.Time, endTime time.Time) ([]*weather_domain.TemperatureObservation, *weather_domain.RetrievalStats, error)
	GetLatestTemperature(stationID string) (*weather_domain.TemperatureObservation, error)
}

type ObservationGetter interface {
	GetLatestObservations(stationID string) (*nws.Observation, error)
	GetObservations(stationID string, params nws.ObservationQueryParams) (*nws.ObservationCollection, error)
}

type weatherObservationService struct {
	temperatureObservationRepo weather_repository.TemperatureObservationRepo
	observationGetter          ObservationGetter
}

func NewWeatherObservationService(
	temperatureObservationRepo weather_repository.TemperatureObservationRepo,
	nwsClient *nws.NWSClient,
) *weatherObservationService {
	return &weatherObservationService{
		temperatureObservationRepo: temperatureObservationRepo,
		observationGetter:          nwsClient.Station,
	}
}

// RetrieveLatestObservation gets and stores the latest observation for a station
func (s *weatherObservationService) RetrieveLatestObservation(
	stationID string,
) (*weather_domain.TemperatureObservation, error) {
	observation, err := s.observationGetter.GetLatestObservations(stationID)
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
	temperature := weather_domain.Temperature{
		Value:           temperatureProperty.Value,
		TemperatureUnit: weather_domain.Celsius,
	}
	temperatureObservation := weather_domain.NewTemperatureObservation(stationID, temperature, observation.Properties.Timestamp)

	if err := s.temperatureObservationRepo.Persist(temperatureObservation); err != nil {
		return nil, fmt.Errorf("storing temperature observation: %w", err)
	}

	return temperatureObservation, nil
}

// RetrieveObservationsInRange gets and stores observations within a specific time range
func (s *weatherObservationService) RetrieveObservationsInRange(
	stationID string,
	startTime time.Time,
	endTime time.Time,
) ([]*weather_domain.TemperatureObservation, *weather_domain.RetrievalStats, error) {
	params := nws.ObservationQueryParams{
		Start: &startTime,
		End:   &endTime,
	}

	observations, err := s.observationGetter.GetObservations(stationID, params)
	if err != nil {
		return nil, nil, fmt.Errorf("fetching observations from NWS: %w", err)
	}

	stats := &weather_domain.RetrievalStats{
		TotalObservations: len(observations.Features),
	}

	// Process all observations and track missing data
	results := lo.FilterMap(observations.Features, func(obs nws.Observation, _ int) (*weather_domain.TemperatureObservation, bool) {
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

		temperature := weather_domain.Temperature{
			Value:           obs.Properties.Temperature.Value,
			TemperatureUnit: weather_domain.Celsius,
		}
		return weather_domain.NewTemperatureObservation(stationID, temperature, obs.Properties.Timestamp), true
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
	stationID string,
) (*weather_domain.TemperatureObservation, error) {
	filter := &weather_domain.TemperatureObservationFilter{
		StationID: &stationID,
	}
	observations, err := s.temperatureObservationRepo.Get(filter)
	if err != nil {
		return nil, fmt.Errorf("fetching temperature observations: %w", err)
	}

	if len(observations) == 0 {
		return nil, fmt.Errorf("no temperature observations found for station %s", stationID)
	}

	latest := lo.MaxBy(observations, func(a *weather_domain.TemperatureObservation, b *weather_domain.TemperatureObservation) bool {
		return a.Timestamp.After(b.Timestamp)
	})

	return latest, nil
}
