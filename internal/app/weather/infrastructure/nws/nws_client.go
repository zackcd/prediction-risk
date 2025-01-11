package nws

type NWSClient struct {
	Station *stationClient
}

func NewNWSClient(baseURL string, userAgent string) *NWSClient {
	client := newClient(baseURL, userAgent)

	return &NWSClient{
		Station: &stationClient{client},
	}
}
