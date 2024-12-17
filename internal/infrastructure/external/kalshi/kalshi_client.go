package kalshi

type KalshiClient struct {
	baseClient *baseClient

	Portfolio *portfolioClient
	Market    *marketClient
}

func NewKalshiClient(baseClient *baseClient) *KalshiClient {
	return &KalshiClient{
		baseClient: baseClient,

		Portfolio: NewPortfolioClient(baseClient),
		Market:    NewMarketClient(baseClient),
	}
}
