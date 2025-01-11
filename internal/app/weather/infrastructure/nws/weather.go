package nws

// WeatherIntensity represents the intensity of weather conditions
type WeatherIntensity string

const (
	WeatherIntensityVeryLight WeatherIntensity = "very_light"
	WeatherIntensityLight     WeatherIntensity = "light"
	WeatherIntensityModerate  WeatherIntensity = "moderate"
	WeatherIntensityHeavy     WeatherIntensity = "heavy"
)

// WeatherModifier represents weather condition modifiers
type WeatherModifier string

const (
	ModifierPatches     WeatherModifier = "patches"
	ModifierBlowing     WeatherModifier = "blowing"
	ModifierLowDrifting WeatherModifier = "low_drifting"
	ModifierFreezing    WeatherModifier = "freezing"
	ModifierShallow     WeatherModifier = "shallow"
	ModifierPartial     WeatherModifier = "partial"
	ModifierShowers     WeatherModifier = "showers"
)

// WeatherType represents the type of weather condition
type WeatherType string

const (
	WeatherBlowingDust     WeatherType = "blowing_dust"
	WeatherBlowingSand     WeatherType = "blowing_sand"
	WeatherBlowingSnow     WeatherType = "blowing_snow"
	WeatherDrizzle         WeatherType = "drizzle"
	WeatherFog             WeatherType = "fog"
	WeatherFreezingFog     WeatherType = "freezing_fog"
	WeatherFreezingDrizzle WeatherType = "freezing_drizzle"
	WeatherFreezingRain    WeatherType = "freezing_rain"
	WeatherFreezingSpray   WeatherType = "freezing_spray"
	WeatherFrost           WeatherType = "frost"
	WeatherHail            WeatherType = "hail"
	WeatherHaze            WeatherType = "haze"
	WeatherIceCrystals     WeatherType = "ice_crystals"
	WeatherIceFog          WeatherType = "ice_fog"
	WeatherRain            WeatherType = "rain"
	WeatherRainShowers     WeatherType = "rain_showers"
	WeatherSleet           WeatherType = "sleet"
	WeatherSmoke           WeatherType = "smoke"
	WeatherSnow            WeatherType = "snow"
	WeatherSnowShowers     WeatherType = "snow_showers"
	WeatherThunderstorms   WeatherType = "thunderstorms"
	WeatherVolcanicAsh     WeatherType = "volcanic_ash"
	WeatherWaterSpouts     WeatherType = "water_spouts"
)

// CloudAmount represents the amount of cloud coverage
type CloudAmount string

const (
	CloudAmountOVC CloudAmount = "OVC" // Overcast
	CloudAmountBKN CloudAmount = "BKN" // Broken
	CloudAmountSCT CloudAmount = "SCT" // Scattered
	CloudAmountFEW CloudAmount = "FEW" // Few
	CloudAmountSKC CloudAmount = "SKC" // Sky Clear
	CloudAmountCLR CloudAmount = "CLR" // Clear
	CloudAmountVV  CloudAmount = "VV"  // Vertical Visibility
)
