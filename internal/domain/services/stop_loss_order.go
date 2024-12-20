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
	GetActiveOrders() ([]*entities.StopLossOrder, error)
	CreateOrder(ticker string, side entities.Side, threshold entities.ContractPrice) (*entities.StopLossOrder, error)
	UpdateOrder(stopLossOrderId uuid.UUID, threshold entities.ContractPrice) (*entities.StopLossOrder, error)
	CancelOrder(stopLossOrderId uuid.UUID) (*entities.StopLossOrder, error)
	ExecuteOrder(stopLossOrderId uuid.UUID, isDryRun bool) (*entities.StopLossOrder, error)
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
) (*entities.StopLossOrder, error) {
	log.Printf("Getting stop loss order: %s", stopLossOrderId)
	order, err := s.repo.GetByID(stopLossOrderId)
	if err != nil {
		log.Printf("Error getting stop loss order %s: %v", stopLossOrderId, err)
		return nil, err
	}
	return order, nil
}

func (s *stopLossService) GetActiveOrders() ([]*entities.StopLossOrder, error) {
	log.Println("Getting all active stop loss orders")
	orders, err := s.repo.GetAll()
	if err != nil {
		log.Printf("Error getting all stop loss orders: %v", err)
		return nil, err
	}

	activeOrders := make([]*entities.StopLossOrder, 0, len(orders))
	for _, order := range orders {
		if order.Status() == entities.SLOStatusActive {
			activeOrders = append(activeOrders, order)
		}
	}

	log.Printf("Found %d active stop loss orders out of %d total orders", len(activeOrders), len(orders))
	return activeOrders, nil
}

func (s *stopLossService) CreateOrder(
	ticker string, side entities.Side,
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

func (s *stopLossService) UpdateOrder(
	stopLossOrderId uuid.UUID,
	threshold entities.ContractPrice,
) (*entities.StopLossOrder, error) {
	log.Printf("Updating stop loss order %s with new threshold: %d", stopLossOrderId, threshold.Value())

	order, err := s.repo.GetByID(stopLossOrderId)
	if err != nil {
		log.Printf("Error getting stop loss order %s for update: %v", stopLossOrderId, err)
		return nil, err
	}

	oldThreshold := order.Threshold().Value()
	order.SetThreshold(threshold)
	log.Printf("Updated threshold for order %s: %d -> %d", order.ID(), oldThreshold, threshold.Value())

	if err := s.repo.Persist(order); err != nil {
		log.Printf("Error persisting updated stop loss order %s: %v", order.ID(), err)
		return nil, err
	}

	return order, nil
}

func (s *stopLossService) CancelOrder(
	stopLossOrderId uuid.UUID,
) (*entities.StopLossOrder, error) {
	log.Printf("Cancelling stop loss order %s", stopLossOrderId)

	order, err := s.repo.GetByID(stopLossOrderId)
	if err != nil {
		log.Printf("Error getting stop loss order %s for cancellation: %v", stopLossOrderId, err)
		return nil, err
	}

	if order.Status() != entities.SLOStatusActive {
		log.Printf("Cannot cancel order %s - invalid status: %s", order.ID(), order.Status())
		return nil, fmt.Errorf("order %s has invalid status %s", order.ID(), order.Status())
	}

	order.SetStatus(entities.SLOStatusCancelled)

	err = s.repo.Persist(order)
	if err != nil {
		log.Printf("Error persisting cancelled order %s: %v", order.ID(), err)
		return nil, fmt.Errorf("persisting cancelled order: %w", err)
	}

	return order, nil
}

func (s *stopLossService) ExecuteOrder(
	stopLossOrderId uuid.UUID,
	isDryRun bool,
) (*entities.StopLossOrder, error) {
	log.Printf("Executing stop loss order %s (dry run: %v)", stopLossOrderId, isDryRun)

	order, err := s.repo.GetByID(stopLossOrderId)
	if err != nil {
		log.Printf("Error getting stop loss order %s for execution: %v", stopLossOrderId, err)
		return nil, err
	}

	if order.Status() != entities.SLOStatusActive {
		log.Printf("Cannot execute order %s - invalid status: %s", order.ID(), order.Status())
		return nil, fmt.Errorf("order %s has invalid status %s", order.ID(), order.Status())
	}

	if !isDryRun {
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

	order.SetStatus(entities.SLOStatusExecuted)
	err = s.repo.Persist(order)
	if err != nil {
		log.Printf("Error persisting executed order %s: %v", order.ID(), err)
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
