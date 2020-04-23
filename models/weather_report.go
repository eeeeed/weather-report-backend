package models

type WeatherData struct {
	Time                int32   `json:"time"`
	Summary             string  `json:"summary"`
	Icon                string  `json:"icon"`
	PrecipIntensity     float32 `json:"precipIntensity"`
	PrecipProbability   float32 `json:"precipProbability"`
	PrecipType          string  `json:"precipType"`
	Temperature         float32 `json:"temperature"`
	ApparentTemperature float32 `json:"apparentTemperature"`
	DewPoint            float32 `json:"dewPoint"`
	Humidity            float32 `json:"humidity"`
	Pressure            float32 `json:"pressure"`
	WindSpeed           float32 `json:"windSpeed"`
	WindGust            float32 `json:"windGust"`
	WindBearing         float32 `json:"windBearing"`
	CloudCover          float32 `json:"cloudCover"`
	UvIndex             float32 `json:"uvIndex"`
	Visibility          float32 `json:"visibility"`
	Ozone               float32 `json:"ozone"`
	Date                string  `json:"date"`
	DayOfWeek           string  `json:"dayOfWeek"`
}

type WeatherReport struct {
	Currently WeatherData `json:"currently"`
}

type HourlyWeatherData struct {
	Summary string        `json:"summary"`
	Data    []WeatherData `json:"data"`
}

type WeatherHourlyReport struct {
	Hourly HourlyWeatherData `json:"hourly"`
}

type ResponseWeatherData struct {
	Address string        `json:"address"`
	Data    []WeatherData `json:"data"`
}
