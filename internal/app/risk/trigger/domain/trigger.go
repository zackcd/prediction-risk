package trigger_domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type TriggerID uuid.UUID

func NewTriggerID() TriggerID {
	return TriggerID(uuid.New())
}

func (t TriggerID) String() string {
	return uuid.UUID(t).String()
}

// OrderStatus represents the current state of an order
type TriggerStatus string

func (s TriggerStatus) String() string {
	return string(s)
}

// IsValid checks if the OrderStatus is one of the defined constants
func (s TriggerStatus) IsValid() bool {
	switch s {
	case StatusActive, StatusTriggered, StatusCancelled, StatusExpired:
		return true
	default:
		return false
	}
}

func NewTriggerStatus(s string) (TriggerStatus, error) {
	switch s {
	case "ACTIVE":
		return StatusActive, nil
	case "TRIGGERED":
		return StatusTriggered, nil
	case "CANCELLED":
		return StatusCancelled, nil
	case "EXPIRED":
		return StatusExpired, nil
	default:
		return "", fmt.Errorf("invalid TriggerStatus: %s", s)
	}
}

func (s TriggerStatus) IsTerminal() bool {
	switch s {
	case StatusTriggered, StatusCancelled, StatusExpired:
		return true
	default:
		return false
	}
}

// Active means the order is currently being monitored
// Executed means the order has been triggered
// Cancelled means the order has been cancelled
// Expired means the event has passed and the order is no longer valid
const (
	StatusActive    TriggerStatus = "ACTIVE"
	StatusTriggered TriggerStatus = "TRIGGERED"
	StatusCancelled TriggerStatus = "CANCELLED"
	StatusExpired   TriggerStatus = "EXPIRED"
)

// TriggerType represents the type of trigger
type TriggerType string

const (
	TriggerTypeStop TriggerType = "STOP"
)

func (t TriggerType) String() string {
	return string(t)
}

// IsValid checks if the OrderStatus is one of the defined constants
func (t TriggerType) IsValid() bool {
	switch t {
	case TriggerTypeStop:
		return true
	default:
		return false
	}
}

type Trigger struct {
	TriggerID   TriggerID
	TriggerType TriggerType
	Status      TriggerStatus
	Condition   TriggerCondition
	Actions     []TriggerAction
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func NewTrigger(
	triggerType TriggerType,
	condition TriggerCondition,
	actions []TriggerAction,
) *Trigger {
	currentTime := time.Now()
	return &Trigger{
		TriggerID:   NewTriggerID(),
		TriggerType: triggerType,
		Status:      StatusActive,
		Condition:   condition,
		Actions:     actions,
		CreatedAt:   currentTime,
		UpdatedAt:   currentTime,
	}
}
