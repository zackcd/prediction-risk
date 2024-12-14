package services

import (
	"fmt"
	"log"
	"prediction-risk/internal/domain"
	"prediction-risk/internal/domain/entities"
	"prediction-risk/internal/domain/repositories"

	"github.com/google/uuid"
)

type StopLossService struct {
	repo         repositories.StopLossOrderRepo
	kalshiClient domain.ExchangeClient
}

func NewStopLossService(repo repositories.StopLossOrderRepo) *StopLossService {
	return &StopLossService{repo: repo}
}

func (s *StopLossService) GetOrder(
	stopLossOrderId uuid.UUID,
) (
	*entities.StopLossOrder,
	error,
) {
	order, err := s.repo.GetByID(stopLossOrderId)
	if err != nil {
		return nil, err
	}

	return order, nil
}

func (s *StopLossService) CreateOrder(
	ticker string,
	threshold entities.ContractPrice,
) (
	*entities.StopLossOrder,
	error,
) {
	order := entities.NewStopLossOrder(ticker, threshold)

	if err := s.repo.Persist(order); err != nil {
		return nil, err
	}

	return order, nil
}

func (s *StopLossService) UpdateOrder(
	stopLossOrderId uuid.UUID,
	threshold entities.ContractPrice,
) (
	*entities.StopLossOrder,
	error,
) {
	order, err := s.repo.GetByID(stopLossOrderId)
	if err != nil {
		return nil, err
	}

	// Perform any necessary business logic here
	// Update the stop loss order
	order.SetThreshold(threshold)

	if err := s.repo.Persist(order); err != nil {
		return nil, err
	}

	return order, nil
}

func (s *StopLossService) CancelOrder(
	stopLossOrderId uuid.UUID,
) (
	*entities.StopLossOrder,
	error,
) {
	// Get the existing stop loss order
	order, err := s.repo.GetByID(stopLossOrderId)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, domain.NewErrNotFound("StopLossOrder", stopLossOrderId.String())
	}

	// Check if the order is already cancelled or executed
	if order.Status() != entities.StatusActive {
		return nil, fmt.Errorf("order %s has invalid status %s", order.ID(), order.Status())
	}

	order.SetStatus(entities.StatusCanceled)

	if err := s.repo.Persist(order); err != nil {
		return nil, fmt.Errorf("persisting canceled order: %w", err)
	}

	log.Printf("Order %s status set to canceled", order.ID())
	return order, nil
}

func (s *StopLossService) GetActiveOrders() ([]*entities.StopLossOrder, error) {
	orders, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	activeOrders := make([]*entities.StopLossOrder, 0, len(orders))
	for _, order := range orders {
		if order.Status() == entities.StatusActive {
			activeOrders = append(activeOrders, order)
		}
	}

	return activeOrders, nil
}

// ExecuteOrder executes the stop loss order
// 1. Get the stop loss order
// 2. Validate the order is active
// 3. Validate the current price is below the threshold
// 4. Retrieve the number of contracts held from the exchange
// 5. Execute the sell order
// 6. Update the stop loss order status to executed
func (s *StopLossService) ExecuteOrder(
	stopLossOrderId uuid.UUID,
) (
	*entities.StopLossOrder,
	error,
) {
	// Get the existing stop loss order
	order, err := s.repo.GetByID(stopLossOrderId)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, domain.NewErrNotFound("StopLossOrder", stopLossOrderId.String())
	}

	if order.Status() != entities.StatusActive {
		return nil, fmt.Errorf("order %s has invalid status %s", order.ID(), order.Status())
	}

	// Get the current price of the ticker

}
