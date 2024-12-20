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

func (c *portfolioClient) CreateOrder(order CreateOrderRequest) (*CreateOrderResponse, error) {
	resp, err := c.client.post(portfolioPath+"/order", order)
	if err != nil {
		return nil, err
	}
	return handleResponse[CreateOrderResponse](resp)
}

func (c *portfolioClient) GetPositions(params GetPositionsOptions) (*PositionsResult, error) {
	result := &PositionsResult{
		MarketPositions: make([]MarketPosition, 0),
		EventPositions:  make([]EventPosition, 0),
	}

	if err := c.collectAllPositions(params, result); err != nil {
		return nil, fmt.Errorf("collecting positions: %w", err)
	}

	return result, nil
}

// collectAllPositions is clearer than "recursive" in the name
func (c *portfolioClient) collectAllPositions(params GetPositionsOptions, result *PositionsResult) error {
	var cursor *string

	for {
		page, err := c.fetchPage(params, cursor, nil)
		if err != nil {
			return fmt.Errorf("fetching page: %w", err)
		}

		result.MarketPositions = append(result.MarketPositions, page.MarketPositions...)
		result.EventPositions = append(result.EventPositions, page.EventPositions...)

		if page.Cursor == nil || len(page.EventPositions)+len(page.MarketPositions) == 0 {
			break
		}
		cursor = page.Cursor
	}

	return nil
}

func (c *portfolioClient) fetchPage(params GetPositionsOptions, cursor *string, limit *int) (*PositionsResponse, error) {
	resp, err := c.client.get(portfolioPath+"/positions", portfolioParamsToMap(params, cursor, limit))
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	return handleResponse[PositionsResponse](resp)
}

// Helper to convert params struct to map for the client
func portfolioParamsToMap(params GetPositionsOptions, cursor *string, limit *int) map[string]string {
	result := make(map[string]string)
	if cursor != nil {
		result["cursor"] = *cursor
	}
	if limit != nil {
		result["limit"] = strconv.Itoa(*limit)
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
