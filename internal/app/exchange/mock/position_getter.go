package exchange_mock

import (
	"prediction-risk/internal/app/exchange/infrastructure/kalshi"

	"github.com/stretchr/testify/mock"
)

type MockPositionGetter struct {
	mock.Mock
}

func (m *MockPositionGetter) GetPositions(params kalshi.GetPositionsOptions) (*kalshi.PositionsResult, error) {
	args := m.Called(params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*kalshi.PositionsResult), args.Error(1)
}
