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

func (c *portfolioClient) GetPositions(opts GetPositionsOptions) (*PositionsResult, error) {
	result := &PositionsResult{
		MarketPositions: make([]MarketPosition, 0),
		EventPositions:  make([]EventPosition, 0),
	}

	params := NewPositionsParams().WithLimit(1000)
	if opts.Ticker != nil {
		params = params.WithTicker(*opts.Ticker)
	}
	if opts.EventTicker != nil {
		params = params.WithEventTicker(*opts.EventTicker)
	}
	if opts.SettlementStatus != nil {
		params = params.WithSettlementStatus(*opts.SettlementStatus)
	}

	if err := c.collectAllPositions(params, result); err != nil {
		return nil, fmt.Errorf("collecting positions: %w", err)
	}

	return result, nil
}

// fetchPage is a clear name for a single API call
func (c *portfolioClient) fetchPage(params PositionsParams) (*PositionsResponse, error) {
	resp, err := c.client.get(portfolioPath+"/positions", paramsToMap(params))
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	return handleResponse[PositionsResponse](resp)
}

// collectAllPositions is clearer than "recursive" in the name
func (c *portfolioClient) collectAllPositions(params PositionsParams, result *PositionsResult) error {
	for {
		page, err := c.fetchPage(params)
		if err != nil {
			return fmt.Errorf("fetching page: %w", err)
		}

		result.MarketPositions = append(result.MarketPositions, page.MarketPositions...)
		result.EventPositions = append(result.EventPositions, page.EventPositions...)

		if page.Cursor == nil {
			break
		}
		params = params.WithCursor(*page.Cursor)
	}

	return nil
}

// Helper to convert params struct to map for the client
func paramsToMap(params PositionsParams) map[string]string {
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
