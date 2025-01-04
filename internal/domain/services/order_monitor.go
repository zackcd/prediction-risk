package services

import (
	"fmt"
	"log"
	"prediction-risk/internal/domain/entities"
	"prediction-risk/internal/infrastructure/external/kalshi"
	"time"
)

type OrderMonitor struct {
	stopOrderService StopOrderService
	exchange         ExchangeService
	interval         time.Duration
	done             chan struct{}
}

func NewOrderMonitor(
	stopOrderService StopOrderService,
	exchange ExchangeService,
	interval time.Duration,
) *OrderMonitor {
	log.Printf("Initializing OrderMonitor with interval: %v", interval)
	return &OrderMonitor{
		stopOrderService: stopOrderService,
		exchange:         exchange,
		interval:         interval,
		done:             make(chan struct{}),
	}
}

func (m *OrderMonitor) Start(isDryRun bool) {
	log.Printf("Starting OrderMonitor (dry run: %v)", isDryRun)
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
				if err := m.checkOrders(isDryRun); err != nil {
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

func (m *OrderMonitor) checkOrders(isDryRun bool) error {
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

		if order.Side() == entities.SideYes {
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
			log.Printf("Executing stop order %s (dry run: %v)...", order.ID(), isDryRun)
			_, err := m.stopOrderService.ExecuteOrder(order.ID(), isDryRun)
			if err != nil {
				log.Printf("ERROR executing order %s: %v", order.ID(), err)
			} else {
				log.Printf("Successfully executed order %s", order.ID())
			}
		}
	}

	return nil
}

func (m *OrderMonitor) shouldExecute(order *entities.StopOrder, market *kalshi.Market) bool {
	var bid int
	if order.Side() == entities.SideYes {
		bid = market.YesBid
	} else if order.Side() == entities.SideNo {
		bid = market.NoBid
	}

	if bid < order.TriggerPrice().Value() {
		log.Printf("%s stop triggered - bid (%d) below threshold (%d)",
			order.Side(), bid, order.TriggerPrice().Value())
		return true
	}
	return false
}
