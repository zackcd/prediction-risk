package kalshi

import (
	"time"
)

type GetMarketsOptions struct {
	Tickers      *[]string
	EventTicker  *string
	SeriesTicker *string
	MaxCloseTime *time.Time
	MinCloseTime *time.Time
	Status       *[]string
}

func NewGetMarketsOptions() GetMarketsOptions {
	return GetMarketsOptions{}
}

func (o GetMarketsOptions) WithTickers(tickers []string) GetMarketsOptions {
	o.Tickers = &tickers
	return o
}

func (o GetMarketsOptions) WithEventTicker(eventTicker string) GetMarketsOptions {
	o.EventTicker = &eventTicker
	return o
}

func (o GetMarketsOptions) WithSeriesTicker(seriesTicker string) GetMarketsOptions {
	o.SeriesTicker = &seriesTicker
	return o
}

func (o GetMarketsOptions) WithMaxCloseTime(maxCloseTime time.Time) GetMarketsOptions {
	o.MaxCloseTime = &maxCloseTime
	return o
}

func (o GetMarketsOptions) WithMinCloseTime(minCloseTime time.Time) GetMarketsOptions {
	o.MinCloseTime = &minCloseTime
	return o
}

func (o GetMarketsOptions) WithStatus(status []string) GetMarketsOptions {
	o.Status = &status
	return o
}

type Market struct {
	// Market Information
	Ticker             string  `json:"ticker"`
	EventTicker        string  `json:"event_ticker"`
	MultiMarketEventID *string `json:"multi_market_event_id,omitempty"`
	Title              string  `json:"title"`
	Subtitle           string  `json:"subtitle"`
	YesSubTitle        string  `json:"yes_sub_title"`
	NoSubTitle         string  `json:"no_sub_title"`
	Description        *string `json:"description,omitempty"`
	ImageURL           *string `json:"image_url,omitempty"`
	Category           string  `json:"category"`
	SubCategory        *string `json:"sub_category,omitempty"`
	StrikePrice        *string `json:"strike_price,omitempty"`
	MarketType         string  `json:"market_type"`

	// Status & Timing
	Status                 string    `json:"status"`
	OpenTime               time.Time `json:"open_time"`
	CloseTime              time.Time `json:"close_time"`
	ExpirationTime         time.Time `json:"expiration_time"`
	ExpectedExpirationTime time.Time `json:"expected_expiration_time"`
	LatestExpirationTime   time.Time `json:"latest_expiration_time"`
	SettlementTimerSeconds int       `json:"settlement_timer_seconds"`
	CanCloseEarly          bool      `json:"can_close_early"`

	// Settlement Information
	Result            *string    `json:"result,omitempty"`
	Settlement        *string    `json:"settlement,omitempty"`
	SettlementNotes   *string    `json:"settlement_notes,omitempty"`
	SettlementSources *string    `json:"settlement_sources,omitempty"`
	SettlementTime    *time.Time `json:"settlement_time,omitempty"`
	ExpirationValue   string     `json:"expiration_value"`
	Rules             *string    `json:"rules,omitempty"`
	RulesPrimary      string     `json:"rules_primary"`
	RulesSecondary    string     `json:"rules_secondary"`

	// Market Data
	MaxBinaryValue     *int   `json:"max_binary_value,omitempty"`
	LastPrice          int    `json:"last_price"`
	PreviousPrice      int    `json:"previous_price"`
	YesPrice           *int   `json:"yes_price,omitempty"`
	NoPrice            *int   `json:"no_price,omitempty"`
	YesBid             int    `json:"yes_bid"`
	NoBid              int    `json:"no_bid"`
	YesAsk             int    `json:"yes_ask"`
	NoAsk              int    `json:"no_ask"`
	PreviousYesBid     int    `json:"previous_yes_bid"`
	PreviousYesAsk     int    `json:"previous_yes_ask"`
	ResponsePriceUnits string `json:"response_price_units"`
	NotionalValue      int    `json:"notional_value"`
	TickSize           int    `json:"tick_size"`
	RiskLimitCents     int    `json:"risk_limit_cents"`

	// Volume & Interest
	Volume       int  `json:"volume"`
	Volume24H    int  `json:"volume_24h"`
	OpenInterest int  `json:"open_interest"`
	Liquidity    int  `json:"liquidity"`
	Views        *int `json:"views,omitempty"`
	ViewsChange  *int `json:"views_change,omitempty"`
}

type MarketResponse struct {
	Market Market `json:"market"`
}

type MarketsResponse struct {
	Cursor  *string  `json:"cursor,omitempty"`
	Markets []Market `json:"markets"`
}

type MarketsResult struct {
	Markets []Market
}
