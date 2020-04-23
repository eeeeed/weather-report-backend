package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"bitbucket.org/eeeeed/arcules-coding/models"
	"bitbucket.org/eeeeed/arcules-coding/service/interfaces"
)

const numOfdaysPerWeek = 7
const excludeParameters = "currently,daily,flags"
const layoutISO = "2006-01-02"

// WeatherServiceImpl is...
type WeatherServiceImpl struct {
}

// WeatherServiceHourlyImpl is...
type WeatherServiceHourlyImpl struct {
	ExternalData
}

// ExternalData is...
type ExternalData interface {
	// this function will be promoted to the parent struct level
	getGeocodeByZip(zip string) (*models.GeoCodeResults, error)
	getWeatherReportByWeekTime(weatherEndpointWithParameters string, weekTime time.Time) (*models.ResponseWeatherData, error)
}

// ExternalDataImpl is...
type ExternalDataImpl struct {
}

// ExternalDataWithGoroutineImpl is...
type ExternalDataWithGoroutineImpl struct {
	ExternalDataImpl
}

// GetGeocodeByZip is...
func (e *ExternalDataImpl) getGeocodeByZip(zip string) (*models.GeoCodeResults, error) {
	var geo interfaces.GeocodeService = NewGeocodeService()
	geoResults, geoErr := geo.GetGeoLatAndLngByZip(zip)
	if geoErr != nil {
		log.Println(geoErr)
		return nil, geoErr
	}
	return geoResults, nil
}

func goFetchWeatherData(wg *sync.WaitGroup, resultsCh chan *models.WeatherData, weatherEndpointWithTimestamp string, excludeParameters string, weekTime time.Time) {
	weatherData, err := getWeatherReportByTimestamp(weatherEndpointWithTimestamp, excludeParameters)
	if err != nil {
		log.Println(err)
		wg.Done()
	}
	weatherData.Date = weekTime.Format(layoutISO)
	weatherData.DayOfWeek = weekTime.Weekday().String()
	resultsCh <- weatherData
	wg.Done()
}

func processResultWeatherData(responseWeatherDatas *models.ResponseWeatherData, resultsCh chan *models.WeatherData, doneCh chan bool) {
	for resultWeatherData := range resultsCh {
		responseWeatherDatas.Data = append(responseWeatherDatas.Data, *resultWeatherData)
	}
	sort.SliceStable(responseWeatherDatas.Data, func(i, j int) bool {
		return responseWeatherDatas.Data[i].Date < responseWeatherDatas.Data[j].Date
	})
	doneCh <- true
}

// GetWeatherReportByWeekTime is...
func (e *ExternalDataWithGoroutineImpl) getWeatherReportByWeekTime(weatherEndpointWithParameters string, weekTime time.Time) (*models.ResponseWeatherData, error) {

	responseWeatherDatas := models.ResponseWeatherData{}

	// A 'done' channel to signal the function is done
	done := make(chan bool)

	// A 'results' channel to collect the result data
	results := make(chan *models.WeatherData)

	// A function to process result weather data
	go processResultWeatherData(&responseWeatherDatas, results, done)

	var wg sync.WaitGroup
	for count := 0; count < numOfdaysPerWeek; count++ {
		weekEarlierTimeStamp := weekTime.Unix()
		timestamp := strconv.FormatInt(weekEarlierTimeStamp, 10)
		weatherEndpointWithTimestamp := weatherEndpointWithParameters + "," + timestamp
		log.Println("weekEarlier: ", weekTime, ", weekEarlierTimeStamp: ", weekEarlierTimeStamp, "timestamp: ", timestamp)
		wg.Add(1)
		go goFetchWeatherData(&wg, results, weatherEndpointWithTimestamp, excludeParameters, weekTime)
		weekTime = weekTime.AddDate(0, 0, 1)
	}
	wg.Wait()
	close(results)

	<-done
	return &responseWeatherDatas, nil
}

