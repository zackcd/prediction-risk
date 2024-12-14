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
	YesPrice       float64   `json:"yes_price"`
	NoPrice        float64   `json:"no_price"`
	YesAsk         float64   `json:"yes_ask"`
	NoAsk          float64   `json:"no_ask"`
	YesBid         float64   `json:"yes_bid"`
	NoBid          float64   `json:"no_bid"`
	LastPrice      float64   `json:"last_price"`
	PreviousPrice  float64   `json:"previous_price"`
	Volume         int       `json:"volume"`
	Volume24H      int       `json:"volume_24h"`
	OpenInterest   int       `json:"open_interest"`
	Result         *string   `json:"result,omitempty"` // Nullable
}

type MarketResponse struct {
	Market Market `json:"market"`
}
