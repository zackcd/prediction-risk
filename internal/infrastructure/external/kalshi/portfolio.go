package kalshi

type portfolioClient struct {
	client *baseClient
}

func NewPortfolioClient(client *baseClient) *portfolioClient {
	return &portfolioClient{
		client: client,
	}
}

func (c *portfolioClient) CreateOrder(order *Order) (*OrderResponse, error) {
	resp, err := c.client.post(portfolioPath+"/order", order)
	if err != nil {
		return nil, err
	}
	return handleResponse[OrderResponse](resp)
}

func (c *portfolioClient) GetPositions() (*PositionResponse, error) {
	resp, err := c.client.get(portfolioPath+"/positions", nil)
	if err != nil {
		return nil, err
	}
	return handleResponse[PositionResponse](resp)
}
