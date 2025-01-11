package trigger_domain

import (
	"prediction-risk/internal/app/contract"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrderSide(t *testing.T) {
	tests := []struct {
		name    string
		side    OrderSide
		isValid bool
	}{
		{"buy side", Buy, true},
		{"sell side", Sell, true},
		{"invalid side", "INVALID", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.isValid, tt.side.IsValid())
			assert.Equal(t, string(tt.side), tt.side.String())
		})
	}
}

func TestNewTriggerAction(t *testing.T) {
	contractID := contract.ContractIdentifier{
		Ticker: contract.Ticker("FOO"),
		Side:   contract.SideYes,
	}
	price := contract.ContractPrice(50)
	size := uint(1)
	tests := []struct {
		name        string
		contract    contract.ContractIdentifier
		side        OrderSide
		size        *uint
		limitPrice  *contract.ContractPrice
		expectError bool
	}{
		{
			name:        "valid sell market order",
			contract:    contractID,
			side:        Sell,
			size:        nil,
			limitPrice:  nil,
			expectError: false,
		},
		{
			name:        "valid sell limit order",
			contract:    contractID,
			side:        Sell,
			size:        nil,
			limitPrice:  &price,
			expectError: false,
		},
		{
			name:        "valid buy order with size",
			contract:    contractID,
			side:        Buy,
			size:        &size,
			limitPrice:  nil,
			expectError: false,
		},
		{
			name:        "invalid buy order without size",
			contract:    contractID,
			side:        Buy,
			size:        nil,
			limitPrice:  nil,
			expectError: true,
		},
		{
			name:        "invalid side",
			contract:    contractID,
			side:        "INVALID",
			size:        nil,
			limitPrice:  nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			action, err := NewTriggerAction(tt.contract, tt.side, tt.size, tt.limitPrice)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, action)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, action)
				assert.Equal(t, tt.contract, action.Contract)
				assert.Equal(t, tt.side, action.Side)
				assert.Equal(t, tt.size, action.Size)
				assert.Equal(t, tt.limitPrice, action.LimitPrice)
			}
		})
	}
}
