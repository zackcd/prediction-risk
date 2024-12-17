package interfaces

import (
	"prediction-risk/internal/domain/entities"

	"github.com/google/uuid"
)

type StopLossOrderRepo interface {
	GetByID(id uuid.UUID) (*entities.StopLossOrder, error)
	GetAll() ([]*entities.StopLossOrder, error)
	Persist(stopLossOrder *entities.StopLossOrder) error
}
