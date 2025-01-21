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

	if !threshold.IsValid() {
		return nil, fmt.Errorf("invalid threshold: %d", threshold)
	}

	return &PriceRule{
		Threshold: threshold,
		Direction: direction,
	}, nil
}

// Checks whether the price rule is satisfied and the trigger should be executed
func (p PriceRule) isSatisfied(price contract.ContractPrice) (bool, error) {
	// Validate the input price is within valid bounds
	if !price.IsValid() {
		return false, fmt.Errorf("invalid price: %d", price.Value())
	}

	switch p.Direction {
	case Above:
		return price >= p.Threshold, nil
	case Below:
		return price <= p.Threshold, nil
	default:
		// This should never happen if NewPriceRule validation is working,
		// but defensive programming is good practice
		return false, fmt.Errorf("invalid direction: %s", p.Direction)
	}
}

func (p PriceRule) String() string {
	return fmt.Sprintf("price rule %s %d", p.Direction, p.Threshold.Value())
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

// IsSatisfied checks if any condition is satisfied
func (c TriggerCondition) IsSatisfied(price contract.ContractPrice) (bool, error) {
	if c.Price != nil {
		return c.Price.isSatisfied(price)
	}

	return false, nil
}
