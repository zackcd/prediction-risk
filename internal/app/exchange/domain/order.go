package exchange_domain

import (
	"prediction-risk/internal/app/contract"
	"time"

	"github.com/google/uuid"
)

type OrderID uuid.UUID

func NewOrderID() OrderID {
	return OrderID(uuid.New())
}

func (o OrderID) String() string {
	return uuid.UUID(o).String()
}

type OrderAction string

const (
	OrderActionBuy  OrderAction = "BUY"
	OrderActionSell OrderAction = "SELL"
)

type MarketOrderType string

const (
	OrderTypeLimit  MarketOrderType = "LIMIT"
	OrderTypeMarket MarketOrderType = "MARKET"
)

type Order struct {
	OrderID         OrderID
	ExchangeOrderID string
	Exchange        Exchange
	Reference       string
	Ticker          string
	Side            contract.Side
	Action          OrderAction
	OrderType       MarketOrderType
	Status          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func NewOrder(
	exchangeOrderID string,
	exchange Exchange,
	reference string,
	ticker string,
	side contract.Side,
	action OrderAction,
	orderType MarketOrderType,
	status string,
) *Order {
	currentTime := time.Now()
	return &Order{
		OrderID:         NewOrderID(),
		ExchangeOrderID: exchangeOrderID,
		Exchange:        exchange,
		Reference:       reference,
		Ticker:          ticker,
		Side:            side,
		Action:          action,
		OrderType:       orderType,
		Status:          status,
		CreatedAt:       currentTime,
		UpdatedAt:       currentTime,
	}
}
