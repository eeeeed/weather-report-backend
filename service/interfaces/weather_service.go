package interfaces

import "bitbucket.org/eeeeed/arcules-coding/models"

type WeatherService interface {
	GetLastWeekWeatherByZip(zip string, numOfWeeksBehind int) (*models.ResponseWeatherData, error)
}
