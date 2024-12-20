package kalshi

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

type marketClient struct {
	client *baseClient
}

func NewMarketClient(client *baseClient) *marketClient {
	return &marketClient{client: client}
}

func (c *marketClient) GetMarket(ticker string) (*MarketResponse, error) {
	resp, err := c.client.get(marketsPath+"/"+ticker, nil)
	if err != nil {
		return nil, err
	}
	return handleResponse[MarketResponse](resp)
}

func (c *marketClient) GetMarkets(params GetMarketsOptions) (*MarketsResult, error) {
	result := &MarketsResult{
		Markets: make([]Market, 0),
	}

	if err := c.collectAllMarkets(params, result); err != nil {
		return nil, fmt.Errorf("collecting markets: %w", err)
	}

	return result, nil
}

func (c *marketClient) collectAllMarkets(params GetMarketsOptions, result *MarketsResult) error {
	var cursor *string

	for {
		page, err := c.fetchPage(params, cursor, nil)
		if err != nil {
			return fmt.Errorf("fetching page: %w", err)
		}

		result.Markets = append(result.Markets, page.Markets...)

		if page.Cursor == nil {
			break
		}
		cursor = page.Cursor
	}

	return nil
}

func (c *marketClient) fetchPage(params GetMarketsOptions, cursor *string, limit *int) (*MarketsResponse, error) {
	resp, err := c.client.get(marketsPath, marketParamsToMap(params, cursor, limit))
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	return handleResponse[MarketsResponse](resp)
}

// Helper to convert params struct to map for the client
func marketParamsToMap(params GetMarketsOptions, cursor *string, limit *int) map[string]string {
	result := make(map[string]string)
	if cursor != nil {
		result["cursor"] = *cursor
	}
	if limit != nil {
		result["limit"] = strconv.Itoa(*limit)
	}
	if params.Tickers != nil {
		result["ticker"] = strings.Join(*params.Tickers, ",")
	}
	if params.EventTicker != nil {
		result["event_ticker"] = *params.EventTicker
	}
	if params.SeriesTicker != nil {
		result["series_ticker"] = *params.SeriesTicker
	}
	if params.MaxCloseTime != nil {
		result["max_close_time"] = params.MaxCloseTime.Format(time.StampMilli)
	}
	if params.MinCloseTime != nil {
		result["min_close_time"] = params.MinCloseTime.Format(time.StampMilli)
	}
	if params.Status != nil {
		result["status"] = strings.Join(*params.Status, ",")
	}
	return result
}
