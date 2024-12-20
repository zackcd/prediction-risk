package mocks

import (
	"prediction-risk/internal/domain/entities"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockStopLossService struct {
   mock.Mock
}

func (m *MockStopLossService) GetOrder(stopLossOrderId uuid.UUID) (*entities.StopLossOrder, error) {
   args := m.Called(stopLossOrderId)
   if args.Get(0) == nil {
       return nil, args.Error(1)
   }
   return args.Get(0).(*entities.StopLossOrder), args.Error(1)
}

func (m *MockStopLossService) CreateOrder(
   ticker string,
   side entities.Side,
   threshold entities.ContractPrice,
) (*entities.StopLossOrder, error) {
   args := m.Called(ticker, side, threshold)
   if args.Get(0) == nil {
       return nil, args.Error(1)
   }
   return args.Get(0).(*entities.StopLossOrder), args.Error(1)
}

func (m *MockStopLossService) UpdateOrder(
   stopLossOrderId uuid.UUID,
   threshold entities.ContractPrice,
) (*entities.StopLossOrder, error) {
   args := m.Called(stopLossOrderId, threshold)
   if args.Get(0) == nil {
       return nil, args.Error(1)
   }
   return args.Get(0).(*entities.StopLossOrder), args.Error(1)
}

func (m *MockStopLossService) CancelOrder(stopLossOrderId uuid.UUID) (*entities.StopLossOrder, error) {
   args := m.Called(stopLossOrderId)
   if args.Get(0) == nil {
       return nil, args.Error(1)
   }
   return args.Get(0).(*entities.StopLossOrder), args.Error(1)
}

func (m *MockStopLossService) GetActiveOrders() ([]*entities.StopLossOrder, error) {
   args := m.Called()
   if args.Get(0) == nil {
       return nil, args.Error(1)
   }
   return args.Get(0).([]*entities.StopLossOrder), args.Error(1)
}

func (m *MockStopLossService) ExecuteOrder(stopLossOrderId uuid.UUID, isDryRun bool) (*entities.StopLossOrder, error) {
   args := m.Called(stopLossOrderId, isDryRun)
   if args.Get(0) == nil {
       return nil, args.Error(1)
   }
   return args.Get(0).(*entities.StopLossOrder), args.Error(1)
}

// Example usage:
/*
func TestSomething(t *testing.T) {
   mock := new(MockStopLossService)

   // Setup expectations
   orderID := uuid.New()
   expectedOrder := entities.NewStopLossOrder("TICKER", entities.SideYes, entities.NewContractPrice(50))
   mock.On("GetOrder", orderID).Return(expectedOrder, nil)

   // Call the method
   order, err := mock.GetOrder(orderID)

   // Assert expectations
   mock.AssertExpectations(t)
   assert.NoError(t, err)
   assert.Equal(t, expectedOrder, order)
}
*/
