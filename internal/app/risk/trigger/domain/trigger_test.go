package trigger_domain

import (
	"prediction-risk/internal/app/contract"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestTriggerID(t *testing.T) {
	t.Run("new trigger ID should be valid UUID", func(t *testing.T) {
		id := NewTriggerID()
		_, err := uuid.Parse(id.String())
		assert.NoError(t, err)
	})

	t.Run("string representation should be valid", func(t *testing.T) {
		originalUUID := uuid.New()
		id := TriggerID(originalUUID)
		assert.Equal(t, originalUUID.String(), id.String())
	})
}

func TestTriggerStatus(t *testing.T) {
	tests := []struct {
		name     string
		status   TriggerStatus
		isValid  bool
		terminal bool
	}{
		{"active status", StatusActive, true, false},
		{"triggered status", StatusTriggered, true, true},
		{"cancelled status", StatusCancelled, true, true},
		{"expired status", StatusExpired, true, true},
		{"invalid status", "INVALID", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.isValid, tt.status.IsValid())
			assert.Equal(t, tt.terminal, tt.status.IsTerminal())
			assert.Equal(t, string(tt.status), tt.status.String())
		})
	}
}

func TestNewTrigger(t *testing.T) {
	contractID := contract.ContractIdentifier{
		Ticker: contract.Ticker("FOO"),
		Side:   contract.SideYes,
	}
	condition := TriggerCondition{
		Contract: contractID,
		Price: &PriceRule{
			Threshold: 50000.0,
			Direction: Below,
		},
	}
	actions := []TriggerAction{{
		Contract: contractID,
		Side:     Sell,
	}}

	t.Run("successful creation", func(t *testing.T) {
		trigger := NewTrigger(TriggerTypeStop, condition, actions)

		assert.NotNil(t, trigger)
		assert.NotEqual(t, uuid.Nil, uuid.UUID(trigger.TriggerID))
		assert.Equal(t, TriggerTypeStop, trigger.TriggerType)
		assert.Equal(t, StatusActive, trigger.Status)
		assert.Equal(t, condition, trigger.Condition)
		assert.Equal(t, actions, trigger.Actions)
		assert.False(t, trigger.CreatedAt.IsZero())
		assert.False(t, trigger.UpdatedAt.IsZero())
	})
}
