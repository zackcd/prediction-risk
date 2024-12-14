package repositories

import (
	"prediction-risk/internal/domain/entities"
	"sync"

	"github.com/google/uuid"
)

type StopLossOrderRepoInMemory struct {
	data  map[uuid.UUID]*entities.StopLossOrder
	mutex sync.RWMutex
}

func NewStopLossRepoInMemory() *StopLossOrderRepoInMemory {
	return &StopLossOrderRepoInMemory{
		data: make(map[uuid.UUID]*entities.StopLossOrder),
	}
}

func (r *StopLossOrderRepoInMemory) GetByID(
	id uuid.UUID,
) (
	*entities.StopLossOrder,
	error,
) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	order, isFound := r.data[id]
	if !isFound {
		return nil, nil
	}

	orderCopy := *order
	return &orderCopy, nil
}

func (r *StopLossOrderRepoInMemory) Persist(
	stopLossOrder entities.StopLossOrder,
) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	orderCopy := stopLossOrder
	r.data[stopLossOrder.ID()] = &orderCopy

	return nil
}
