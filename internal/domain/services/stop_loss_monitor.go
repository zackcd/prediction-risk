package services

import (
	"fmt"
	"log"
	"prediction-risk/internal/domain/entities"
	"prediction-risk/internal/infrastructure/external/kalshi"
	"time"
)

type StopLossMonitor struct {
	stopLossService StopLossService
	exchange        ExchangeService
	interval        time.Duration
	done            chan struct{}
}

func NewStopLossMonitor(
	stopLossService StopLossService,
	exchange ExchangeService,
	interval time.Duration,
) *StopLossMonitor {
	log.Printf("Initializing StopLossMonitor with interval: %v", interval)
	return &StopLossMonitor{
		stopLossService: stopLossService,
		exchange:        exchange,
		interval:        interval,
		done:            make(chan struct{}),
	}
}

func (m *StopLossMonitor) Start(isDryRun bool) {
	log.Printf("Starting StopLossMonitor (dry run: %v)", isDryRun)
	go func() {
		ticker := time.NewTicker(m.interval)
		defer ticker.Stop()

		for {
			select {
			case <-m.done:
				log.Println("StopLossMonitor stopping")
				return
			case <-ticker.C:
				log.Println("Running stop loss check...")
				if err := m.checkOrders(isDryRun); err != nil {
					log.Printf("ERROR checking orders: %v", err)
				}
			}
		}
	}()
}

func (m *StopLossMonitor) Stop() {
	log.Println("Stopping StopLossMonitor...")
	close(m.done)
}

func (m *StopLossMonitor) checkOrders(isDryRun bool) error {
	activeOrders, err := m.stopLossService.GetActiveOrders()
	if err != nil {
		return fmt.Errorf("getting active orders: %w", err)
	}
	log.Printf("Found %d active stop loss orders", len(activeOrders))

	for _, order := range activeOrders {
		log.Printf("Checking order %s (ticker: %s, side: %s, threshold: %d)...",
			order.ID(),
			order.Ticker(),
			order.Side(),
			order.Threshold().Value(),
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
				order.Threshold().Value(),
			)
		} else {
			log.Printf("Market %s NO bid: %d (threshold: %d)",
				order.Ticker(),
				market.NoBid,
				order.Threshold().Value(),
			)
		}

		log.Printf("Order %s should execute: %v", order.ID(), shouldExecute)

		if shouldExecute {
			log.Printf("Executing stop loss order %s (dry run: %v)...", order.ID(), isDryRun)
			_, err := m.stopLossService.ExecuteOrder(order.ID(), isDryRun)
			if err != nil {
				log.Printf("ERROR executing order %s: %v", order.ID(), err)
			} else {
				log.Printf("Successfully executed order %s", order.ID())
			}
		}
	}

	return nil
}

func (m *StopLossMonitor) shouldExecute(order *entities.StopLossOrder, market *kalshi.Market) bool {
	var bid int
	if order.Side() == entities.SideYes {
		bid = market.YesBid
	} else if order.Side() == entities.SideNo {
		bid = market.NoBid
	}

	if bid < order.Threshold().Value() {
		log.Printf("%s stop loss triggered - bid (%d) below threshold (%d)",
			order.Side(), bid, order.Threshold().Value())
		return true
	}
	return false
}
