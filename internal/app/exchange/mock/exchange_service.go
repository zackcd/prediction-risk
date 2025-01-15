package exchange_mock

import (
	"github.com/stretchr/testify/mock"
)

// MockExchangeService is a mock implementation of the ExchangeService interface
type MockExchangeService struct {
	mock.Mock
}
