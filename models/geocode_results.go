package models

type GeoLatAndLng struct {
	Lat float32 `json:"lat"`
	Lng float32 `json:"lng"`
}

type GeoLocation struct {
	Location GeoLatAndLng `json:"location"`
}

type GeoCodeResult struct {
	FormattedAddress string `json:"formatted_address"`
	Geometry	GeoLocation `json:"geometry"`
}

type GeoCodeResults struct {
	Results	[]GeoCodeResult `json:"results"`
}