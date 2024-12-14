package kalshi

import (
	"fmt"
	"strconv"
)

type portfolioClient struct {
	client *baseClient
}

func NewPortfolioClient(client *baseClient) *portfolioClient {
	return &portfolioClient{
		client: client,
	}
}

func (c *portfolioClient) CreateOrder(order *CreateOrderRequest) (*CreateOrderResponse, error) {
	resp, err := c.client.post(portfolioPath+"/order", order)
	if err != nil {
		return nil, err
	}
	return handleResponse[CreateOrderResponse](resp)
}

func (c *portfolioClient) GetPositions(params *PositionsParams) (*PositionsResponse, error) {
	resp, err := c.client.get(portfolioPath+"/positions", paramsToMap(params))
	if err != nil {
		return nil, fmt.Errorf("getting positions: %w", err)
	}
	return handleResponse[PositionsResponse](resp)
}

// Helper to convert params struct to map for the client
func paramsToMap(params *PositionsParams) map[string]string {
	if params == nil {
		return nil
	}

	result := make(map[string]string)
	if params.Cursor != nil {
		result["cursor"] = *params.Cursor
	}
	if params.Limit != nil {
		result["limit"] = strconv.Itoa(*params.Limit)
	}
	if params.SettlementStatus != nil {
		result["settlement_status"] = *params.SettlementStatus
	}
	if params.Ticker != nil {
		result["ticker"] = *params.Ticker
	}
	if params.EventTicker != nil {
		result["event_ticker"] = *params.EventTicker
	}
	return result
}
