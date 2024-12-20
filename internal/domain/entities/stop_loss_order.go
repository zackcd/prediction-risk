package entities

import (
	"time"

	"github.com/google/uuid"
)

type StopLossOrder struct {
	id        uuid.UUID
	ticker    string
	side      Side
	threshold ContractPrice // 0-100 - in cents
	status    StopLossOrderStatus
	createdAt time.Time
	updatedAt time.Time
}

func NewStopLossOrder(
	ticker string,
	side Side,
	threshold ContractPrice,
) *StopLossOrder {
	currentTime := time.Now().UTC()

	return &StopLossOrder{
		id:        uuid.New(),
		ticker:    ticker,
		side:      side,
		threshold: threshold,
		status:    SLOStatusActive,
		createdAt: currentTime,
		updatedAt: currentTime,
	}
}

func (o *StopLossOrder) ID() uuid.UUID               { return o.id }
func (o *StopLossOrder) Ticker() string              { return o.ticker }
func (o *StopLossOrder) Side() Side                  { return o.side }
func (o *StopLossOrder) Threshold() ContractPrice    { return o.threshold }
func (o *StopLossOrder) Status() StopLossOrderStatus { return o.status }
func (o *StopLossOrder) CreatedAt() time.Time        { return o.createdAt }
func (o *StopLossOrder) UpdatedAt() time.Time        { return o.updatedAt }

// If you need to update the threshold
func (o *StopLossOrder) SetThreshold(threshold ContractPrice) {
	o.threshold = threshold
	o.updateTimestamp()
}

// If you need to update the status
func (o *StopLossOrder) SetStatus(status StopLossOrderStatus) {
	o.status = status
	o.updateTimestamp()
}

// internal helper for timestamp updates
func (o *StopLossOrder) updateTimestamp() {
	o.updatedAt = time.Now().UTC()
}

type StopLossOrderStatus string

// Active means the order is currently being monitored
// Executed means the order has been triggered
// Cancelled means the order has been cancelled
// Expired means the event has passed and the order is no longer valid
const (
	SLOStatusActive    StopLossOrderStatus = "ACTIVE"
	SLOStatusExecuted  StopLossOrderStatus = "EXECUTED"
	SLOStatusCancelled StopLossOrderStatus = "CANCELLED"
	SLOStatusExpired   StopLossOrderStatus = "EXPIRED"
)
