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
			threshold:   50000.0,
			direction:   Above,
			expectError: false,
		},
		{
			name:        "valid below condition",
			contract:    contractID,
			threshold:   50000.0,
			direction:   Below,
			expectError: false,
		},
		{
			name:        "invalid direction",
			contract:    contractID,
			threshold:   50000.0,
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
