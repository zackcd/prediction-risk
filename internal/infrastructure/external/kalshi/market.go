package kalshi

type marketClient struct {
	client *baseClient
}

func NewMarketClient(client *baseClient) *marketClient {
	return &marketClient{
		client: client,
	}
}

func (c *marketClient) GetMarket(ticker string) (*MarketResponse, error) {
	resp, err := c.client.get(marketsPath+"/markets/"+ticker, nil)
	if err != nil {
		return nil, err
	}
	return handleResponse[MarketResponse](resp)
}
