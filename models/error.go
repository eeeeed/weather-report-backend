package models

type ErrorCode int

const (
	OtherError ErrorCode = iota
	InvalidZipOrAddress
)

type WeatherReportError struct {
	err  string    //error description
	code ErrorCode // error code
}

func NewWeatherReportError(err string, code ErrorCode) *WeatherReportError {
	return &WeatherReportError{err, code}
}

func (e *WeatherReportError) Error() string {
	return e.err
}

func (e *WeatherReportError) Code() ErrorCode {
	return e.code
}
