package nws

import "time"

// Collection response for observations
type ObservationCollection struct {
    Type     string        `json:"type"`
    Features []Observation `json:"features"`
}

// Single observation and Latest observation response
type Observation struct {
    ID         string               `json:"id"`
    Type       string               `json:"type"`
    Properties ObservationProperties `json:"properties"`
}

type ObservationProperties struct {
    Geometry                    string              `json:"geometry"`
    ID                         string              `json:"@id"`
    Type                       string              `json:"@type"`
    Elevation                  *QuantitativeValue   `json:"elevation"`
    Station                    string              `json:"station"`
    Timestamp                  time.Time           `json:"timestamp"`
    RawMessage                 string              `json:"rawMessage"`
    TextDescription            string              `json:"textDescription"`
    PresentWeather            []PresentWeather    `json:"presentWeather"`
    Temperature               *QuantitativeValue   `json:"temperature"`
    Dewpoint                  *QuantitativeValue   `json:"dewpoint"`
    WindDirection             *QuantitativeValue   `json:"windDirection"`
    WindSpeed                 *QuantitativeValue   `json:"windSpeed"`
    WindGust                  *QuantitativeValue   `json:"windGust"`
    BarometricPressure        *QuantitativeValue   `json:"barometricPressure"`
    SeaLevelPressure          *QuantitativeValue   `json:"seaLevelPressure"`
    Visibility                *QuantitativeValue   `json:"visibility"`
    MaxTemperatureLast24Hours *QuantitativeValue   `json:"maxTemperatureLast24Hours"`
    MinTemperatureLast24Hours *QuantitativeValue   `json:"minTemperatureLast24Hours"`
    PrecipitationLastHour     *QuantitativeValue   `json:"precipitationLastHour"`
    PrecipitationLast3Hours   *QuantitativeValue   `json:"precipitationLast3Hours"`
    PrecipitationLast6Hours   *QuantitativeValue   `json:"precipitationLast6Hours"`
    RelativeHumidity          *QuantitativeValue   `json:"relativeHumidity"`
    WindChill                 *QuantitativeValue   `json:"windChill"`
    HeatIndex                 *QuantitativeValue   `json:"heatIndex"`
    CloudLayers               []CloudLayer         `json:"cloudLayers,omitempty"`
}

type PresentWeather struct {
    Intensity   WeatherIntensity `json:"intensity,omitempty"`
    Modifier    WeatherModifier  `json:"modifier,omitempty"`
    Weather     WeatherType      `json:"weather"`
    RawString   string           `json:"rawString"`
    InVicinity  bool            `json:"inVicinity"`
}

type CloudLayer struct {
    Base   QuantitativeValue `json:"base"`
    Amount CloudAmount      `json:"amount"`
}

// Add methods to validate enum values
func (w WeatherIntensity) IsValid() bool {
    switch w {
    case WeatherIntensityVeryLight, WeatherIntensityLight,
         WeatherIntensityModerate, WeatherIntensityHeavy:
        return true
    }
    return false
}

func (m WeatherModifier) IsValid() bool {
    switch m {
    case ModifierPatches, ModifierBlowing, ModifierLowDrifting,
         ModifierFreezing, ModifierShallow, ModifierPartial, ModifierShowers:
        return true
    }
    return false
}

func (w WeatherType) IsValid() bool {
    switch w {
    case WeatherBlowingDust, WeatherBlowingSand, WeatherBlowingSnow,
         WeatherDrizzle, WeatherFog, WeatherFreezingFog, WeatherFreezingDrizzle,
         WeatherFreezingRain, WeatherFreezingSpray, WeatherFrost, WeatherHail,
         WeatherHaze, WeatherIceCrystals, WeatherIceFog, WeatherRain,
         WeatherRainShowers, WeatherSleet, WeatherSmoke, WeatherSnow,
         WeatherSnowShowers, WeatherThunderstorms, WeatherVolcanicAsh,
         WeatherWaterSpouts:
        return true
    }
    return false
}

func (c CloudAmount) IsValid() bool {
    switch c {
    case CloudAmountOVC, CloudAmountBKN, CloudAmountSCT,
         CloudAmountFEW, CloudAmountSKC, CloudAmountCLR,
         CloudAmountVV:
        return true
    }
    return false
}
