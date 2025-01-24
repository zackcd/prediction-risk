package weather_repository

import (
	"fmt"
	weather_domain "prediction-risk/internal/app/weather/domain"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type TemperatureObservationRepo interface {
	Get(filter *weather_domain.TemperatureObservationFilter) ([]*weather_domain.TemperatureObservation, error)
	Persist(observation *weather_domain.TemperatureObservation) error
}

type temperatureObservationRepo struct {
	db *sqlx.DB
}

func NewTemperatureObservationRepo(db *sqlx.DB) *temperatureObservationRepo {
	return &temperatureObservationRepo{db}
}

type dbTemperatureObservation struct {
	ObservationID   string    `db:"observation_id"`
	StationID       string    `db:"station_id"`
	Temperature     float64   `db:"temperature"`
	TemperatureUnit string    `db:"temperature_unit"`
	Timestamp       time.Time `db:"timestamp"`
	CreatedAt       time.Time `db:"created_at"`
	UpdatedAt       time.Time `db:"updated_at"`
}

func (d dbTemperatureObservation) toDomain() *weather_domain.TemperatureObservation {
	observationID, _ := uuid.Parse(d.ObservationID)

	return &weather_domain.TemperatureObservation{
		ObservationID: weather_domain.ObservationID(observationID),
		StationID:     d.StationID,
		Temperature: weather_domain.Temperature{
			Value:           d.Temperature,
			TemperatureUnit: weather_domain.TemperatureUnit(d.TemperatureUnit),
		},
		Timestamp: d.Timestamp,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}

func (r *temperatureObservationRepo) Get(filter *weather_domain.TemperatureObservationFilter) ([]*weather_domain.TemperatureObservation, error) {
	query := `
		SELECT
			observation_id,
			station_id,
			temperature,
			temperature_unit,
			timestamp,
			created_at,
			updated_at
		FROM weather.temperature_observation
		WHERE 1=1
	`

	var args []interface{}
	var conditions []string

	if filter != nil && filter.StationID != nil {
		args = append(args, *filter.StationID)
		conditions = append(conditions, fmt.Sprintf("station_id = $%d", len(args)))
	}

	if len(conditions) > 0 {
		query += " AND " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY timestamp DESC"

	var dbObservations []dbTemperatureObservation
	if err := r.db.Select(&dbObservations, query, args...); err != nil {
		return nil, fmt.Errorf("error fetching temperature observations: %w", err)
	}

	observations := make([]*weather_domain.TemperatureObservation, len(dbObservations))
	for i, dbObs := range dbObservations {
		observations[i] = dbObs.toDomain()
	}

	return observations, nil
}

func (r *temperatureObservationRepo) Persist(observation *weather_domain.TemperatureObservation) error {
	// First try to get by observation_id
	var existing dbTemperatureObservation
	getQuery := `
        SELECT observation_id, station_id, temperature, temperature_unit, timestamp, created_at, updated_at
        FROM weather.temperature_observation
        WHERE observation_id = $1
    `
	err := r.db.Get(&existing, getQuery, observation.ObservationID.String())

	if err == nil {
		// If found, update the existing record
		updateQuery := `
            UPDATE weather.temperature_observation
            SET
                temperature = $1,
                temperature_unit = $2,
                updated_at = $3
            WHERE observation_id = $4
        `
		_, err = r.db.Exec(updateQuery,
			observation.Temperature.Value,
			observation.Temperature.TemperatureUnit,
			observation.UpdatedAt,
			observation.ObservationID.String(),
		)
		if err != nil {
			return fmt.Errorf("error updating temperature observation: %w", err)
		}
		return nil
	}

	// If not found by ID, try insert with station/timestamp conflict handling
	insertQuery := `
        INSERT INTO weather.temperature_observation (
            observation_id,
            station_id,
            temperature,
            temperature_unit,
            timestamp,
            created_at,
            updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7)
        ON CONFLICT ON CONSTRAINT unique_station_timestamp DO UPDATE
        SET
            temperature = EXCLUDED.temperature,
            temperature_unit = EXCLUDED.temperature_unit,
            updated_at = EXCLUDED.updated_at
    `

	_, err = r.db.Exec(insertQuery,
		observation.ObservationID.String(),
		observation.StationID,
		observation.Temperature.Value,
		observation.Temperature.TemperatureUnit,
		observation.Timestamp,
		observation.CreatedAt,
		observation.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("error persisting temperature observation: %w", err)
	}

	return nil
}