// GetWeatherReportByWeekTime is...
func (e *ExternalDataImpl) getWeatherReportByWeekTime(weatherEndpointWithParameters string, weekTime time.Time) (*models.ResponseWeatherData, error) {

	responseWeatherDatas := models.ResponseWeatherData{}

	for count := 0; count < numOfdaysPerWeek; count++ {
		weekEarlierTimeStamp := weekTime.Unix()
		timestamp := strconv.FormatInt(weekEarlierTimeStamp, 10)
		weatherEndpointWithTimestamp := weatherEndpointWithParameters + "," + timestamp
		log.Println("weekEarlier: ", weekTime, ", weekEarlierTimeStamp: ", weekEarlierTimeStamp, "timestamp: ", timestamp)
		log.Println("weatherEndpointWithTimestamp: ", weatherEndpointWithTimestamp)
		weatherData, err := getWeatherReportByTimestamp(weatherEndpointWithTimestamp, excludeParameters)
		if err != nil {
			log.Println(err)
			continue
		}
		weatherData.Date = weekTime.Format(layoutISO)
		weatherData.DayOfWeek = weekTime.Weekday().String()
		responseWeatherDatas.Data = append(responseWeatherDatas.Data, *weatherData)
		weekTime = weekTime.AddDate(0, 0, 1)
	}
	return &responseWeatherDatas, nil
}

func NewWeatherService() *WeatherServiceHourlyImpl {
	externalData := &ExternalDataWithGoroutineImpl{}
	return &WeatherServiceHourlyImpl{externalData}
}

func (w *WeatherServiceHourlyImpl) GetLastWeekWeatherByZip(zip string, numOfWeeksBehind int) (*models.ResponseWeatherData, error) {

	geoResults, geoErr := w.getGeocodeByZip(zip)
	if geoErr != nil {
		log.Println(geoErr)
		return nil, geoErr
	}

	formattedAddress := geoResults.Results[0].FormattedAddress
	lat := geoResults.Results[0].Geometry.Location.Lat
	lng := geoResults.Results[0].Geometry.Location.Lng

	currentTime := time.Now()
	year, month, day := currentTime.Date()

	// Set time to noon
	currentTime = time.Date(year, month, day, 12, 0, 0, 0, time.Local)

	// Move back to Monday
	for currentTime.Weekday() != time.Sunday {
		currentTime = currentTime.AddDate(0, 0, -1)
	}

	// Move 1 week earlier
	weekEarlier := currentTime.AddDate(0, 0, -(numOfdaysPerWeek * numOfWeeksBehind))

	weatherEndpoint := os.Getenv("WEATHER_ENDPOINT")
	weatherApiKey := os.Getenv("WEATHER_APIKEY")

	latAndLngHolder := []string{fmt.Sprintf("%g", lat), fmt.Sprintf("%g", lng)}
	latAndLng := strings.Join(latAndLngHolder, ",")

	weatherEndpointWithParameters := weatherEndpoint + weatherApiKey + "/" + latAndLng

	responseWeatherDatas, weatherReportErr := w.getWeatherReportByWeekTime(weatherEndpointWithParameters, weekEarlier)
	if weatherReportErr != nil {
		log.Println(weatherReportErr)
		return nil, weatherReportErr
	}
	responseWeatherDatas.Address = formattedAddress

	return responseWeatherDatas, nil
}

