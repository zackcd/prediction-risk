package order

import (
	"fmt"
	"log"
	"prediction-risk/internal/domain/contract"
	"prediction-risk/internal/infrastructure/external/kalshi"

	"github.com/samber/lo"
)

type StopOrderRepo interface {
	GetByID(orderId OrderID) (*StopOrder, error)
	GetAll() ([]*StopOrder, error)
	Persist(stopOrder *StopOrder) error
}

type StopOrderService interface {
	GetOrder(orderId OrderID) (*StopOrder, error)
	GetActiveOrders() ([]*StopOrder, error)
	CreateOrder(ticker string, side contract.Side, triggerPrice contract.ContractPrice, limitPrice *contract.ContractPrice) (*StopOrder, error)
	UpdateOrder(orderId OrderID, triggerPrice *contract.ContractPrice, limitPrice *contract.ContractPrice) (*StopOrder, error)
	CancelOrder(orderId OrderID) (*StopOrder, error)
	ExecuteOrder(orderId OrderID, isDryRun bool) (*StopOrder, error)
}

type stopOrderService struct {
	repo     StopOrderRepo
	exchange ExchangeProvider
}

func NewStopOrderService(
	repo StopOrderRepo,
	exchange ExchangeProvider,
) *stopOrderService {
	return &stopOrderService{
		repo:     repo,
		exchange: exchange,
	}
}

func (s *stopOrderService) GetOrder(
	orderId OrderID,
) (*StopOrder, error) {
	log.Printf("Getting stop order: %s", orderId)
	order, err := s.repo.GetByID(orderId)
	if err != nil {
		log.Printf("Error getting stop order %s: %v", orderId, err)
		return nil, err
	}
	return order, nil
}

func (s *stopOrderService) GetActiveOrders() ([]*StopOrder, error) {
	log.Println("Getting all active stop orders")
	orders, err := s.repo.GetAll()
	if err != nil {
		log.Printf("Error getting all stop orders: %v", err)
		return nil, err
	}

	activeOrders := make([]*StopOrder, 0, len(orders))
	for _, order := range orders {
		if order.Status() == OrderStatusActive {
			activeOrders = append(activeOrders, order)
		}
	}

	log.Printf("Found %d active stop orders out of %d total orders", len(activeOrders), len(orders))
	return activeOrders, nil
}

func (s *stopOrderService) CreateOrder(
	ticker string,
	side contract.Side,
	triggerPrice contract.ContractPrice,
	limitPrice *contract.ContractPrice,
) (*StopOrder, error) {
	log.Printf("Creating stop order - ticker: %s, side: %s, trigger price: %d, limit price: %p",
		ticker, side, triggerPrice.Value(), limitPrice)

	order := NewStopOrder(ticker, side, triggerPrice, limitPrice, nil)
	log.Printf("Created stop order %s", order.ID())

	if err := s.repo.Persist(order); err != nil {
		log.Printf("Error persisting stop order %s: %v", order.ID(), err)
		return nil, err
	}

	return order, nil
}

func (s *stopOrderService) UpdateOrder(
	orderId OrderID,
	triggerPrice *contract.ContractPrice,
	limitPrice *contract.ContractPrice,
) (*StopOrder, error) {
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
	orderId OrderID,
) (*StopOrder, error) {
	log.Printf("Cancelling stop order %s", orderId)

	order, err := s.repo.GetByID(orderId)
	if err != nil {
		log.Printf("Error getting stop order %s for cancellation: %v", orderId, err)
		return nil, err
	}

	if order.Status() != OrderStatusActive {
		log.Printf("Cannot cancel order %s - invalid status: %s", order.ID(), order.Status())
		return nil, fmt.Errorf("order %s has invalid status %s", order.ID(), order.Status())
	}

	order.UpdateStatus(OrderStatusCancelled)

	err = s.repo.Persist(order)
	if err != nil {
		log.Printf("Error persisting cancelled order %s: %v", order.ID(), err)
		return nil, fmt.Errorf("persisting cancelled order: %w", err)
	}

	return order, nil
}

func (s *stopOrderService) ExecuteOrder(
	orderId OrderID,
	isDryRun bool,
) (*StopOrder, error) {
	log.Printf("Executing stop order %s (dry run: %v)", orderId, isDryRun)

	order, err := s.repo.GetByID(orderId)
	if err != nil {
		log.Printf("Error getting stop order %s for execution: %v", orderId, err)
		return nil, err
	}

	log.Printf("Getting positions for order %s execution", order.ID())
	positionsResp, err := s.exchange.GetPositions()
	if err != nil {
		log.Printf("Error getting positions for order %s: %v", order.ID(), err)
		return nil, fmt.Errorf("getting positions: %w", err)
	}

	position, found := lo.Find(positionsResp.MarketPositions, func(mp kalshi.MarketPosition) bool {
		return mp.Ticker == order.Ticker()
	})
	if !found {
		log.Printf("No position found for ticker %s", order.Ticker())
		return nil, fmt.Errorf("no position found for ticker %s", order.Ticker())
	}

	// Check if order specifies a quantity
	count := abs(position.Position)
	log.Printf("Executing stop order for %d contracts of %s", count, order.Ticker())

	_, err = s.exchange.CreateSellOrder(
		order.Ticker(),
		count,
		order.Side(),
		order.ID().String(),
		order.LimitPrice(),
	)
	if err != nil {
		log.Printf("Error executing stop order for %s: %v", order.ID(), err)
		return nil, fmt.Errorf("executing stop order: %w", err)
	}

	log.Printf("Successfully  executed stop order %s", order.ID())

	order.UpdateStatus(OrderStatusTriggered)
	err = s.repo.Persist(order)
	if err != nil {
		log.Printf("Error persisting executed order %s: %v", order.ID(), err)
		return nil, fmt.Errorf("persisting executed order: %w", err)
	}

	return order, nil
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
