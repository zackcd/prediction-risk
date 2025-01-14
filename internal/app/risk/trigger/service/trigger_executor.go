package trigger_service

import (
	"fmt"
	trigger_domain "prediction-risk/internal/app/risk/trigger/domain"
)

type TriggerExecutor struct {
	triggerService   *TriggerService
	exchangeProvider ExchangeProvider
}

func NewTriggerExecutor(
	triggerService *TriggerService,
	exchangeProvider ExchangeProvider,
) *TriggerExecutor {
	return &TriggerExecutor{
		triggerService:   triggerService,
		exchangeProvider: exchangeProvider,
	}
}

func ExecuteTrigger(trigger *trigger_domain.Trigger) (*trigger_domain.Trigger, error) {
	// Check if the trigger is exeutable (it must be active)
	if trigger.Status != trigger_domain.StatusActive {
		return nil, fmt.Errorf("trigger is not active, status: %s", trigger.Status)
	}

	// Get the positions indicated in the trigger's action(s)

	// For each action, execute the action

	return nil, nil
}

func (t *TriggerExecutor) executeAction(action *trigger_domain.TriggerAction) error {
	// Get the positions for the contract in the action

	// Execute the action

	return nil
}
