package service

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"bitbucket.org/eeeeed/arcules-coding/models"
)

type GeocodeServiceImpl struct {
}

func NewGeocodeService() *GeocodeServiceImpl {
	return &GeocodeServiceImpl{}
}

func (geo *GeocodeServiceImpl) GetGeoLatAndLngByZip(zip string) (*models.GeoCodeResults, error) {

	geocodeEndpoint := os.Getenv("GEOCODE_ENDPOINT")
	geocodeApiKey := os.Getenv("GEOCODE_APIKEY")
	req, err := http.NewRequest("GET", geocodeEndpoint, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	q := req.URL.Query()
	q.Add("key", geocodeApiKey)
	q.Add("address", zip)
	req.URL.RawQuery = q.Encode()

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

	// log.Println("geoCodeResults body: ", string(body))

	geoCodeResults := models.GeoCodeResults{}
	jsonErr := json.Unmarshal(body, &geoCodeResults)
	if jsonErr != nil {
		log.Println(jsonErr)
		return nil, jsonErr
	}

	geoCodeResultsInJson, _ := json.Marshal(geoCodeResults)
	log.Println("geoCodeResults: ", string(geoCodeResultsInJson))

	if len(geoCodeResults.Results) == 0 {
		return nil, models.NewWeatherReportError("Invalid zip '"+zip+"'", models.InvalidZipOrAddress)
	}

	return &geoCodeResults, nil
}
