package services

import (
	"log"
	"prediction-risk/internal/domain/entities"

	"github.com/google/uuid"
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

	oldThreshold := order.Threshold().Value()
	order.SetThreshold(threshold)
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
}

func (s *takeProfitService) ExecuteOrder(
	takeProfitOrderId uuid.UUID,
	isDryRun bool,
) (*entities.TakeProfitOrder, error) {
}
