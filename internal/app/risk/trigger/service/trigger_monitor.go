package trigger_service

import (
	"fmt"
	"log"
	"prediction-risk/internal/app/contract"
	exchange_service "prediction-risk/internal/app/exchange/service"
	trigger_domain "prediction-risk/internal/app/risk/trigger/domain"
	"time"

	"github.com/samber/lo"
)

type TriggerMonitor struct {
	triggerService  *TriggerService
	triggerExecutor *TriggerExecutor
	exchangeService exchange_service.ExchangeService
	interval        time.Duration
	done            chan struct{}
	isDryRun        bool
}

func NewTriggerMonitor(
	triggerService *TriggerService,
	triggerExecutor *TriggerExecutor,
	exchangeService exchange_service.ExchangeService,
	interval time.Duration,
	isDryRun bool,
) *TriggerMonitor {
	log.Printf("Initializing TriggerMonitor with interval: %v", interval)
	return &TriggerMonitor{
		triggerService:  triggerService,
		triggerExecutor: triggerExecutor,
		exchangeService: exchangeService,
		interval:        interval,
		done:            make(chan struct{}),
	}
}

func (m *TriggerMonitor) Start() {
	log.Printf("Starting TriggerMonitor")
	go func() {
		ticker := time.NewTicker(m.interval)
		defer ticker.Stop()

		for {
			select {
			case <-m.done:
				log.Println("TriggerMonitor stopped")
				return
			case <-ticker.C:
				log.Println("Running trigger check...")
				if err := m.checkTriggers(); err != nil {
					log.Printf("Error checking triggers: %v", err)
				}
			}
		}
	}()
}

func (m *TriggerMonitor) Stop() {
	log.Println("Stopping TriggerMonitor...")
	close(m.done)
}

func (m *TriggerMonitor) checkTriggers() error {
	triggers, err := m.triggerService.Get()
	if err != nil {
		return fmt.Errorf("getting orders: %w", err)
	}
	activeTriggers := lo.Filter(triggers, func(o *trigger_domain.Trigger, _ int) bool {
		return o.Status == trigger_domain.StatusActive
	})
	log.Printf("Found %d active stop triggers", len(activeTriggers))

	executedTriggers := make([]*trigger_domain.Trigger, 0)
	executionErrors := make([]error, 0)

	for _, trigger := range activeTriggers {
		log.Printf("Checking %s trigger %s...",
			trigger.TriggerType,
			trigger.TriggerID,
		)

		// Get current price of the contract
		market, err := m.exchangeService.GetMarket(trigger.Condition.Contract.Ticker)
		if err != nil {
			executionErrors = append(executionErrors, err)
			continue
		}

		// Get the current price of the contract
		var currentPrice contract.ContractPrice
		switch trigger.Condition.Contract.Side {
		case contract.SideYes:
			currentPrice = market.Pricing.YesSide.Ask
		case contract.SideNo:
			currentPrice = market.Pricing.NoSide.Ask
		default:
			executionErrors = append(executionErrors, fmt.Errorf("invalid contract side: %s", trigger.Condition.Contract.Side))
			continue
		}

		// Check if the trigger condition is met
		isSatisfed, err := trigger.Condition.IsSatisfied(currentPrice)
		if err != nil {
			executionErrors = append(executionErrors, err)
			continue
		}

		// If the trigger condition is met, execute the trigger
		if isSatisfed {
			executedTrigger, err := m.triggerExecutor.ExecuteTrigger(trigger)
			if err != nil {
				executionErrors = append(executionErrors, err)
				continue
			} else {
				executedTriggers = append(executedTriggers, executedTrigger)
			}
		}
	}

	// Log the executed triggers and errors
	log.Printf("Executed %d triggers", len(executedTriggers))
	for _, trigger := range executedTriggers {
		log.Printf("Executed trigger %s", trigger.TriggerID)
	}
	log.Printf("Encountered %d execution errors", len(executionErrors))
	for _, err := range executionErrors {
		log.Printf("Execution error: %v", err)
	}

	return nil
}

func (m *TriggerMonitor) processTrigger(trigger *trigger_domain.Trigger) (*trigger_domain.Trigger, error) {
	log.Printf("Checking %s trigger %s...",
		trigger.TriggerType,
		trigger.TriggerID,
	)

	// Get current price of the contract
	market, err := m.exchangeService.GetMarket(trigger.Condition.Contract.Ticker)
	if err != nil {
		return nil, err
	}

	// Get the current price of the contract
	var currentPrice contract.ContractPrice
	switch trigger.Condition.Contract.Side {
	case contract.SideYes:
		currentPrice = market.Pricing.YesSide.Ask
	case contract.SideNo:
		currentPrice = market.Pricing.NoSide.Ask
	default:
		return nil, fmt.Errorf("invalid contract side: %s", trigger.Condition.Contract.Side)
	}

	// Check if the trigger condition is met
	isSatisfed, err := trigger.Condition.IsSatisfied(currentPrice)
	if err != nil {
		return nil, err
	}

	// If the trigger condition is met, execute the trigger
	if isSatisfed {
		executedTrigger, err := m.triggerExecutor.ExecuteTrigger(trigger)
		if err != nil {
			return nil, err
		} else {
			return executedTrigger, nil
		}
	}

	return trigger, nil
}
