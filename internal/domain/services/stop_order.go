package services

import (
	"fmt"
	"log"
	"prediction-risk/internal/domain/entities"

	"github.com/google/uuid"
)

type StopOrderRepo interface {
	GetByID(orderId uuid.UUID) (*entities.StopOrder, error)
	GetAll() ([]*entities.StopOrder, error)
	Persist(stopOrder *entities.StopOrder) error
}

type StopOrderService interface {
	GetOrder(orderId uuid.UUID) (*entities.StopOrder, error)
	GetActiveOrders() ([]*entities.StopOrder, error)
	CreateOrder(ticker string, side entities.Side, triggerPrice entities.ContractPrice, limitPrice *entities.ContractPrice) (*entities.StopOrder, error)
	UpdateOrder(orderId uuid.UUID, triggerPrice *entities.ContractPrice, limitPrice *entities.ContractPrice) (*entities.StopOrder, error)
	CancelOrder(orderId uuid.UUID) (*entities.StopOrder, error)
	ExecuteOrder(orderId uuid.UUID, isDryRun bool) (*entities.StopOrder, error)
}

type stopOrderService struct {
	repo     StopOrderRepo
	executor OrderExecutor
}

func NewStopOrderService(
	repo StopOrderRepo,
	executor OrderExecutor,
) *stopOrderService {
	return &stopOrderService{
		repo:     repo,
		executor: executor,
	}
}

func (s *stopOrderService) GetOrder(
	orderId uuid.UUID,
) (*entities.StopOrder, error) {
	log.Printf("Getting stop order: %s", orderId)
	order, err := s.repo.GetByID(orderId)
	if err != nil {
		log.Printf("Error getting stop order %s: %v", orderId, err)
		return nil, err
	}
	return order, nil
}

func (s *stopOrderService) GetActiveOrders() ([]*entities.StopOrder, error) {
	log.Println("Getting all active stop orders")
	orders, err := s.repo.GetAll()
	if err != nil {
		log.Printf("Error getting all stop orders: %v", err)
		return nil, err
	}

	activeOrders := make([]*entities.StopOrder, 0, len(orders))
	for _, order := range orders {
		if order.Status() == entities.OrderStatusActive {
			activeOrders = append(activeOrders, order)
		}
	}

	log.Printf("Found %d active stop orders out of %d total orders", len(activeOrders), len(orders))
	return activeOrders, nil
}

func (s *stopOrderService) CreateOrder(
	ticker string,
	side entities.Side,
	triggerPrice entities.ContractPrice,
	limitPrice *entities.ContractPrice,
) (*entities.StopOrder, error) {
	log.Printf("Creating stop order - ticker: %s, side: %s, trigger price: %d, limit price: %p",
		ticker, side, triggerPrice.Value(), limitPrice)

	order := entities.NewStopOrder(ticker, side, triggerPrice, limitPrice)
	log.Printf("Created stop order %s", order.ID())

	if err := s.repo.Persist(order); err != nil {
		log.Printf("Error persisting stop order %s: %v", order.ID(), err)
		return nil, err
	}

	return order, nil
}

func (s *stopOrderService) UpdateOrder(
	orderId uuid.UUID,
	triggerPrice *entities.ContractPrice,
	limitPrice *entities.ContractPrice,
) (*entities.StopOrder, error) {
	log.Printf("Updating stop order %s", orderId)

	order, err := s.repo.GetByID(orderId)
	if err != nil {
		log.Printf("Error getting stop order %s for update: %v", orderId, err)
		return nil, err
	}

	if triggerPrice != nil {
		order.SetTriggerPrice(*triggerPrice)
	}
	order.SetLimitPrice(limitPrice)

	if err := s.repo.Persist(order); err != nil {
		log.Printf("Error persisting updated stop order %s: %v", order.ID(), err)
		return nil, err
	}

	return order, nil
}

func (s *stopOrderService) CancelOrder(
	orderId uuid.UUID,
) (*entities.StopOrder, error) {
	log.Printf("Cancelling stop order %s", orderId)

	order, err := s.repo.GetByID(orderId)
	if err != nil {
		log.Printf("Error getting stop order %s for cancellation: %v", orderId, err)
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

func (s *stopOrderService) ExecuteOrder(
	orderId uuid.UUID,
	isDryRun bool,
) (*entities.StopOrder, error) {
	log.Printf("Executing stop order %s (dry run: %v)", orderId, isDryRun)

	order, err := s.repo.GetByID(orderId)
	if err != nil {
		log.Printf("Error getting stop order %s for execution: %v", orderId, err)
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
