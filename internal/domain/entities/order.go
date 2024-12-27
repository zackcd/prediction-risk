package entities

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type OrderType string

func (t OrderType) String() string {
	return string(t)
}

const (
	OrderTypeStopLoss   OrderType = "STOP_LOSS"
	OrderTypeTakeProfit OrderType = "TAKE_PROFIT"
)

// OrderStatus represents the current state of an order
type OrderStatus string

func (s OrderStatus) String() string {
	return string(s)
}

// IsValid checks if the OrderStatus is one of the defined constants
func (s OrderStatus) IsValid() bool {
	switch s {
	case OrderStatusActive, OrderStatusTriggered, OrderStatusCancelled, OrderStatusExpired:
		return true
	default:
		return false
	}
}

// ParseOrderStatus creates an OrderStatus from a string, returning an error if invalid
func ParseOrderStatus(s string) (OrderStatus, error) {
	status := OrderStatus(s)
	if !status.IsValid() {
		return "", fmt.Errorf("invalid order status: %q", s)
	}
	return status, nil
}

// Active means the order is currently being monitored
// Executed means the order has been triggered
// Cancelled means the order has been cancelled
// Expired means the event has passed and the order is no longer valid
const (
	OrderStatusActive    OrderStatus = "ACTIVE"
	OrderStatusTriggered OrderStatus = "TRIGGERED"
	OrderStatusCancelled OrderStatus = "CANCELLED"
	OrderStatusExpired   OrderStatus = "EXPIRED"
)

type Order interface {
	ID() uuid.UUID
	Type() OrderType
	Ticker() string
	Side() Side
	TriggerPrice() ContractPrice
	Status() OrderStatus
	CreatedAt() time.Time
	UpdatedAt() time.Time
	UpdateTriggerPrice(ContractPrice)
	UpdateStatus(OrderStatus) error
}

type order struct {
	id           uuid.UUID
	orderType    OrderType
	ticker       string
	side         Side
	triggerPrice ContractPrice
	status       OrderStatus
	createdAt    time.Time
	updatedAt    time.Time
}

// newBaseOrder creates a new baseOrder with common initialization
func newOrder(
	orderType OrderType,
	ticker string,
	side Side,
	triggerPrice ContractPrice,
) order {
	now := time.Now().UTC()
	return order{
		id:           uuid.New(),
		orderType:    orderType,
		ticker:       ticker,
		side:         side,
		triggerPrice: triggerPrice,
		status:       OrderStatusActive,
		createdAt:    now,
		updatedAt:    now,
	}
}

// Implement getters for the base order
func (o *order) ID() uuid.UUID               { return o.id }
func (o *order) Type() OrderType             { return o.orderType }
func (o *order) Ticker() string              { return o.ticker }
func (o *order) Side() Side                  { return o.side }
func (o *order) TriggerPrice() ContractPrice { return o.triggerPrice }
func (o *order) Status() OrderStatus         { return o.status }
func (o *order) CreatedAt() time.Time        { return o.createdAt }
func (o *order) UpdatedAt() time.Time        { return o.updatedAt }

// UpdateTriggerPrice updates the trigger price and timestamp
func (o *order) UpdateTriggerPrice(triggerPrice ContractPrice) {
	o.triggerPrice = triggerPrice
	o.updateTimestamp()
}

func (o *order) UpdateStatus(status OrderStatus) error {
	if !status.IsValid() {
		return fmt.Errorf("invalid status: %q", status)
	}
	o.status = status
	o.updateTimestamp()
	return nil
}

func (o *order) updateTimestamp() {
	o.updatedAt = time.Now().UTC()
}
