package service

import (
	"errors"
	"log"
	"testing"

	"bitbucket.org/eeeeed/arcules-coding/models"
)

type testInvalidZip struct {
	ExternalDataWithGoroutineImpl
}

func (g *testInvalidZip) getGeocodeByZip(zip string) (*models.GeoCodeResults, error) {
	return nil, models.NewWeatherReportError("Invalid zip '"+zip+"'", models.InvalidZipOrAddress)
}

func TestHandlingInvalidZip(t *testing.T) {
	log.Println("\n***** Test 'TestHandlingInvalidZip' started...")
	testGeocodeData := &testInvalidZip{}
	testWeatherServiceHourlyImpl := &WeatherServiceHourlyImpl{testGeocodeData}
	_, err := testWeatherServiceHourlyImpl.GetLastWeekWeatherByZip("00009", 1)
	if err != nil {
		var e *models.WeatherReportError
		if !errors.As(err, &e) {
			// err is a *WeatherReportError, and e is set to the error's value
			t.Errorf("Error instance should be in type 'WeatherReportError'")
		}
	}
	log.Println("Test TestHandlingInvalidZip' finished...")
}

type testCorrecrtFormattedAddress struct {
	ExternalDataWithGoroutineImpl
}

func (g *testCorrecrtFormattedAddress) getGeocodeByZip(zip string) (*models.GeoCodeResults, error) {
	returnGeocodeResults := models.GeoCodeResults{
		Results: []models.GeoCodeResult{
			{
				FormattedAddress: "Irvine, CA 92606, USA",
				Geometry: models.GeoLocation{
					Location: models.GeoLatAndLng{
						Lat: 33.6912614,
						Lng: -117.8223506,
					},
				}},
		},
	}
	return &returnGeocodeResults, nil
}

func TestCorrecrtFormattedAddress(t *testing.T) {
	log.Println("\n***** Test 'TestCorrecrtFormattedAddress' started...")
	testGeocodeData := &testCorrecrtFormattedAddress{}
	testWeatherServiceHourlyImpl := &WeatherServiceHourlyImpl{testGeocodeData}
	weatherResults, err := testWeatherServiceHourlyImpl.GetLastWeekWeatherByZip("92606", 1)
	if err != nil {
		t.Errorf("Error instance should be nill")
	} else {
		if weatherResults.Address != "Irvine, CA 92606, USA" {
			t.Errorf("Address value should be 'Irvine, CA 92606, USA'")
		}
	}
	log.Println("Test 'TestCorrecrtFormattedAddress' finished...")
}
