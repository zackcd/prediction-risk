package weather_service

import (
	"fmt"
	"log"
	weather_domain "prediction-risk/internal/app/weather/domain"
	"time"
)

type WeatherMonitor struct {
	stationID                 string
	weatherObservationService WeatherObservationService
	interval                  time.Duration
	done                      chan struct{}
}

func NewWeatherMonitor(
	stationID string,
	weatherObservationService WeatherObservationService,
	interval time.Duration,
) *WeatherMonitor {
	return &WeatherMonitor{
		stationID:                 stationID,
		weatherObservationService: weatherObservationService,
		interval:                  interval,
		done:                      make(chan struct{}),
	}
}

func (m *WeatherMonitor) Start() {
	log.Printf("Starting WeatherMonitor for station: %v", m.stationID)

	// First get any missed observations from the last 24 hours
	startTime := time.Now().UTC().Add(-24 * time.Hour)
	endTime := time.Now().UTC()

	observations, stats, err := m.weatherObservationService.RetrieveObservationsInRange(
		m.stationID,
		startTime,
		endTime,
	)
	if err != nil {
		log.Printf("Error retrieving historical observations: %v", err)
	} else {
		log.Printf("Retrieved %d historical observations (stored: %d, missing temp: %d, errors: %d)",
			stats.TotalObservations,
			stats.StoredObservations,
			stats.MissingTemperature,
			len(stats.ObservationsWithError),
		)

		// Process historical observations
		for _, observation := range observations {
			if _, err := m.processWeatherObservation(observation); err != nil {
				log.Printf("Error processing historical observation: %v", err)
			}
		}
	}

	go func() {
		ticker := time.NewTicker(m.interval)
		defer ticker.Stop()

		for {
			select {
			case <-m.done:
				log.Println("WeatherMonitor stopped")
				return
			case <-ticker.C:
				log.Println("Running weather observation check...")
				if err := m.checkWeatherObservation(); err != nil {
					log.Printf("Error checking weather observation: %v", err)
				}
			}
		}
	}()
}

func (m *WeatherMonitor) Stop() {
	log.Println("Stopping WeatherMonitor...")
	close(m.done)
}

func (m *WeatherMonitor) checkWeatherObservation() error {
	observation, err := m.weatherObservationService.RetrieveLatestObservation(m.stationID)
	if err != nil {
		return fmt.Errorf("failed to retrieve latest weather observation: %w", err)
	}
	log.Printf("Retrieved latest weather observation: %v", observation)

	_, err = m.processWeatherObservation(observation)
	if err != nil {
		return fmt.Errorf("failed to process weather observation: %w", err)
	}

	return nil
}

func (m *WeatherMonitor) processWeatherObservation(
	observation *weather_domain.TemperatureObservation,
) (*weather_domain.TemperatureObservation, error) {
	log.Printf("Processing weather observation: %v", observation)
	return observation, nil
}
