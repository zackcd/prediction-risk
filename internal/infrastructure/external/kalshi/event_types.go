package kalshi

import "time"

type Event struct {
	EventTicker          string     `json:"event_ticker"`
	SeriesTicker         string     `json:"series_ticker"`
	CollateralReturnType string     `json:"collateral_return_type"`
	MutuallyExclusive    bool       `json:"mutually_exclusive"`
	Title                string     `json:"title"`
	SubTitle             string     `json:"sub_title"`
	StrikeDate           *time.Time `json:"strike_date,omitempty"`
	StrikePeriod         *string    `json:"strike_period,omitempty"`
}

type GetEventsOptions struct {
	Statuses     *[]string
	SeriesTicker *string
	Limit        *int
}

func NewGetEventsOptions() GetEventsOptions {
	return GetEventsOptions{}
}

func (o GetEventsOptions) WithStatuses(statuses []string) GetEventsOptions {
	o.Statuses = &statuses
	return o
}

func (o GetEventsOptions) WithSeriesTicker(seriesTicker string) GetEventsOptions {
	o.SeriesTicker = &seriesTicker
	return o
}

func (o GetEventsOptions) WithLimit(limit int) GetEventsOptions {
	o.Limit = &limit
	return o
}

type EventResponse struct {
	Event   Event    `json:"event"`
	Markets []Market `json:"markets,omitempty"`
}

type EventsResponse struct {
	Cursor *string `json:"cursor,omitempty"`
	Events []Event `json:"events"`
}

type EventsResult struct {
	Events []Event
}
