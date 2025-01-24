package nws

import "time"

type ForecastPeriod struct {
	Number                     int               `json:"number"`
	Name                       string            `json:"name"`
	StartTime                  time.Time         `json:"startTime"`
	EndTime                    time.Time         `json:"endTime"`
	IsDaytime                  bool              `json:"isDaytime"`
	TemperatureTrend           string            `json:"temperatureTrend"`
	ProbabilityOfPrecipitation QuantitativeValue `json:"probabilityOfPrecipitation"`
	Dewpoint                   QuantitativeValue `json:"dewpoint"`
	RelativeHumidity           QuantitativeValue `json:"relativeHumidity"`
	WindDirection              string            `json:"windDirection"`
	ShortForecast              string            `json:"shortForecast"`
	DetailedForecast           string            `json:"detailedForecast"`
}

type ForecastProperties struct {
	Geometry          string            `json:"geometry"`
	Units             string            `json:"units"`
	ForecastGenerator string            `json:"forecastGenerator"`
	GeneratedAt       time.Time         `json:"generatedAt"`
	UpdateTime        time.Time         `json:"updateTime"`
	Elevation         QuantitativeValue `json:"elevation"`
	Periods           []ForecastPeriod  `json:"periods"`
}

type Forecast struct {
	ID         string             `json:"id"`
	Type       string             `json:"type"`
	Properties ForecastProperties `json:"properties"`
}
