package nws

type NWSClient struct {
	Station   *stationClient
	Gridpoint *gridpointClient
}

func NewNWSClient(baseURL string, userAgent string) *NWSClient {
	client := newClient(baseURL, userAgent)

	return &NWSClient{
		Station:   &stationClient{client},
		Gridpoint: &gridpointClient{client},
	}
}
