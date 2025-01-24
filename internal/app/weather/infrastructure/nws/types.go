package nws

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
