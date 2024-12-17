package services

import (
	"fmt"
	"log"
	entity "prediction-risk/internal/domain/entities"
	interfaces "prediction-risk/internal/domain/ports"
	"prediction-risk/internal/infrastructure/external/kalshi"

	"github.com/google/uuid"
	"github.com/samber/lo"
)

type StopLossService struct {
	repo   interfaces.StopLossOrderRepo
	kalshi *kalshi.KalshiClient
}

func NewStopLossService(repo interfaces.StopLossOrderRepo) *StopLossService {
	return &StopLossService{repo: repo}
}

func (s *StopLossService) GetOrder(
	stopLossOrderId uuid.UUID,
) (
	*entity.StopLossOrder,
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
	side entity.Side,
	threshold entity.ContractPrice,
) (
	*entity.StopLossOrder,
	error,
) {
	order := entity.NewStopLossOrder(ticker, side, threshold)

	if err := s.repo.Persist(order); err != nil {
		return nil, err
	}

	return order, nil
}

func (s *StopLossService) UpdateOrder(
	stopLossOrderId uuid.UUID,
	threshold entity.ContractPrice,
) (
	*entity.StopLossOrder,
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
	*entity.StopLossOrder,
	error,
) {
	// Get the existing stop loss order
	order, err := s.repo.GetByID(stopLossOrderId)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, entity.NewErrNotFound("StopLossOrder", stopLossOrderId.String())
	}

	// Check if the order is already cancelled or executed
	if order.Status() != entity.StatusActive {
		return nil, fmt.Errorf("order %s has invalid status %s", order.ID(), order.Status())
	}

	order.SetStatus(entity.StatusCanceled)

	if err := s.repo.Persist(order); err != nil {
		return nil, fmt.Errorf("persisting canceled order: %w", err)
	}

	log.Printf("Order %s status set to canceled", order.ID())
	return order, nil
}

func (s *StopLossService) GetActiveOrders() ([]*entity.StopLossOrder, error) {
	orders, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}

	activeOrders := make([]*entity.StopLossOrder, 0, len(orders))
	for _, order := range orders {
		if order.Status() == entity.StatusActive {
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
func (s *StopLossService) ExecuteOrder(
	stopLossOrderId uuid.UUID,
) (
	*entity.StopLossOrder,
	error,
) {
	// Get the existing stop loss order
	order, err := s.repo.GetByID(stopLossOrderId)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, entity.NewErrNotFound("StopLossOrder", stopLossOrderId.String())
	}

	if order.Status() != entity.StatusActive {
		return nil, fmt.Errorf("order %s has invalid status %s", order.ID(), order.Status())
	}

	positionsResp, err := s.kalshi.Portfolio.GetPositions(kalshi.GetPositionsOptions{})
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

	var orderSide kalshi.OrderSide
	if order.Side() == entity.SideYes {
		orderSide = kalshi.OrderSideYes
	} else {
		orderSide = kalshi.OrderSideNo
	}

	sellRequest := &kalshi.CreateOrderRequest{
		Ticker:        order.Ticker(),
		ClientOrderID: order.ID().String(),
		Side:          orderSide,
		Action:        kalshi.OrderActionSell,
		Count:         count,
		Type:          "market",
	}

	_, err = s.kalshi.Portfolio.CreateOrder(sellRequest)
	if err != nil {
		return nil, fmt.Errorf("creating sell order: %w", err)
	}

	// Update the stop loss order status to executed and persist
	order.SetStatus(entity.StatusExecuted)
	s.repo.Persist(order)

	return order, nil
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
