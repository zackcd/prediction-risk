package weather_domain

import (
	"time"
)

type NWSStation struct {
	StationID string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}
