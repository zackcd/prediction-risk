package exchange

import "prediction-risk/internal/domain/contract"

type Exchange string

const (
	ExchangeKalshi Exchange = "KALSHI"
)

type OrderAction string

const (
	OrderActionBuy  OrderAction = "BUY"
	OrderActionSell OrderAction = "SELL"
)

type ExchangeOrderType string

const (
	OrderTypeLimit  ExchangeOrderType = "LIMIT"
	OrderTypeMarket ExchangeOrderType = "MARKET"
)

type ExchangeOrder struct {
	ExchangeOrderID string
	Exchange        Exchange
	InternalOrderID string
	Ticker          string
	Side            contract.Side
	Action          OrderAction
	OrderType       ExchangeOrderType
	Status          string
}
