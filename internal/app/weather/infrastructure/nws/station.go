package nws

import (
	"fmt"
	"net/url"
)

type stationClient struct {
	*client
}

func (c *stationClient) GetStation(stationID string) (*Station, error) {
	path := fmt.Sprintf("/stations/%s", stationID)
	resp, err := c.get(path, nil)
	if err != nil {
		return nil, err
	}
	return handleResponse[Station](resp)
}

// ListAllStations retrieves all stations using pagination
func (c *stationClient) ListAllStations() (*StationCollection, error) {
	result := &StationCollection{
		Type:                CollectionType,
		Features:            make([]Station, 0),
		ObservationStations: make([]string, 0),
	}

	var cursor *string
	for {
		params := make(map[string]string)
		if cursor != nil {
			params["cursor"] = *cursor
		}

		resp, err := c.get("/stations", params)
		if err != nil {
			return nil, fmt.Errorf("fetching stations page: %w", err)
		}

		page, err := handleResponse[StationsResponse](resp)
		if err != nil {
			return nil, fmt.Errorf("parsing stations page: %w", err)
		}

		result.Features = append(result.Features, page.Features...)
		result.ObservationStations = append(result.ObservationStations, page.ObservationStations...)

		// Break if no more pages
		if page.Pagination == nil || len(page.ObservationStations) == 0 {
			break
		}
		nextUrl, err := url.Parse(page.Pagination.Next)
		if err != nil {
			return nil, fmt.Errorf("parsing next pagination URL: %w", err)
		}

		cursorStr := nextUrl.Query().Get("cursor")
		cursor = &cursorStr
	}

	return result, nil
}

func (c *stationClient) GetObservations(stationID string, params ObservationQueryParams) (*ObservationCollection, error) {
	path := fmt.Sprintf("/stations/%s/observations", stationID)
	resp, err := c.get(path, observationParamsToMap(params))
	if err != nil {
		return nil, err
	}
	return handleResponse[ObservationCollection](resp)
}

func (c *stationClient) GetLatestObservations(stationID string) (*Observation, error) {
	path := fmt.Sprintf("/stations/%s/observations/latest", stationID)
	resp, err := c.get(path, nil)
	if err != nil {
		return nil, err
	}
	return handleResponse[Observation](resp)
}

func observationParamsToMap(params ObservationQueryParams) map[string]string {
	m := make(map[string]string)
	if params.Start != nil {
		m["start"] = params.Start.Format("2006-01-02T15:04:05Z")
	}
	if params.End != nil {
		m["end"] = params.End.Format("2006-01-02T15:04:05Z")
	}
	return m
}
