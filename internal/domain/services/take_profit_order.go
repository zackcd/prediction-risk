package services

import (
	"fmt"
	"log"
	"prediction-risk/internal/domain/entities"
	"prediction-risk/internal/infrastructure/external/kalshi"

	"github.com/google/uuid"
	"github.com/samber/lo"
)

type TakeProfitOrderRepo interface {
	GetByID(id uuid.UUID) (*entities.TakeProfitOrder, error)
	GetAll() ([]*entities.TakeProfitOrder, error)
	Persist(stopLossOrder *entities.TakeProfitOrder) error
}

type TakeProfitService interface {
	GetOrder(takeProfitOrderId uuid.UUID) (*entities.TakeProfitOrder, error)
	GetOrders() ([]*entities.TakeProfitOrder, error)
	CreateOrder(ticker string, side entities.Side, threshold entities.ContractPrice) (*entities.TakeProfitOrder, error)
	UpdateOrder(takeProfitOrderId uuid.UUID, threshold entities.ContractPrice) (*entities.TakeProfitOrder, error)
	CancelOrder(takeProfitOrderId uuid.UUID) (*entities.TakeProfitOrder, error)
	ExecuteOrder(takeProfitOrderId uuid.UUID, isDryRun bool) (*entities.TakeProfitOrder, error)
}

type takeProfitService struct {
	repo     TakeProfitOrderRepo
	exchange ExchangeService
}

func NewTakeProfitService() *takeProfitService {
	return &takeProfitService{}
}

func (s *takeProfitService) GetOrder(
	takeProfitOrderId uuid.UUID,
) (*entities.TakeProfitOrder, error) {
	log.Printf("Getting take profit order: %s", takeProfitOrderId)
	order, err := s.repo.GetByID(takeProfitOrderId)
	if err != nil {
		log.Printf("Error getting take profit order %s: %v", takeProfitOrderId, err)
		return nil, err
	}
	return order, nil
}

func (s *takeProfitService) GetOrders() ([]*entities.TakeProfitOrder, error) {
	log.Printf("Getting take profit orders")
	orders, err := s.repo.GetAll()
	if err != nil {
		log.Printf("Error getting all take profit orders: %v", err)
		return nil, err
	}

	log.Printf("Got %d take profit orders", len(orders))
	return orders, nil
}

func (s *takeProfitService) CreateOrder(
	ticker string,
	side entities.Side,
	threshold entities.ContractPrice,
) (*entities.TakeProfitOrder, error) {
	log.Printf("Creating take profit order - ticker: %s, side: %s, threshold: %d",
		ticker, side, threshold.Value())

	order := entities.NewTakeProfitOrder(ticker, side, threshold)
	log.Printf("Created take profit order: %s", order.ID())

	if err := s.repo.Persist(order); err != nil {
		log.Printf("Error persisting take profit order %s: %v", order.ID(), err)
		return nil, err
	}

	return order, nil
}

func (s *takeProfitService) UpdateOrder(
	takeProfitOrderId uuid.UUID,
	threshold entities.ContractPrice,
) (*entities.TakeProfitOrder, error) {
	log.Printf("Updating take profit order %s with threshold: %d", takeProfitOrderId, threshold.Value())

	order, err := s.GetOrder(takeProfitOrderId)
	if err != nil {
		log.Printf("Error getting take profit order %s: %v", takeProfitOrderId, err)
		return nil, err
	}

	oldThreshold := order.TriggerPrice().Value()
	order.UpdateTriggerPrice(threshold)
	log.Printf("Updated threshold for order %s: %d -> %d", order.ID(), oldThreshold, threshold.Value())

	if err := s.repo.Persist(order); err != nil {
		log.Printf("Error persisting updated take profit order %s: %v", order.ID(), err)
		return nil, err
	}

	return order, nil
}

func (s *takeProfitService) CancelOrder(
	takeProfitOrderId uuid.UUID,
) (*entities.TakeProfitOrder, error) {
	log.Printf("Cancelling take profit order %s", takeProfitOrderId)

	order, err := s.repo.GetByID(takeProfitOrderId)
	if err != nil {
		log.Printf("Error getting take profit order %s for cancellation: %v", takeProfitOrderId, err)
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

func (s *takeProfitService) ExecuteOrder(
	takeProfitOrderId uuid.UUID,
	isDryRun bool,
) (*entities.TakeProfitOrder, error) {
	log.Printf("Executing take profit order %s (dry run: %v)", takeProfitOrderId, isDryRun)

	order, err := s.GetOrder(takeProfitOrderId)
	if err != nil {
		log.Printf("Error getting take profit order %s: %v", takeProfitOrderId, err)
		return nil, err
	}

	if order.Status() != entities.OrderStatusActive {
		log.Printf("Cannot execute order %s - invalid status: %s", order.ID(), order.Status())
		return nil, fmt.Errorf("order %s has invalid status %s", order.ID(), order.Status())
	}

	if !isDryRun {
		// Execute the order
		log.Printf("Getting postitions for order %s execution", order.ID())
		positionsResp, err := s.exchange.GetPositions()
		if err != nil {
			log.Printf("Error getting positions for order %s: %v", order.ID(), err)
			return nil, fmt.Errorf("getting positions: %w", err)
		}

		position, found := lo.Find(positionsResp.MarketPositions, func(mp kalshi.MarketPosition) bool {
			return mp.Ticker == order.Ticker()
		})
		if !found {
			log.Printf("No position found for order %s", order.ID())
			return nil, fmt.Errorf("no position found for order %s", order.ID())
		}

		count := abs(position.Position)
		log.Printf("Executing sell order for %d contracts of %s", count, order.Ticker())

		_, err = s.exchange.CreateSellOrder(
			order.Ticker(),
			count,
			order.Side(),
			order.ID().String(),
		)
		if err != nil {
			log.Printf("Error creating sell order for %s: %v", order.ID(), err)
			return nil, fmt.Errorf("creating sell order: %w", err)
		}
		log.Printf("Successfully created sell order for stop loss %s", order.ID())
	}

	order.UpdateStatus(entities.OrderStatusTriggered)
	err = s.repo.Persist(order)
	if err != nil {
		log.Printf("Error persisting executed order %s: %v", order.ID(), err)
		return nil, fmt.Errorf("persisting executed order: %w", err)
	}

	return order, fmt.Errorf("not implemented")
}
