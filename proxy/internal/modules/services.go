package modules

import (
	"test/internal/infrastructure/component"
	"test/internal/models"
	"test/internal/modules/geo/service"
	"test/internal/provider"
)

type GeoServicer interface {
	AddressSearch(input string) ([]*models.Address, error)
	GeoCode(lat, lng string) ([]*models.Address, error)
}

type Services struct {
	Geo GeoServicer
}

func NewServices() *Services {
	geoServiceProxy := provider.NewGeoServiceProxy()
	cache, _ := component.NewCache("redis:6379")
	geoService := service.NewGeoService(geoServiceProxy, cache)
	return &Services{
		Geo: geoService,
	}
}