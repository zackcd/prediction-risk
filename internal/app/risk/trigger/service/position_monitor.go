package trigger_service

// import (
// 	"fmt"
// 	"log"
// 	"prediction-risk/internal/app/contract"
// 	"prediction-risk/internal/app/exchange/infrastructure/kalshi"
// 	exchange_service "prediction-risk/internal/app/exchange/service"
// 	trigger_domain "prediction-risk/internal/app/risk/trigger/domain"
// 	"time"
// )

// type PositionMonitor struct {
// 	exchangeService  exchange_service.ExchangeService
// 	stopOrderService StopOrderService
// 	interval         time.Duration
// 	done             chan struct{}
// }

// func NewPositionMonitor(
// 	exchangeService exchange_service.ExchangeService,
// 	stopOrderService StopOrderService,
// 	interval time.Duration,
// ) *PositionMonitor {
// 	log.Printf("Initializing PositionMonitor with interval: %v", interval)
// 	return &PositionMonitor{
// 		exchangeService:  exchangeService,
// 		stopOrderService: stopOrderService,
// 		interval:         interval,
// 		done:             make(chan struct{}),
// 	}
// }

// func (m *PositionMonitor) Start() {
// 	log.Println("Starting PositionMonitor")

// 	// Initial sync
// 	if err := m.syncPositions(); err != nil {
// 		log.Printf("Error during initial position sync: %v", err)
// 	}

// 	go func() {
// 		ticker := time.NewTicker(m.interval)
// 		defer ticker.Stop()

// 		for {
// 			select {
// 			case <-m.done:
// 				log.Println("PositionMonitor stopped")
// 				return
// 			case <-ticker.C:
// 				log.Println("Running position check...")
// 				if err := m.syncPositions(); err != nil {
// 					log.Printf("Error during position sync: %v", err)
// 				}
// 			}
// 		}
// 	}()
// }

// func (m *PositionMonitor) Stop() {
// 	log.Println("Stopping PositionMonitor...")
// 	close(m.done)
// }

// func (m *PositionMonitor) syncPositions() error {
// 	// Get open positions from exchange
// 	positions, err := m.exchangeService.GetPositions()
// 	if err != nil {
// 		return fmt.Errorf("getting positions: %w", err)
// 	}
// 	log.Printf("Found %d positions", len(positions.MarketPositions))

// 	// Get active stop orders
// 	activeOrders, err := m.stopOrderService.GetActiveOrders()
// 	if err != nil {
// 		return fmt.Errorf("getting active stop orders: %w", err)
// 	}
// 	log.Printf("Found %d active stop orders", len(activeOrders))

// 	// Create mapping of stop orders by ticker - buying Sides offset so we don't need to key by side
// 	ordersByTicker := make(map[string]*trigger_domain.StopOrder)
// 	for _, order := range activeOrders {
// 		ordersByTicker[order.Ticker()] = order
// 	}

// 	// Process each position
// 	for _, position := range positions.MarketPositions {
// 		log.Printf("Processing position %s", position.Ticker)

// 		if position.Position == 0 {
// 			// Position is closed, cancel any existing stop order
// 			if order, exists := ordersByTicker[position.Ticker]; exists {
// 				log.Printf("Cancelling stop order for closed position %s", position.Ticker)
// 				if _, err := m.stopOrderService.CancelOrder(order.ID()); err != nil {
// 					log.Printf("Error cancelling stop order for closed position %s: %v", position.Ticker, err)
// 				}
// 			}
// 			continue
// 		}

// 		// Check if there is a stop order for this position
// 		if _, exists := ordersByTicker[position.Ticker]; !exists {
// 			log.Printf("No stop order found for position %s, creating new stop order", position.Ticker)

// 			triggerPrice, err := m.calculateStopPrice(position)
// 			if err != nil {
// 				log.Printf("Error calculating stop price for %s: %v", position.Ticker, err)
// 				continue
// 			}

// 			// Determine side based on position
// 			var side contract.Side
// 			if position.Position > 0 {
// 				side = contract.SideYes
// 			} else {
// 				side = contract.SideNo
// 			}

// 			if _, err := m.stopOrderService.CreateOrder(
// 				position.Ticker,
// 				side,
// 				triggerPrice,
// 				nil,
// 			); err != nil {
// 				log.Printf("Error creating stop order for %s: %v", position.Ticker, err)

// 			} else {
// 				log.Printf("Successfully created stop order for %s", position.Ticker)
// 			}
// 		}
// 	}

// 	// Cancel stop orders for positions that are no longer open
// 	positionsByTicker := make(map[string]bool)
// 	for _, pos := range positions.MarketPositions {
// 		positionsByTicker[pos.Ticker] = true
// 	}

// 	for ticker, order := range ordersByTicker {
// 		if _, exists := positionsByTicker[ticker]; !exists {
// 			log.Printf("Cancelling orphaned stop order for %s", ticker)
// 			if _, err := m.stopOrderService.CancelOrder(order.ID()); err != nil {
// 				log.Printf("Error cancelling orphaned stop order for %s: %v", ticker, err)
// 			}
// 		}
// 	}

// 	return nil
// }

// func (m *PositionMonitor) calculateStopPrice(position kalshi.MarketPosition) (contract.ContractPrice, error) {
// 	market, err := m.exchangeService.GetMarket(position.Ticker)
// 	if err != nil {
// 		return contract.ContractPrice(0), fmt.Errorf("getting market data: %w", err)
// 	}

// 	// Example: Set stop loss 10% away from current price -- TODO: move to config
// 	stopPrice := market.LastPrice
// 	stopPrice = int(float64(market.LastPrice) * 0.9)
// 	return contract.NewContractPrice(stopPrice)
// }
