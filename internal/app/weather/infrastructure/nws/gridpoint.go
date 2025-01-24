package nws

import "fmt"

type gridpointClient struct {
	*client
}

func (c *gridpointClient) GetForecast(
	officeID string,
	xCoordinate int,
	yCoordinate int,
) (*Forecast, error) {
	path := fmt.Sprintf("/gridpoints/%s/%d,%d/forecast", officeID, xCoordinate, yCoordinate)
	resp, err := c.get(path, nil)
	if err != nil {
		return nil, err
	}
	return handleResponse[Forecast](resp)
}

func (c *gridpointClient) GetHourlyForecast(
	officeID string,
	xCoordinate int,
	yCoordinate int,
) (*Forecast, error) {
	path := fmt.Sprintf("/gridpoints/%s/%d,%d/forecast/hourly", officeID, xCoordinate, yCoordinate)
	resp, err := c.get(path, nil)
	if err != nil {
		return nil, err
	}
	return handleResponse[Forecast](resp)
}
