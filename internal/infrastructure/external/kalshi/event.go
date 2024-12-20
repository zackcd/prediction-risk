package kalshi

import (
	"fmt"
	"strconv"
	"strings"
)

type eventClient struct {
	client *baseClient
}

func newEventClient(client *baseClient) *eventClient {
	return &eventClient{client: client}
}

func (c *eventClient) GetEvent(eventTicker string) (*EventResponse, error) {
	resp, err := c.client.get(eventsPath+"/"+eventTicker, nil)
	if err != nil {
		return nil, err
	}
	return handleResponse[EventResponse](resp)
}

func (c *eventClient) GetEvents(params GetEventsOptions) (*EventsResult, error) {
	result := &EventsResult{
		Events: make([]Event, 0),
	}

	if err := c.collectAllEvents(params, result); err != nil {
		return nil, fmt.Errorf("collecting events: %w", err)
	}

	return result, nil
}

func (c *eventClient) collectAllEvents(params GetEventsOptions, result *EventsResult) error {
	var cursor *string
	var remaining int
	if params.Limit != nil {
		remaining = *params.Limit
	} else {
		remaining = 200 // default limit
	}

	for {
		// Calculate page size for this request
		pageSize := remaining
		if pageSize > 200 { // Assuming API max page size is 100
			pageSize = 200
		}

		page, err := c.fetchPage(params, cursor, &pageSize)
		if err != nil {
			return fmt.Errorf("fetching page: %w", err)
		}

		// Only take what we need from this page
		if len(page.Events) > remaining {
			result.Events = append(result.Events, page.Events[:remaining]...)
		} else {
			result.Events = append(result.Events, page.Events...)
		}

		remaining -= len(page.Events)

		if remaining <= 0 || page.Cursor == nil || len(page.Events) == 0 {
			break
		}
		cursor = page.Cursor
	}

	return nil
}

func (c *eventClient) fetchPage(params GetEventsOptions, cursor *string, limit *int) (*EventsResponse, error) {
	resp, err := c.client.get(eventsPath, eventParamsToMap(params, cursor, limit))
	if err != nil {
		return nil, fmt.Errorf("API request failed: %w", err)
	}
	return handleResponse[EventsResponse](resp)
}

// Helper to convert params struct to map for the client
func eventParamsToMap(params GetEventsOptions, cursor *string, limit *int) map[string]string {
	result := make(map[string]string)
	if cursor != nil {
		result["cursor"] = *cursor
	}
	if limit != nil {
		result["limit"] = strconv.Itoa(*limit)
	}
	if params.SeriesTicker != nil {
		result["series_ticker"] = *params.SeriesTicker
	}
	if params.Statuses != nil {
		result["status"] = strings.Join(*params.Statuses, ",")
	}
	result["with_nested_markets"] = strconv.FormatBool(true)
	return result
}
