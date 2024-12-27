package services

import (
	"fmt"
	"log"
	"prediction-risk/internal/domain/entities"
	"prediction-risk/internal/infrastructure/external/kalshi"

	"github.com/samber/lo"
)

type OrderExecutor interface {
	ExecuteOrder(order entities.Order, isDryRun bool) error
}

type orderExecutor struct {
	exchange ExchangeService
}

func NewOrderExecutor(exchange ExchangeService) *orderExecutor {
	return &orderExecutor{exchange: exchange}
}

func (e *orderExecutor) executeOrderOnExchange(order entities.Order) error {
	// Match on order type to determine execution flow
	switch order.Type() {
	case entities.OrderTypeStopLoss, entities.OrderTypeTakeProfit:
		return e.executeClosePosition(order)
	default:
		return fmt.Errorf("unsupported order type: %s", order.Type())
	}
}

func (e *orderExecutor) executeClosePosition(order entities.Order) error {
	log.Printf("Getting positions for order %s execution", order.ID())
	positionsResp, err := e.exchange.GetPositions()
	if err != nil {
		log.Printf("Error getting positions for order %s: %v", order.ID(), err)
		return fmt.Errorf("getting positions: %w", err)
	}

	position, found := lo.Find(positionsResp.MarketPositions, func(mp kalshi.MarketPosition) bool {
		return mp.Ticker == order.Ticker()
	})
	if !found {
		log.Printf("No position found for ticker %s", order.Ticker())
		return fmt.Errorf("no position found for ticker %s", order.Ticker())
	}

	// Check if order specifies a quantity
	count := abs(position.Position)
	log.Printf("Executing close position for %d contracts of %s", count, order.Ticker())

	_, err = e.exchange.CreateSellOrder(
		order.Ticker(),
		count,
		order.Side(),
		order.ID().String(),
	)
	if err != nil {
		log.Printf("Error closing position for %s: %v", order.ID(), err)
		return fmt.Errorf("closing position: %w", err)
	}

	log.Printf("Successfully closed position for order %s", order.ID())
	return nil
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
