package trigger_service

import (
	"fmt"
	"log"
	trigger_domain "prediction-risk/internal/app/risk/trigger/domain"
	"time"

	"github.com/samber/lo"
)

type TriggerMonitor struct {
	triggerService *TriggerService
	interval       time.Duration
	done           chan struct{}
	isDryRun       bool
}

func NewTriggerMonitor(
	triggerService *TriggerService,
	interval time.Duration,
	isDryRun bool,
) *TriggerMonitor {
	log.Printf("Initializing TriggerMonitor with interval: %v", interval)
	return &TriggerMonitor{
		triggerService: triggerService,
		interval:       interval,
		done:           make(chan struct{}),
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
				if err := m.checkOrders(); err != nil {
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

func (m *TriggerMonitor) checkOrders() error {
	triggers, err := m.triggerService.Get()
	if err != nil {
		return fmt.Errorf("getting orders: %w", err)
	}
	activeTriggers := lo.Filter(triggers, func(o *trigger_domain.Trigger, _ int) bool {
		return o.Status == trigger_domain.StatusActive
	})
	log.Printf("Found %d active stop triggers", len(activeTriggers))

	for _, trigger := range activeTriggers {
		log.Printf("Checking %s trigger %s...",
			trigger.TriggerType,
			trigger.TriggerID,
		)
	}

	return nil
}
