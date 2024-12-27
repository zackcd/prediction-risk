package mocks

import (
	"prediction-risk/internal/domain/entities"

	"github.com/stretchr/testify/mock"
)

type MockOrderExecutor struct {
	mock.Mock
}

func (m *MockOrderExecutor) ExecuteOrder(order entities.Order, isDryRun bool) error {
	args := m.Called(order, isDryRun)
	return args.Error(0)
}
