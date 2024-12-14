package kalshi

type KalshiClient struct {
	baseClient *baseClient

	Exchange  *exchangeClient
	Portfolio *portfolioClient
	Market    *marketClient
}

func NewKalshiClient(baseClient *baseClient) *KalshiClient {
	return &KalshiClient{
		baseClient: baseClient,

		Exchange:  NewExchangeClient(baseClient),
		Portfolio: NewPortfolioClient(baseClient),
		Market:    NewMarketClient(baseClient),
	}
}
