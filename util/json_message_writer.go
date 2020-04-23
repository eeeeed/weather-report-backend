package writers

import (
	"encoding/json"
	"strconv"

	"bitbucket.org/eeeeed/arcules-coding/models"
)

type JSONErrorWriter struct {
	Error models.WeatherReportError `json:"error"`
}

type JSONDataWriter struct {
	Data interface{} `json:"data"`
}

func NewWeatherReportErrorWriter(err models.WeatherReportError) *JSONErrorWriter {
	return &JSONErrorWriter{
		Error: err,
	}
}

func NewDataWriter(data interface{}) *JSONDataWriter {
	return &JSONDataWriter{
		Data: data,
	}
}

func (jew *JSONErrorWriter) JSONString() (string, error) {
	messageResponse := map[string]interface{}{
		"data": map[string]string{
			"code":    strconv.Itoa(int(jew.Error.Code())),
			"message": jew.Error.Error(),
		},
	}
	bytesValue, err := json.Marshal(messageResponse)
	if err != nil {
		return "", err
	}
	return string(bytesValue), nil
}

func (jdw *JSONDataWriter) JSONString() (string, error) {
	dataResponse := map[string]interface{}{
		"data": jdw.Data,
	}
	bytesValue, err := json.Marshal(dataResponse)
	if err != nil {
		return "", err
	}
	return string(bytesValue), nil
}
