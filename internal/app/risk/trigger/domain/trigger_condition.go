package trigger_domain

import (
	"fmt"
	"prediction-risk/internal/app/contract"
)

// Direction represents which way a price needs to move to trigger
type Direction string

const (
	Above Direction = "ABOVE"
	Below Direction = "BELOW"
)

func (d Direction) String() string {
	return string(d)
}

func (d Direction) IsValid() bool {
	switch d {
	case Above, Below:
		return true
	default:
		return false
	}
}

// PriceRule represents a price-based rule
type PriceRule struct {
	Threshold contract.ContractPrice
	Direction Direction
}

func newPriceRule(threshold contract.ContractPrice, direction Direction) (*PriceRule, error) {
	if !direction.IsValid() {
		return nil, fmt.Errorf("invalid direction: %s", direction)
	}

	return &PriceRule{
		Threshold: threshold,
		Direction: direction,
	}, nil
}

// Condition represents what we're watching for, with only one rule active at a time
type TriggerCondition struct {
	Contract contract.ContractIdentifier
	Price    *PriceRule // nil if not a price condition
}

// NewPriceCondition creates a price-based condition with validation
func NewPriceCondition(
	contract contract.ContractIdentifier,
	threshold contract.ContractPrice,
	direction Direction,
) (*TriggerCondition, error) {
	priceRule, err := newPriceRule(threshold, direction)
	if err != nil {
		return nil, err
	}

	return &TriggerCondition{
		Contract: contract,
		Price:    priceRule,
	}, nil
}
