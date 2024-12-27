package services

import (
	"fmt"
	"log"
	"prediction-risk/internal/domain/entities"

	"github.com/google/uuid"
)

type StopLossOrderRepo interface {
	GetByID(id uuid.UUID) (*entities.StopLossOrder, error)
	GetAll() ([]*entities.StopLossOrder, error)
	Persist(stopLossOrder *entities.StopLossOrder) error
}

type StopLossOrderService interface {
	GetOrder(stopLossOrderId uuid.UUID) (*entities.StopLossOrder, error)
	GetActiveOrders() ([]*entities.StopLossOrder, error)
	CreateOrder(ticker string, side entities.Side, threshold entities.ContractPrice) (*entities.StopLossOrder, error)
	UpdateOrder(stopLossOrderId uuid.UUID, threshold entities.ContractPrice) (*entities.StopLossOrder, error)
	CancelOrder(stopLossOrderId uuid.UUID) (*entities.StopLossOrder, error)
	ExecuteOrder(stopLossOrderId uuid.UUID, isDryRun bool) (*entities.StopLossOrder, error)
}

type stopLossOrderService struct {
	repo     StopLossOrderRepo
	executor OrderExecutor
}

func NewStopLossService(
	repo StopLossOrderRepo,
	executor OrderExecutor,
) *stopLossOrderService {
	return &stopLossOrderService{
		repo:     repo,
		executor: executor,
	}
}

func (s *stopLossOrderService) GetOrder(
	stopLossOrderId uuid.UUID,
) (*entities.StopLossOrder, error) {
	log.Printf("Getting stop loss order: %s", stopLossOrderId)
	order, err := s.repo.GetByID(stopLossOrderId)
	if err != nil {
		log.Printf("Error getting stop loss order %s: %v", stopLossOrderId, err)
		return nil, err
	}
	return order, nil
}

func (s *stopLossOrderService) GetActiveOrders() ([]*entities.StopLossOrder, error) {
	log.Println("Getting all active stop loss orders")
	orders, err := s.repo.GetAll()
	if err != nil {
		log.Printf("Error getting all stop loss orders: %v", err)
		return nil, err
	}

	activeOrders := make([]*entities.StopLossOrder, 0, len(orders))
	for _, order := range orders {
		if order.Status() == entities.OrderStatusActive {
			activeOrders = append(activeOrders, order)
		}
	}

	log.Printf("Found %d active stop loss orders out of %d total orders", len(activeOrders), len(orders))
	return activeOrders, nil
}

func (s *stopLossOrderService) CreateOrder(
	ticker string,
	side entities.Side,
	threshold entities.ContractPrice,
) (*entities.StopLossOrder, error) {
	log.Printf("Creating stop loss order - ticker: %s, side: %s, threshold: %d",
		ticker, side, threshold.Value())

	order := entities.NewStopLossOrder(ticker, side, threshold)
	log.Printf("Created stop loss order %s", order.ID())

	if err := s.repo.Persist(order); err != nil {
		log.Printf("Error persisting stop loss order %s: %v", order.ID(), err)
		return nil, err
	}

	return order, nil
}

func (s *stopLossOrderService) UpdateOrder(
	stopLossOrderId uuid.UUID,
	threshold entities.ContractPrice,
) (*entities.StopLossOrder, error) {
	log.Printf("Updating stop loss order %s with new threshold: %d", stopLossOrderId, threshold.Value())

	order, err := s.repo.GetByID(stopLossOrderId)
	if err != nil {
		log.Printf("Error getting stop loss order %s for update: %v", stopLossOrderId, err)
		return nil, err
	}

	oldThreshold := order.TriggerPrice().Value()
	order.UpdateTriggerPrice(threshold)
	log.Printf("Updated threshold for order %s: %d -> %d", order.ID(), oldThreshold, threshold.Value())

	if err := s.repo.Persist(order); err != nil {
		log.Printf("Error persisting updated stop loss order %s: %v", order.ID(), err)
		return nil, err
	}

	return order, nil
}

func (s *stopLossOrderService) CancelOrder(
	stopLossOrderId uuid.UUID,
) (*entities.StopLossOrder, error) {
	log.Printf("Cancelling stop loss order %s", stopLossOrderId)

	order, err := s.repo.GetByID(stopLossOrderId)
	if err != nil {
		log.Printf("Error getting stop loss order %s for cancellation: %v", stopLossOrderId, err)
		return nil, err
	}

	if order.Status() != entities.OrderStatusActive {
		log.Printf("Cannot cancel order %s - invalid status: %s", order.ID(), order.Status())
		return nil, fmt.Errorf("order %s has invalid status %s", order.ID(), order.Status())
	}

	order.UpdateStatus(entities.OrderStatusCancelled)

	err = s.repo.Persist(order)
	if err != nil {
		log.Printf("Error persisting cancelled order %s: %v", order.ID(), err)
		return nil, fmt.Errorf("persisting cancelled order: %w", err)
	}

	return order, nil
}

func (s *stopLossOrderService) ExecuteOrder(
	stopLossOrderId uuid.UUID,
	isDryRun bool,
) (*entities.StopLossOrder, error) {
	log.Printf("Executing stop loss order %s (dry run: %v)", stopLossOrderId, isDryRun)

	order, err := s.repo.GetByID(stopLossOrderId)
	if err != nil {
		log.Printf("Error getting stop loss order %s for execution: %v", stopLossOrderId, err)
		return nil, err
	}

	if err := s.executor.ExecuteOrder(order, isDryRun); err != nil {
		log.Printf("Error executing order %s: %v", order.ID(), err)
		return nil, fmt.Errorf("executing order: %w", err)
	}

	order.UpdateStatus(entities.OrderStatusTriggered)
	err = s.repo.Persist(order)
	if err != nil {
		log.Printf("Error persisting executed order %s: %v", order.ID(), err)
		return nil, fmt.Errorf("persisting executed order: %w", err)
	}

	return order, nil
}