func (w *WeatherServiceImpl) GetLastWeekWeatherByZip(zip string, numOfWeeksBehind int) (*models.ResponseWeatherData, error) {

	// get lat and lng from zip
	var geo interfaces.GeocodeService = NewGeocodeService()
	geoResults, geoErr := geo.GetGeoLatAndLngByZip(zip)
	if geoErr != nil {
		log.Println(geoErr)
		return nil, geoErr
	}

	formattedAddress := geoResults.Results[0].FormattedAddress
	lat := geoResults.Results[0].Geometry.Location.Lat
	lng := geoResults.Results[0].Geometry.Location.Lng

	currentTime := time.Now()
	year, month, day := currentTime.Date()

	// Set time to noon
	currentTime = time.Date(year, month, day, 12, 0, 0, 0, time.Local)

	// Move back to Monday
	for currentTime.Weekday() != time.Sunday {
		currentTime = currentTime.AddDate(0, 0, -1)
	}

	// Move 1 week earlier
	weekEarlier := currentTime.AddDate(0, 0, -(numOfdaysPerWeek * numOfWeeksBehind))

	weatherEndpoint := os.Getenv("WEATHER_ENDPOINT")
	weatherApiKey := os.Getenv("WEATHER_APIKEY")

	latAndLngHolder := []string{fmt.Sprintf("%g", lat), fmt.Sprintf("%g", lng)}
	latAndLng := strings.Join(latAndLngHolder, ",")

	weatherEndpointWithParameters := weatherEndpoint + weatherApiKey + "/" + latAndLng

	responseWeatherDatas := models.ResponseWeatherData{}
	responseWeatherDatas.Address = formattedAddress

	for count := 0; count < numOfdaysPerWeek; count++ {
		weekEarlierTimeStamp := weekEarlier.Unix()
		timestamp := strconv.FormatInt(weekEarlierTimeStamp, 10)
		weatherEndpointWithTimestamp := weatherEndpointWithParameters + "," + timestamp
		log.Println("weekEarlier: ", weekEarlier, ", weekEarlierTimeStamp: ", weekEarlierTimeStamp, "timestamp: ", timestamp)
		// log.Println("weatherEndpointWithTimestamp: ", weatherEndpointWithTimestamp)
		weatherData, err := getWeatherReportByTimestamp(weatherEndpointWithTimestamp, "")
		if err != nil {
			log.Println(err)
			continue
		}
		weatherData.Date = weekEarlier.Format(layoutISO)
		weatherData.DayOfWeek = weekEarlier.Weekday().String()
		responseWeatherDatas.Data = append(responseWeatherDatas.Data, *weatherData)
		weekEarlier = weekEarlier.AddDate(0, 0, 1)
	}

	return &responseWeatherDatas, nil
}

func getWeatherReportByTimestamp(weatherEndpointWithParameters string, exclude string) (*models.WeatherData, error) {

	req, err := http.NewRequest("GET", weatherEndpointWithParameters, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if exclude != "" {
		q := req.URL.Query()
		q.Add("exclude", exclude)
		req.URL.RawQuery = q.Encode()
	}

	log.Println(req.URL.String())

	client := &http.Client{}

	res, getErr := client.Do(req)
	if getErr != nil {
		log.Println(getErr)
		return nil, getErr
	}

	defer res.Body.Close()

	body, readErr := ioutil.ReadAll(res.Body)
	if readErr != nil {
		log.Println(readErr)
		return nil, readErr
	}

	if exclude == "" {
		weatherReport := models.WeatherReport{}
		jsonErr := json.Unmarshal(body, &weatherReport)
		if jsonErr != nil {
			log.Println(jsonErr)
			return nil, jsonErr
		}
		return &weatherReport.Currently, nil
	} else {
		weatherHourlyReport := models.WeatherHourlyReport{}
		jsonErr := json.Unmarshal(body, &weatherHourlyReport)
		if jsonErr != nil {
			log.Println(jsonErr)
			return nil, jsonErr
		}
		dataCount := len(weatherHourlyReport.Hourly.Data)

		if dataCount == 0 {
			return nil, errors.New("There is no hourly weather data")
		}
		// Get weather data at noon
		hourlyData := weatherHourlyReport.Hourly.Data[dataCount/2]
		hourlyData.Summary = weatherHourlyReport.Hourly.Summary
		log.Println("Noon timestamp: ", hourlyData.Time)
		return &hourlyData, nil
	}

}
