package trigger_domain

import (
	"fmt"
	"prediction-risk/internal/app/contract"
)

// OrderSide represents whether we're buying or selling
type OrderSide string

const (
	Buy  OrderSide = "BUY"
	Sell OrderSide = "SELL"
)

func (s OrderSide) String() string {
	return string(s)
}

func (s OrderSide) IsValid() bool {
	switch s {
	case Buy, Sell:
		return true
	default:
		return false
	}
}

func NewOrderSide(s string) (OrderSide, error) {
	switch s {
	case "BUY":
		return Buy, nil
	case "SELL":
		return Sell, nil
	default:
		return "", fmt.Errorf("invalid OrderSide: %s", s)
	}
}

// Action represents what to do when the condition is met
type TriggerAction struct {
	Contract   contract.ContractIdentifier
	Side       OrderSide
	Size       *uint                   // nil means "full position" for sells
	LimitPrice *contract.ContractPrice // nil means "market order"
}

func NewTriggerAction(
	contract contract.ContractIdentifier,
	side OrderSide,
	size *uint,
	limitPrice *contract.ContractPrice,
) (*TriggerAction, error) {
	if !side.IsValid() {
		return nil, fmt.Errorf("invalid side: %s", side)
	}

	// If side is Buy, size must not be nil
	if side == Buy && size == nil {
		return nil, fmt.Errorf("size must be provided for buy actions")
	}

	return &TriggerAction{
		Contract:   contract,
		Side:       side,
		Size:       size,
		LimitPrice: limitPrice,
	}, nil
}
