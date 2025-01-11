package trigger_domain

import (
	"prediction-risk/internal/app/contract"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewStopTrigger(t *testing.T) {
	contractID := contract.ContractIdentifier{
		Ticker: contract.Ticker("FOO"),
		Side:   contract.SideYes,
	}
	limitPrice := contract.ContractPrice(49)
	badLimitPrice := contract.ContractPrice(40)
	tests := []struct {
		name        string
		contract    contract.ContractIdentifier
		stopPrice   contract.ContractPrice
		limitPrice  *contract.ContractPrice
		expectError bool
	}{
		{
			name:        "valid stop trigger",
			contract:    contractID,
			stopPrice:   contract.ContractPrice(50),
			limitPrice:  &limitPrice,
			expectError: false,
		},
		{
			name:        "valid stop market trigger",
			contract:    contractID,
			stopPrice:   contract.ContractPrice(50),
			limitPrice:  nil,
			expectError: false,
		},
		{
			name:        "invalid stop price",
			contract:    contractID,
			stopPrice:   contract.ContractPrice(-10),
			limitPrice:  &limitPrice,
			expectError: true,
		},
		{
			name:        "limit price too low",
			contract:    contractID,
			stopPrice:   contract.ContractPrice(50),
			limitPrice:  &badLimitPrice, // More than 10% below stop price
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trigger, err := NewStopTrigger(tt.contract, tt.stopPrice, tt.limitPrice)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, trigger)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, trigger)
				assert.Equal(t, TriggerTypeStop, trigger.TriggerType)
				assert.Equal(t, tt.contract, trigger.Condition.Contract)
				assert.Equal(t, Below, trigger.Condition.Price.Direction)
				assert.Equal(t, tt.stopPrice, trigger.Condition.Price.Threshold)
				assert.Equal(t, tt.limitPrice, trigger.Actions[0].LimitPrice)
			}
		})
	}
}

func TestValidateStopTrigger(t *testing.T) {
	validContract := contract.ContractIdentifier{
		Ticker: contract.Ticker("FOO"),
		Side:   contract.SideYes,
	}
	validStopPrice := contract.ContractPrice(50)
	validLimitPrice := contract.ContractPrice(49)

	tests := []struct {
		name         string
		setupTrigger func() *Trigger
		expectError  bool
		errorMessage string
	}{
		{
			name: "valid trigger",
			setupTrigger: func() *Trigger {
				trigger, _ := NewStopTrigger(validContract, validStopPrice, &validLimitPrice)
				return trigger
			},
			expectError: false,
		},
		{
			name: "nil trigger",
			setupTrigger: func() *Trigger {
				return nil
			},
			expectError:  true,
			errorMessage: "trigger cannot be nil",
		},
		{
			name: "wrong trigger type",
			setupTrigger: func() *Trigger {
				trigger, _ := NewStopTrigger(validContract, validStopPrice, &validLimitPrice)
				trigger.TriggerType = "INVALID"
				return trigger
			},
			expectError:  true,
			errorMessage: "invalid trigger type",
		},
		{
			name: "mismatched contracts",
			setupTrigger: func() *Trigger {
				trigger, _ := NewStopTrigger(validContract, validStopPrice, &validLimitPrice)
				trigger.Actions[0].Contract =
					contract.ContractIdentifier{
						Ticker: contract.Ticker("BAR"),
						Side:   contract.SideYes,
					}
				return trigger
			},
			expectError:  true,
			errorMessage: "condition contract",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			trigger := tt.setupTrigger()
			err := ValidateStopTrigger(trigger)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMessage != "" {
					assert.Contains(t, err.Error(), tt.errorMessage)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
