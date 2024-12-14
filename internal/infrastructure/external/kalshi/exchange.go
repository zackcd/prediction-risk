package kalshi

type exchangeClient struct {
	client *baseClient
}

func NewExchangeClient(client *baseClient) *exchangeClient {
	return &exchangeClient{
		client: client,
	}
}

func (c *exchangeClient) GetExchangeSchedule() (*ExchangeScheduleResponse, error) {
	resp, err := c.client.get(exchangePath+"/schedule", nil)
	if err != nil {
		return nil, err
	}
	return handleResponse[ExchangeScheduleResponse](resp)
}

func (c *exchangeClient) GetExchangeAnnouncements() (*ExchangeAnnouncementsResponse, error) {
	resp, err := c.client.get(exchangePath+"/announcements", nil)
	if err != nil {
		return nil, err
	}
	return handleResponse[ExchangeAnnouncementsResponse](resp)
}
