package weather_repository

import (
	"database/sql"
	"fmt"
	weather_domain "prediction-risk/internal/app/weather/domain"
	"time"

	"github.com/jmoiron/sqlx"
)

type StationRepo interface {
	GetStationById(stationID string) (*weather_domain.NWSStation, error)
	Persist(station *weather_domain.NWSStation) error
}

type stationRepo struct {
	db *sqlx.DB
}

func NewStationRepo(db *sqlx.DB) *stationRepo {
	return &stationRepo{db}
}

// Database model for weather station
type dbStation struct {
	StationID string    `db:"station_id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

// Convert database model to domain model
func (s dbStation) toDomain() *weather_domain.NWSStation {
	return &weather_domain.NWSStation{
		StationID: s.StationID,
		Name:      s.Name,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}

func (r *stationRepo) GetByID(stationID string) (*weather_domain.NWSStation, error) {
	var station dbStation
	query := `
		SELECT station_id, name, created_at, updated_at
		FROM weather.nws_station
		WHERE station_id = $1
	`

	err := r.db.Get(&station, query, stationID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("station not found with ID %s", stationID)
		}
		return nil, fmt.Errorf("error fetching station: %w", err)
	}

	return station.toDomain(), nil
}

func (r *stationRepo) Persist(station *weather_domain.NWSStation) error {
	query := `
		INSERT INTO weather.nws_station (
			station_id, name, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4
		)
		ON CONFLICT (station_id) DO UPDATE
		SET
			name = EXCLUDED.name,
			updated_at = EXCLUDED.updated_at
	`

	_, err := r.db.Exec(query,
		station.StationID,
		station.Name,
		station.CreatedAt,
		station.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("error persisting station: %w", err)
	}

	return nil
}
