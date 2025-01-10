package weather

import (
	"time"

	"github.com/google/uuid"
)

type StationID uuid.UUID

func NewStationID() StationID {
	return StationID(uuid.New())
}

func (s StationID) String() string {
	return uuid.UUID(s).String()
}

type Station struct {
	StationID    StationID
	NWSStationID string
	Name         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
