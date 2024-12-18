package services

import (
	"fmt"
	"log"
	"prediction-risk/internal/domain/entities"
	"prediction-risk/internal/infrastructure/external/kalshi"

	"github.com/google/uuid"
	"github.com/samber/lo"
)

type StopLossOrderRepo interface {
	GetByID(id uuid.UUID) (*entities.StopLossOrder, error)
	GetAll() ([]*entities.StopLossOrder, error)
	Persist(stopLossOrder *entities.StopLossOrder) error
}

type StopLossService interface {
	GetOrder(stopLossOrderId uuid.UUID) (*entities.StopLossOrder, error)
	CreateOrder(ticker string, side entities.Side, threshold entities.ContractPrice) (*entities.StopLossOrder, error)
	UpdateOrder(stopLossOrderId uuid.UUID, threshold entities.ContractPrice) (*entities.StopLossOrder, error)
	CancelOrder(stopLossOrderId uuid.UUID) (*entities.StopLossOrder, error)
	GetActiveOrders() ([]*entities.StopLossOrder, error)
	ExecuteOrder(stopLossOrderId uuid.UUID) (*entities.StopLossOrder, error)
}

type stopLossService struct {
	repo     StopLossOrderRepo
	exchange ExchangeService
}

func NewStopLossService(
	repo StopLossOrderRepo,
	exchange ExchangeService,
) *stopLossService {
	return &stopLossService{
		repo:     repo,
		exchange: exchange,
	}
}

func (s *stopLossService) GetOrder(
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

func (s *stopLossService) CreateOrder(
	ticker string,
	side entities.Side,
	threshold entities.ContractPrice,
) (
	*entities.StopLossOrder,
	error,
) {
	order := entities.NewStopLossOrder(ticker, side, threshold)

	if err := s.repo.Persist(order); err != nil {
		return nil, err
	}

	return order, nil
}

func (s *stopLossService) UpdateOrder(
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

func (s *stopLossService) CancelOrder(
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
		return nil, entities.NewErrNotFound("StopLossOrder", stopLossOrderId.String())
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

func (s *stopLossService) GetActiveOrders() ([]*entities.StopLossOrder, error) {
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
// 3. Retrieve the number of contracts held from the exchange
// 4. Execute the sell order
// 5. Update the stop loss order status to executed
func (s *stopLossService) ExecuteOrder(
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
		return nil, entities.NewErrNotFound("StopLossOrder", stopLossOrderId.String())
	}

	if order.Status() != entities.StatusActive {
		return nil, fmt.Errorf("order %s has invalid status %s", order.ID(), order.Status())
	}

	positionsResp, err := s.exchange.GetPositions()
	if err != nil {
		return nil, fmt.Errorf("getting positions: %w", err)
	}

	position, found := lo.Find(positionsResp.MarketPositions, func(mp kalshi.MarketPosition) bool {
		return mp.Ticker == order.Ticker()
	})
	if !found {
		return nil, fmt.Errorf("no position found for ticker %s", order.Ticker())
	}

	count := abs(position.Position)

	// Execute the sell order
	_, err = s.exchange.CreateSellOrder(
		order.Ticker(),
		count,
		order.Side(),
		order.ID().String(),
	)
	if err != nil {
		return nil, fmt.Errorf("creating sell order: %w", err)
	}

	// Update the stop loss order status to executed and persist
	order.SetStatus(entities.StatusExecuted)
	err = s.repo.Persist(order)
	if err != nil {
		return nil, fmt.Errorf("persisting executed order: %w", err)
	}

	return order, nil
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
