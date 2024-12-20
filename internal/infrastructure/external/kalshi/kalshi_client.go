package kalshi

import (
	"crypto/rsa"
)

type KalshiClient struct {
	baseClient *baseClient

	Portfolio *portfolioClient
	Market    *marketClient
	Event     *eventClient
}

func NewKalshiClient(host, keyID string, privateKey *rsa.PrivateKey) *KalshiClient {
	baseClient := NewBaseClient(host, keyID, privateKey)

	return &KalshiClient{
		baseClient: baseClient,

		Portfolio: NewPortfolioClient(baseClient),
		Market:    NewMarketClient(baseClient),
		Event:     newEventClient(baseClient),
	}
}
