package kalshi

import (
	"crypto/rsa"
)

type KalshiClient struct {
	client *client

	Portfolio *portfolioClient
	Market    *marketClient
	Event     *eventClient
}

func NewKalshiClient(host, keyID string, privateKey *rsa.PrivateKey) *KalshiClient {
	client := newClient(host, keyID, privateKey)

	return &KalshiClient{
		client: client,

		Portfolio: NewPortfolioClient(client),
		Market:    NewMarketClient(client),
		Event:     newEventClient(client),
	}
}
