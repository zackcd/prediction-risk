package trigger_domain

import (
	"errors"
	"fmt"
	"prediction-risk/internal/app/contract"
)

// Ensure condition and action are for the same contract
func NewStopTrigger(
	contract contract.ContractIdentifier,
	triggerPrice contract.ContractPrice,
	limitPrice *contract.ContractPrice,
) (*Trigger, error) {
	condition, err := NewPriceCondition(contract, triggerPrice, Below)
	if err != nil {
		return nil, err
	}

	action, err := NewTriggerAction(contract, Sell, nil, limitPrice)
	if err != nil {
		return nil, err
	}

	trigger := NewTrigger(TriggerTypeStop, *condition, []TriggerAction{*action})

	if err := ValidateStopTrigger(trigger); err != nil {
		return nil, err
	}

	return trigger, nil
}

func ValidateStopTrigger(t *Trigger) error {
	// Basic trigger validation
	if t == nil {
		return errors.New("trigger cannot be nil")
	}
	if t.TriggerType != TriggerTypeStop {
		return fmt.Errorf("invalid trigger type: expected %s, got %s", TriggerTypeStop, t.TriggerType)
	}

	// Condition validation
	if t.Condition.Price == nil {
		return errors.New("stop trigger must have a price condition")
	}
	// Price validation
	if !t.Condition.Price.Threshold.IsValid() {
		return fmt.Errorf("invalid stop price: %v", t.Condition.Price.Threshold)
	}
	if t.Condition.Price.Direction != Below {
		return fmt.Errorf("stop trigger price direction must be Below, got %s", t.Condition.Price.Direction)
	}

	// Actions validation
	if len(t.Actions) != 1 {
		return fmt.Errorf("stop trigger must have exactly one action, got %d", len(t.Actions))
	}

	action := t.Actions[0]
	if action.Side != Sell {
		return fmt.Errorf("stop trigger action must be Sell, got %s", action.Side)
	}

	// Contract consistency validation
	if t.Condition.Contract != action.Contract {
		return fmt.Errorf("condition contract (%v) must match action contract (%v)",
			t.Condition.Contract, action.Contract)
	}

	// If there's a limit price, it must be valid
	if action.LimitPrice != nil {
		// Limit price must be valid
		if !action.LimitPrice.IsValid() {
			return fmt.Errorf("invalid limit price: %v", *action.LimitPrice)
		}
		// Optional: Validate limit price is "reasonable" compared to stop price
		// e.g., not too far below stop price to prevent extreme slippage
		if float64(*action.LimitPrice) < float64(t.Condition.Price.Threshold)*0.9 { // 10% max slippage
			return fmt.Errorf("limit price (%v) too low compared to stop price (%v)",
				*action.LimitPrice, t.Condition.Price.Threshold)
		}
	}

	// Status validation
	if !t.Status.IsValid() {
		return fmt.Errorf("invalid trigger status: %s", t.Status)
	}

	return nil
}
