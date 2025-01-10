package order

import (
	"fmt"
	"prediction-risk/internal/domain/contract"
	"time"

	"github.com/google/uuid"
)

type OrderID uuid.UUID

func NewOrderID() OrderID {
	return OrderID(uuid.New())
}

func (o OrderID) String() string {
	return uuid.UUID(o).String()
}

type OrderType string

const (
	OrderTypeStop OrderType = "STOP"
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

func (s OrderStatus) IsTerminal() bool {
	switch s {
	case OrderStatusTriggered, OrderStatusCancelled, OrderStatusExpired:
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
	ID() OrderID
	OrderType() OrderType
	Ticker() string
	Side() contract.Side
	Status() OrderStatus
	CreatedAt() time.Time
	UpdatedAt() time.Time
	UpdateStatus(OrderStatus) error
}

type order struct {
	orderId   OrderID
	orderType OrderType
	ticker    string
	side      contract.Side
	status    OrderStatus
	createdAt time.Time
	updatedAt time.Time
}

// newBaseOrder creates a new baseOrder with common initialization
func newOrder(
	orderType OrderType,
	ticker string,
	side contract.Side,
	orderId *OrderID,
) order {
	now := time.Now().UTC()
	id := OrderID(uuid.New())
	if orderId != nil {
		id = *orderId
	}
	return order{
		orderId:   id,
		orderType: orderType,
		ticker:    ticker,
		side:      side,
		status:    OrderStatusActive,
		createdAt: now,
		updatedAt: now,
	}
}

// Implement getters for the base order
func (o *order) ID() OrderID          { return o.orderId }
func (o *order) OrderType() OrderType { return o.orderType }
func (o *order) Ticker() string       { return o.ticker }
func (o *order) Side() contract.Side  { return o.side }
func (o *order) Status() OrderStatus  { return o.status }
func (o *order) CreatedAt() time.Time { return o.createdAt }
func (o *order) UpdatedAt() time.Time { return o.updatedAt }

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
