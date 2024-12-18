package mocks

import "prediction-risk/internal/infrastructure/external/kalshi"

// MockKalshiClient follows the same structure as the real KalshiClient
type MockKalshiClient struct {
	baseClient *kalshi.baseClient // This can be nil for testing
	Market     *MockMarketClient
	Portfolio  *MockPortfolioClient
}

// Convert returns a *kalshi.KalshiClient that can be used where the real client is expected
func (m *MockKalshiClient) ToKalshiClient() *kalshi.KalshiClient {
	return &kalshi.KalshiClient{
		Market:    m.Market,
		Portfolio: m.Portfolio,
	}
}

// NewMockKalshiClient creates a new mock client with initialized sub-clients
func NewMockKalshiClient() *MockKalshiClient {
	return &MockKalshiClient{
		Market:    &MockMarketClient{},
		Portfolio: &MockPortfolioClient{},
	}
}
