package entities

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type TakeProfitOrder struct {
	id        uuid.UUID
	ticker    string
	side      Side
	threshold ContractPrice
	status    TakeProfitOrderStatus
	createdAt time.Time
	updatedAt time.Time
}

func NewTakeProfitOrder(ticker string, side Side, threshold ContractPrice) *TakeProfitOrder {
	currentTime := time.Now().UTC()
	return &TakeProfitOrder{
		id:        uuid.New(),
		ticker:    ticker,
		side:      side,
		threshold: threshold,
		status:    TPOStatusActive,
		createdAt: currentTime,
		updatedAt: currentTime,
	}
}

func (o *TakeProfitOrder) ID() uuid.UUID                 { return o.id }
func (o *TakeProfitOrder) Ticker() string                { return o.ticker }
func (o *TakeProfitOrder) Side() Side                    { return o.side }
func (o *TakeProfitOrder) Threshold() ContractPrice      { return o.threshold }
func (o *TakeProfitOrder) Status() TakeProfitOrderStatus { return o.status }
func (o *TakeProfitOrder) CreatedAt() time.Time          { return o.createdAt }
func (o *TakeProfitOrder) UpdatedAt() time.Time          { return o.updatedAt }

func (o *TakeProfitOrder) SetThreshold(threshold ContractPrice) {
	o.threshold = threshold
	o.updatedAt = time.Now().UTC()
}

func (o *TakeProfitOrder) SetStatus(status TakeProfitOrderStatus) {
	o.status = status
	o.updatedAt = time.Now().UTC()
}

type TakeProfitOrderStatus interface{ String() string }
type takeProfitOrderStatus string

func (s takeProfitOrderStatus) String() string {
	return s.String()
}

func NewTakeProfitOrderStatus(value string) (TakeProfitOrderStatus, error) {
	switch value {
	case string(TPOStatusActive), string(TPOStatusExecuted):
		return takeProfitOrderStatus(value), nil
	default:
		return nil, fmt.Errorf("invalid take profit order status value: %q", value)
	}
}

const (
	TPOStatusActive    takeProfitOrderStatus = "ACTIVE"
	TPOStatusExecuted  takeProfitOrderStatus = "EXECUTED"
	TPOStatusCancelled takeProfitOrderStatus = "CANCELLED"
)
