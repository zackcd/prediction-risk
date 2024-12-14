package services

import (
	"fmt"
	"prediction-risk/internal/domain/entities"
	"prediction-risk/internal/infrastructure/external/kalshi"
	"time"
)

type StopLossOrderMonitor struct {
	stopLossService *StopLossService
	kalshiClient    *kalshi.KalshiClient
	interval        time.Duration
	done            chan struct{}
}

func NewStopLossOrderMonitor(
	StopLossService *StopLossService,
	kalshiClient *kalshi.KalshiClient,
	interval time.Duration,
) *StopLossOrderMonitor {
	return &StopLossOrderMonitor{
		stopLossService: StopLossService,
		kalshiClient:    kalshiClient,
		interval:        interval,
		done:            make(chan struct{}),
	}
}

func (m *StopLossOrderMonitor) Start() {
	go func() {
		ticker := time.NewTicker(m.interval)
		defer ticker.Stop()

		for {
			select {
			case <-m.done:
				return
			case <-ticker.C:
				if err := m.checkOrders(); err != nil {
					fmt.Printf("error checking orders: %v\n", err)
				}
			}
		}
	}()
}

func (m *StopLossOrderMonitor) Stop() {
	close(m.done)
}

func (m *StopLossOrderMonitor) checkOrders() error {
	activeOrders, err := m.stopLossService.GetActiveOrders()
	if err != nil {
		return fmt.Errorf("getting active orders: %w", err)
	}

	for _, order := range activeOrders {
		market, err := m.kalshiClient.Market.GetMarket(order.Ticker())
		if err != nil {
			fmt.Printf("Error getting market data for %s: %v", order.Ticker, err)
			continue // Skip this order if we can't get the price, but keep checking others
		}

		if m.shouldExecute(order, market) {
			_, err := m.stopLossService.ExecuteOrder(order.ID())
			if err != nil {
				fmt.Printf("Error executing order %s: %v", order.ID(), err)
			}
		}

	}

	return nil
}

// Checks if the stop loss order should be executed
// If the order side is YES: check if the yes price is below the threshold
// If the order side is NO: check if the no price is below the threshold
func (m *StopLossOrderMonitor) shouldExecute(
	order *entities.StopLossOrder,
	market *kalshi.MarketResponse,
) bool {
	if order.Side() == entities.SideYes &&
		market.Market.YesPrice < order.Threshold().Value() {
		return true
	}

	if order.Side() == entities.SideNo &&
		market.Market.NoPrice < order.Threshold().Value() {
		return true
	}

	return false
}
