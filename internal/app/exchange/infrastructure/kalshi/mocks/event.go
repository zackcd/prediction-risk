package kalshi_mocks

import (
	"prediction-risk/internal/app/exchange/infrastructure/kalshi"

	"github.com/stretchr/testify/mock"
)

type MockEventService struct {
	mock.Mock
}

func (m *MockMarketService) GetEvent(eventTicker string) (*kalshi.EventResponse, error) {
	args := m.Called(eventTicker)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*kalshi.EventResponse), args.Error(1)
}

func (m *MockMarketService) GetEvents(params kalshi.GetEventsOptions) (*kalshi.EventsResult, error) {
	args := m.Called(params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*kalshi.EventsResult), args.Error(1)
}
