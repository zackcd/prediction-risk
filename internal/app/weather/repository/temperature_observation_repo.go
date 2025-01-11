package weather_repository

import (
	"github.com/jmoiron/sqlx"
)

type TemperatureObservationRepo struct {
	db *sqlx.DB
}

func NewTemperatureObservationRepo(db *sqlx.DB) *TemperatureObservationRepo {
	return &TemperatureObservationRepo{db}
}
