package entities

type Exchange string

const (
	ExchangeKalshi Exchange = "KALSHI"
)

type OrderAction string

const (
	OrderActionBuy  OrderAction = "BUY"
	OrderActionSell OrderAction = "SELL"
)

type OrderType string

const (
	OrderTypeLimit  OrderType = "LIMIT"
	OrderTypeMarket OrderType = "MARKET"
)

type OrderStatus string

type Order struct {
	ExchangeOrderID string
	Exchange        Exchange
	InternalOrderID string
	Ticker          string
	Side            Side
	Action          OrderAction
	OrderType       OrderType
	Status          OrderStatus
}

type Position struct{}
