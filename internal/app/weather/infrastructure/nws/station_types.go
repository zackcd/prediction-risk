package nws

type StationCollection struct {
	Type                string    `json:"type"`
	Features            []Station `json:"features"`
	ObservationStations []string  `json:"observationStations"`
}

type StationsResponse struct {
	Type                string      `json:"type"`
	Features            []Station   `json:"features"`
	ObservationStations []string    `json:"observationStations"`
	Pagination          *Pagination `json:"pagination"`
}

type Station struct {
	ID         string            `json:"id"`
	Type       string            `json:"type"`
	Geometry   Geometry          `json:"geometry"`
	Properties StationProperties `json:"properties"`
}

type Geometry struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

type StationProperties struct {
	ID                string            `json:"@id"`
	Type              string            `json:"@type"`
	Elevation         QuantitativeValue `json:"elevation"`
	StationIdentifier string            `json:"stationIdentifier"`
	Name              string            `json:"name"`
	TimeZone          string            `json:"timeZone"`
	Forecast          string            `json:"forecast"`
	County            string            `json:"county"`
	FireWeatherZone   string            `json:"fireWeatherZone"`
}
