// Package p contains an HTTP Cloud Function.
package function

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"bitbucket.org/eeeeed/arcules-coding/models"
	"bitbucket.org/eeeeed/arcules-coding/service"
	"bitbucket.org/eeeeed/arcules-coding/service/interfaces"
	writers "bitbucket.org/eeeeed/arcules-coding/util"
	"github.com/joho/godotenv"
	_ "github.com/joho/godotenv/autoload"
)

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func init() {
	// loads values from .env into the system
	if fileExists(".env") {
		log.Print(".env file exists")
	} else {
		log.Print(".env file does not exist (or is a directory)")
	}

	if err := godotenv.Load(); err != nil {
		log.Print("godotenv.Load() failed...")
		log.Print("Go manually load .env")
		err = godotenv.Load(".env")
		if err != nil {
			log.Print("Manually load .env failed... ")
			log.Print("Call os.Setenv() to set env variables... ")
			const GEOCODE_ENDPOINT = "https://maps.googleapis.com/maps/api/geocode/json"
			const GEOCODE_APIKEY = ""
			const WEATHER_ENDPOINT = "https://api.darksky.net/forecast/"
			const WEATHER_APIKEY = ""

			os.Setenv("GEOCODE_ENDPOINT", GEOCODE_ENDPOINT)
			os.Setenv("GEOCODE_APIKEY", GEOCODE_APIKEY)
			os.Setenv("WEATHER_ENDPOINT", WEATHER_ENDPOINT)
			os.Setenv("WEATHER_APIKEY", WEATHER_APIKEY)
		}
	}
}

func WeatherReport(w http.ResponseWriter, r *http.Request) {

	// err := godotenv.Load()
	// if err != nil {
	// 	log.Println("Error loading .env file")
	// 	var e = models.NewWeatherReportError("Something bad happened to our server, please try again later", models.OtherError)
	// 	setErrorCodeAndMessage(w, *e, http.StatusInternalServerError)
	// 	// return
	// }

	w.Header().Set("Content-Type", "application/json")

	// Set CORS headers for the preflight request
	if r.Method == http.MethodOptions {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Max-Age", "3600")
		w.WriteHeader(http.StatusNoContent)
		return
	}
	// Set CORS headers for the main request.
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Parse the request from query string
	query := r.URL.Query()
	zip := query.Get("zip")
	weeksBehind := query.Get("weeksBehind")
	log.Println("zip: ", zip)
	log.Println("weeksBehind: ", weeksBehind)

	// Check if zip is specified
	if zip == "" {
		var e = models.NewWeatherReportError("Parameter 'zip' is missing", models.OtherError)
		setErrorCodeAndMessage(w, *e, http.StatusBadRequest)
		return
	}

	weeksBehindInInt, err := strconv.Atoi(weeksBehind)
	if err != nil || weeksBehindInInt == 0 {
		weeksBehindInInt = 1
	}

	var weatherService interfaces.WeatherService = service.NewWeatherService()
	weatherResults, err := weatherService.GetLastWeekWeatherByZip(zip, weeksBehindInInt)
	if err != nil {
		log.Println(err)
		var e *models.WeatherReportError
		if errors.As(err, &e) {
			// err is a *WeatherReportError, and e is set to the error's value
			setErrorCodeAndMessage(w, *e, http.StatusBadRequest)
		} else {
			e = models.NewWeatherReportError(err.Error(), models.OtherError)
			setErrorCodeAndMessage(w, *e, http.StatusInternalServerError)
		}
		return
	}

	jdw := writers.NewDataWriter(weatherResults)
	weatherResultInJson, err := jdw.JSONString()
	// log.Println("weatherResults: ", string(weatherResultInJson))

	fmt.Fprint(w, string(weatherResultInJson))
}

func setErrorCodeAndMessage(w http.ResponseWriter, weatherErr models.WeatherReportError, errCode int) {
	w.WriteHeader(errCode)
	jw := writers.NewWeatherReportErrorWriter(weatherErr)
	errMessageInJson, err := jw.JSONString()
	log.Println("jsonString: " + errMessageInJson)
	if err != nil {
		w.Write([]byte(err.Error()))
		log.Println("Just return: ")
	} else {
		w.Write([]byte(errMessageInJson))
		log.Println("writing err message: ")
	}
}
