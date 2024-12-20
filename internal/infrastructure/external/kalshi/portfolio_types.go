package kalshi

import "time"

type GetPositionsOptions struct {
	Ticker           *string
	EventTicker      *string
	SettlementStatus *string
}

func NewGetPositionsOptions() GetPositionsOptions {
	return GetPositionsOptions{}
}

func (o GetPositionsOptions) WithTicker(ticker string) GetPositionsOptions {
	o.EventTicker = &ticker
	return o
}

func (o GetPositionsOptions) WithEventTicker(eventTicker string) GetPositionsOptions {
	o.EventTicker = &eventTicker
	return o
}

func (o GetPositionsOptions) WithSettlementStatus(settlementStatus string) GetPositionsOptions {
	o.SettlementStatus = &settlementStatus
	return o
}

// Primary response type
type PositionsResponse struct {
	Cursor          *string          `json:"cursor"`
	EventPositions  []EventPosition  `json:"event_positions"`
	MarketPositions []MarketPosition `json:"market_positions"`
}

type PositionsResult struct {
	EventPositions  []EventPosition  `json:"event_positions"`
	MarketPositions []MarketPosition `json:"market_positions"`
}

type EventPosition struct {
	EventExposure     int    `json:"event_exposure"`
	EventTicker       string `json:"event_ticker"`
	FeesPaid          int    `json:"fees_paid"`
	RealizedPNL       int    `json:"realized_pnl"`
	RestingOrderCount int    `json:"resting_order_count"`
	TotalCost         int    `json:"total_cost"`
}

type MarketPosition struct {
	FeesPaid           int       `json:"fees_paid"`
	LastUpdatedTS      time.Time `json:"last_updated_ts"`
	MarketExposure     int       `json:"market_exposure"`
	Position           int       `json:"position"` // Number of contracts bought in this market. Negative means NO contracts and positive means YES contracts.
	RealizedPNL        int       `json:"realized_pnl"`
	RestingOrdersCount int       `json:"resting_orders_count"`
	Ticker             string    `json:"ticker"`
	TotalTradedCost    int       `json:"total_traded_cost"`
}

type OrderAction string

const (
	OrderActionBuy  OrderAction = "buy"
	OrderActionSell OrderAction = "sell"
)

// Request type
type CreateOrderRequest struct {
	Ticker            string      `json:"ticker"`
	ClientOrderID     string      `json:"client_order_id"`
	Side              OrderSide   `json:"side"`
	Action            OrderAction `json:"action"`
	Count             int         `json:"count"`
	Type              string      `json:"type"`                // "limit" or "market"
	YesPrice          *int        `json:"yes_price,omitempty"` // In cents
	NoPrice           *int        `json:"no_price,omitempty"`  // In cents
	ExpirationTs      *int64      `json:"expiration_ts,omitempty"`
	SellPositionFloor *int        `json:"sell_position_floor,omitempty"`
	BuyMaxCost        *int        `json:"buy_max_cost,omitempty"`
}

const (
	SettlementStatusOpen    = "open"
	SettlementStatusSettled = "settled"
	SettlementStatusClosed  = "closed"
)

// Response type
type CreateOrderResponse struct {
	Order Order `json:"order"`
}

type Order struct {
	Action         string    `json:"action"`
	ClientOrderID  string    `json:"client_order_id"`
	CreatedTime    time.Time `json:"created_time"`
	ExpirationTime time.Time `json:"expiration_time"`
	ID             string    `json:"order_id"`
	NoPrice        int       `json:"no_price"` // In cents
	Side           OrderSide `json:"side"`
	Status         string    `json:"status"`
	Ticker         string    `json:"ticker"`
	Type           string    `json:"type"`      // "limit" or "market"
	YesPrice       int       `json:"yes_price"` // In cents
}

type OrderSide string

// Constants for the various enums
const (
	// Side
	OrderSideYes OrderSide = "yes"
	OrderSideNo  OrderSide = "no"
)

const (
	// Type
	OrderTypeLimit  = "limit"
	OrderTypeMarket = "market"

	// Status
	OrderStatusOpen     = "open"
	OrderStatusCanceled = "canceled"
	OrderStatusExecuted = "executed"
	OrderStatusExpired  = "expired"
	OrderStatusReduced  = "reduced"
)
