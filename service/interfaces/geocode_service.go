package interfaces

import "bitbucket.org/eeeeed/arcules-coding/models"

type GeocodeService interface {
	GetGeoLatAndLngByZip(zip string) (*models.GeoCodeResults, error)
}
