package inmemory

import (
	"prediction-risk/internal/domain/entities"
	"sync"

	"github.com/google/uuid"
)

type StopOrderRepoInMemory struct {
	data  map[uuid.UUID]*entities.StopOrder
	mutex sync.RWMutex
}

func NewStopOrderRepoInMemory() *StopOrderRepoInMemory {
	return &StopOrderRepoInMemory{
		data: make(map[uuid.UUID]*entities.StopOrder),
	}
}

func (r *StopOrderRepoInMemory) GetByID(
	id uuid.UUID,
) (
	*entities.StopOrder,
	error,
) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	order, isFound := r.data[id]
	if !isFound {
		return nil, entities.NewErrNotFound("StopOrder", id.String())
	}

	orderCopy := *order
	return &orderCopy, nil
}

func (r *StopOrderRepoInMemory) GetAll() (
	[]*entities.StopOrder,
	error,
) {
	r.mutex.RLock()
	defer r.mutex.RUnlock()

	orders := make([]*entities.StopOrder, 0, len(r.data))
	for _, order := range r.data {
		orderCopy := *order
		orders = append(orders, &orderCopy)
	}

	return orders, nil
}

func (r *StopOrderRepoInMemory) Persist(
	stopOrder *entities.StopOrder,
) error {
	r.mutex.Lock()
	defer r.mutex.Unlock()

	orderCopy := stopOrder
	r.data[stopOrder.ID()] = orderCopy

	return nil
}
