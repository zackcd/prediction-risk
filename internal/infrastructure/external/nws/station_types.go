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

type QuantitativeValue struct {
	Value          float64         `json:"value"`
	UnitCode       string          `json:"unitCode"`
	MaxValue       *float64        `json:"maxValue,omitempty"`
	MinValue       *float64        `json:"minValue,omitempty"`
	QualityControl *QualityControl `json:"qualityControl,omitempty"`
}

type Pagination struct {
	Next string `json:"next"`
}

const (
	CollectionType      = "FeatureCollection"
	StationType         = "Feature"
	GeometryType        = "Point"
	StationPropertyType = "wx:ObservationStation"
)

type QualityControl string

const (
	QualityControlZ = "Z"
	QualityControlC = "C"
	QualityControlS = "S"
	QualityControlV = "V"
	QualityControlX = "X"
	QualityControlQ = "Q"
	QualityControlG = "G"
	QualityControlB = "B"
	QualityControlT = "T"
)

// IsValid returns true if the quality control value is valid
func (q QualityControl) IsValid() bool {
	switch q {
	case QualityControlZ, QualityControlC, QualityControlS,
		QualityControlV, QualityControlX, QualityControlQ,
		QualityControlG, QualityControlB, QualityControlT:
		return true
	default:
		return false
	}
}
