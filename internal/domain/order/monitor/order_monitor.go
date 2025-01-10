package monitor

import (
	"fmt"
	"log"
	"prediction-risk/internal/domain/contract"
	"prediction-risk/internal/domain/exchange"
	"prediction-risk/internal/domain/order"
	"prediction-risk/internal/infrastructure/external/kalshi"
	"time"
)

type OrderMonitor struct {
	stopOrderService order.StopOrderService
	exchange         exchange.ExchangeService
	interval         time.Duration
	done             chan struct{}
	isDryRun         bool
}

func NewOrderMonitor(
	stopOrderService order.StopOrderService,
	exchange exchange.ExchangeService,
	interval time.Duration,
	isDryRun bool,
) *OrderMonitor {
	log.Printf("Initializing OrderMonitor with interval: %v", interval)
	return &OrderMonitor{
		stopOrderService: stopOrderService,
		exchange:         exchange,
		interval:         interval,
		done:             make(chan struct{}),
	}
}

func (m *OrderMonitor) Start() {
	log.Printf("Starting OrderMonitor")
	go func() {
		ticker := time.NewTicker(m.interval)
		defer ticker.Stop()

		for {
			select {
			case <-m.done:
				log.Println("OrderMonitor stopped")
				return
			case <-ticker.C:
				log.Println("Running order check...")
				if err := m.checkOrders(); err != nil {
					log.Printf("Error checking orders: %v", err)
				}
			}
		}
	}()
}

func (m *OrderMonitor) Stop() {
	log.Println("Stopping OrderMonitor...")
	close(m.done)
}

func (m *OrderMonitor) checkOrders() error {
	activeOrders, err := m.stopOrderService.GetActiveOrders()
	if err != nil {
		return fmt.Errorf("getting active orders: %w", err)
	}
	log.Printf("Found %d active stop orders", len(activeOrders))

	for _, order := range activeOrders {
		log.Printf("Checking order %s (ticker: %s, side: %s, threshold: %d)...",
			order.ID(),
			order.Ticker(),
			order.Side(),
			order.TriggerPrice().Value(),
		)

		// Get Market from KalshiClient
		market, err := m.exchange.GetMarket(order.Ticker())
		if err != nil {
			log.Printf("ERROR getting market data for %s: %v", order.Ticker(), err)
			continue // Skip this order if we can't get the price, but keep checking others
		}

		shouldExecute := m.shouldExecute(order, market)

		if order.Side() == contract.SideYes {
			log.Printf("Market %s YES bid: %d (threshold: %d)",
				order.Ticker(),
				market.YesBid,
				order.TriggerPrice().Value(),
			)
		} else {
			log.Printf("Market %s NO bid: %d (threshold: %d)",
				order.Ticker(),
				market.NoBid,
				order.TriggerPrice().Value(),
			)
		}

		log.Printf("Order %s should execute: %v", order.ID(), shouldExecute)

		if shouldExecute {
			log.Printf("Executing stop order %s (dry run: %v)...", order.ID(), m.isDryRun)
			_, err := m.stopOrderService.ExecuteOrder(order.ID(), m.isDryRun)
			if err != nil {
				log.Printf("ERROR executing order %s: %v", order.ID(), err)
			} else {
				log.Printf("Successfully executed order %s", order.ID())
			}
		}
	}

	return nil
}

func (m *OrderMonitor) shouldExecute(order *order.StopOrder, market *kalshi.Market) bool {
	var bid int
	if order.Side() == contract.SideYes {
		bid = market.YesBid
	} else if order.Side() == contract.SideNo {
		bid = market.NoBid
	}

	if bid < order.TriggerPrice().Value() {
		log.Printf("%s stop triggered - bid (%d) below threshold (%d)",
			order.Side(), bid, order.TriggerPrice().Value())
		return true
	}
	return false
}
