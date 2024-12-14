package kalshi

import (
	"time"
)

type MaintenanceWindow struct {
	EndDatetime   time.Time `json:"end_datetime"`
	StartDatetime time.Time `json:"start_datetime"`
}

type TradingHours struct {
	OpenTime  string `json:"open_time"`
	CloseTime string `json:"close_time"`
}

type StandardHours struct {
	EndTime   time.Time      `json:"end_time"`
	StartTime time.Time      `json:"start_time"`
	Monday    []TradingHours `json:"monday"`
	Tuesday   []TradingHours `json:"tuesday"`
	Wednesday []TradingHours `json:"wednesday"`
	Thursday  []TradingHours `json:"thursday"`
	Friday    []TradingHours `json:"friday"`
	Saturday  []TradingHours `json:"saturday"`
	Sunday    []TradingHours `json:"sunday"`
}

type Schedule struct {
	MaintenanceWindows []MaintenanceWindow `json:"maintenance_windows"`
	StandardHours      []StandardHours     `json:"standard_hours"`
}

type ExchangeScheduleResponse struct {
	Schedule Schedule `json:"schedule"`
}

type Announcement struct {
	DeliveryTime time.Time `json:"delivery_time"`
	Message      string    `json:"message"`
	Status       string    `json:"status"`
	Type         string    `json:"type"`
}

type ExchangeAnnouncementsResponse struct {
	Announcements []Announcement `json:"announcements"`
}

type Market struct {
	Ticker         string    `json:"ticker"`
	Title          string    `json:"title"`
	Status         string    `json:"status"`
	OpenTime       time.Time `json:"open_time"`
	CloseTime      time.Time `json:"close_time"`
	ExpirationTime time.Time `json:"expiration_time"`
	Category       string    `json:"category"`
	SubCategory    string    `json:"sub_category"`
	YesPrice       int       `json:"yes_price"`
	NoPrice        int       `json:"no_price"`
	YesAsk         int       `json:"yes_ask"`
	NoAsk          int       `json:"no_ask"`
	YesBid         int       `json:"yes_bid"`
	NoBid          int       `json:"no_bid"`
	LastPrice      int       `json:"last_price"`
	PreviousPrice  int       `json:"previous_price"`
	Volume         int       `json:"volume"`
	Volume24H      int       `json:"volume_24h"`
	OpenInterest   int       `json:"open_interest"`
	Result         *string   `json:"result,omitempty"` // Nullable
}

type MarketResponse struct {
	Market Market `json:"market"`
}

// Primary response type
type PositionsResponse struct {
	Cursor          *string          `json:"cursor"`
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
	MarketExposure     int       `json:"market_exposure`
	Position           int       `json:"position"` // Number of contracts bought in this market. Negative means NO contracts and positive means YES contracts.
	RealizedPNL        int       `json:"realized_pnl"`
	RestingOrdersCount int       `json:"resting_orders_count"`
	Ticker             string    `json:"ticker"`
	TotalTradedCost    int       `json:"total_traded_cost"`
}

// Optional query parameters for the request
type PositionsParams struct {
	Cursor           *string `json:"cursor,omitempty"`
	Limit            *int    `json:"limit,omitempty"`
	SettlementStatus *string `json:"settlement_status,omitempty"`
	Ticker           *string `json:"ticker,omitempty"`
	EventTicker      *string `json:"event_ticker,omitempty"`
}

// Helper to create params with cleaner syntax
func NewPositionsParams() *PositionsParams {
	return &PositionsParams{}
}

// Builder pattern for setting optional fields
func (p *PositionsParams) WithCursor(cursor string) *PositionsParams {
	p.Cursor = &cursor
	return p
}

func (p *PositionsParams) WithLimit(limit int) *PositionsParams {
	p.Limit = &limit
	return p
}

func (p *PositionsParams) WithTicker(ticker string) *PositionsParams {
	p.Ticker = &ticker
	return p
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
