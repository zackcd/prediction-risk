package nws

import (
	"fmt"
)

type stationClient struct {
	*client
}

func (c *stationClient) Get(stationID string) (*Station, error) {
	path := fmt.Sprintf("/stations/%s", stationID)
	resp, err := c.get(path, nil)
	if err != nil {
		return nil, err
	}
	return handleResponse[Station](resp)
}

func (c *stationClient) List(params *StationQueryParams) (*StationCollection, error) {
	path := fmt.Sprintf("/stations")
	resp, err := c.get(path, nil)
	if err != nil {
		return nil, err
	}
	return handleResponse[StationCollection](resp)
}

func (c *stationClient) collectAllStations(params StationQueryParams) {}

func (c *stationClient) GetObservations(stationID string) (*ObservationCollection, error) {
	path := fmt.Sprintf("/stations/%s/observations", stationID)
	resp, err := c.get(path, nil)
	if err != nil {
		return nil, err
	}
	return handleResponse[ObservationCollection](resp)
}

func (c *stationClient) collectAllObservations(stationID string) {}

func (c *stationClient) GetLatestObservations(stationID string) (*Observation, error) {
	path := fmt.Sprintf("/stations/%s/observations/latest", stationID)
	resp, err := c.get(path, nil)
	if err != nil {
		return nil, err
	}
	return handleResponse[Observation](resp)
}
