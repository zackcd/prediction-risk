package trigger_domain

import (
	"prediction-risk/internal/app/contract"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDirection(t *testing.T) {
	tests := []struct {
		name      string
		direction Direction
		isValid   bool
	}{
		{"above direction", Above, true},
		{"below direction", Below, true},
		{"invalid direction", "INVALID", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.isValid, tt.direction.IsValid())
			assert.Equal(t, string(tt.direction), tt.direction.String())
		})
	}
}

func TestNewPriceCondition(t *testing.T) {
	contractID := contract.ContractIdentifier{
		Ticker: contract.Ticker("FOO"),
		Side:   contract.SideYes,
	}
	tests := []struct {
		name        string
		contract    contract.ContractIdentifier
		threshold   contract.ContractPrice
		direction   Direction
		expectError bool
	}{
		{
			name:        "valid above condition",
			contract:    contractID,
			threshold:   50,
			direction:   Above,
			expectError: false,
		},
		{
			name:        "valid below condition",
			contract:    contractID,
			threshold:   50,
			direction:   Below,
			expectError: false,
		},
		{
			name:        "invalid direction",
			contract:    contractID,
			threshold:   50,
			direction:   "INVALID",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			condition, err := NewPriceCondition(tt.contract, tt.threshold, tt.direction)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, condition)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, condition)
				assert.Equal(t, tt.contract, condition.Contract)
				assert.NotNil(t, condition.Price)
				assert.Equal(t, tt.threshold, condition.Price.Threshold)
				assert.Equal(t, tt.direction, condition.Price.Direction)
			}
		})
	}
}

func TestPriceRule_IsSatisfied(t *testing.T) {
	tests := []struct {
		name      string
		rule      PriceRule
		price     contract.ContractPrice
		want      bool
		wantError bool
	}{
		{
			name:      "Above: Price higher than threshold",
			rule:      PriceRule{Threshold: contract.ContractPrice(50), Direction: Above},
			price:     contract.ContractPrice(100),
			want:      true,
			wantError: false,
		},
		{
			name:      "Above: Price equal to threshold",
			rule:      PriceRule{Threshold: contract.ContractPrice(100), Direction: Above},
			price:     contract.ContractPrice(100),
			want:      true,
			wantError: false,
		},
		{
			name:      "Above: Price lower than threshold",
			rule:      PriceRule{Threshold: contract.ContractPrice(100), Direction: Above},
			price:     contract.ContractPrice(50),
			want:      false,
			wantError: false,
		},
		{
			name:      "Below: Price lower than threshold",
			rule:      PriceRule{Threshold: contract.ContractPrice(100), Direction: Below},
			price:     contract.ContractPrice(50),
			want:      true,
			wantError: false,
		},
		{
			name:      "Below: Price equal to threshold",
			rule:      PriceRule{Threshold: contract.ContractPrice(100), Direction: Below},
			price:     contract.ContractPrice(100),
			want:      true,
			wantError: false,
		},
		{
			name:      "Below: Price higher than threshold",
			rule:      PriceRule{Threshold: contract.ContractPrice(50), Direction: Below},
			price:     contract.ContractPrice(100),
			want:      false,
			wantError: false,
		},
		{
			name:      "Invalid direction",
			rule:      PriceRule{Threshold: contract.ContractPrice(100), Direction: Direction("Invalid")},
			price:     contract.ContractPrice(100),
			want:      false,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.rule.isSatisfied(tt.price)
			if (err != nil) != tt.wantError {
				t.Errorf("PriceRule.isSatisfied() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if got != tt.want {
				t.Errorf("PriceRule.isSatisfied() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTriggerCondition_IsSatisfied(t *testing.T) {
	validPrice := contract.ContractPrice(100)
	aboveRule := &PriceRule{Threshold: validPrice, Direction: Above}

	tests := []struct {
		name      string
		condition TriggerCondition
		price     contract.ContractPrice
		want      bool
		wantError bool
	}{
		{
			name: "With price rule",
			condition: TriggerCondition{
				Contract: contract.ContractIdentifier{
					Ticker: contract.Ticker("FOO"),
					Side:   contract.SideYes,
				},
				Price: aboveRule,
			},
			price:     validPrice,
			want:      true, // We're not testing the rule itself, just that it's called
			wantError: false,
		},
		{
			name: "Without price rule",
			condition: TriggerCondition{
				Contract: contract.ContractIdentifier{
					Ticker: contract.Ticker("FOO"),
					Side:   contract.SideYes,
				},
				Price: nil,
			},
			price:     validPrice,
			want:      false,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.condition.IsSatisfied(tt.price)
			if (err != nil) != tt.wantError {
				t.Errorf("TriggerCondition.IsSatisfied() error = %v, wantError %v", err, tt.wantError)
				return
			}
			if got != tt.want {
				t.Errorf("TriggerCondition.IsSatisfied() = %v, want %v", got, tt.want)
			}
		})
	}
}
