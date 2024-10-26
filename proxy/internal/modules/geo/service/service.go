package service

import (
	"context"
	"encoding/json"
	"test/internal/infrastructure/component"
	"test/internal/models"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var GetCacheDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "get_cache_duration",
		Help:    "Get cache duration in seconds",
		Buckets: []float64{0.1, 0.5, 1, 2, 5},
	},
	[]string{"endpoint"},
)

var GetExternalApiDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Name:    "get_external_api_duration",
		Help:    "Get external API duration in seconds",
		Buckets: []float64{0.1, 0.5, 1, 2, 5},
	},
	[]string{"endpoint"},
)

type GeoProvider interface {
	AddressSearch(input string) ([]*models.Address, error)
	GeoCode(lat, lon string) ([]*models.Address, error)
}

type GeoService struct {
	geoProvider GeoProvider
	cache       *component.Cache
}

type GeoServicer interface {
	AddressSearch(input string) ([]*models.Address, error)
	GeoCode(lat, lng string) ([]*models.Address, error)
}

func NewGeoService(geoProvider GeoProvider, cache *component.Cache) GeoServicer {
	/* prometheus.MustRegister(getCacheDuration)
	prometheus.MustRegister(getExternalApiDuration) */

	return &GeoService{
		geoProvider: geoProvider,
		cache:       cache,
	}
}

func (g *GeoService) GeoCode(lat, lng string) ([]*models.Address, error) {
	startTime := time.Now()

	a := "geo_code:"+lat+lng
	jsonCashe, err := g.cache.Get(context.Background(), a)
	var addresses = make([]*models.Address, 0)
	json.Unmarshal([]byte(jsonCashe.(string)), &addresses)

	duration := time.Since(startTime).Seconds()
	GetCacheDuration.WithLabelValues("geo_code").Observe(duration)

	if err == nil {
		return addresses, nil
	}

	if err.Error() == "redis: nil" {
		startTime = time.Now()

		addresses, err = g.geoProvider.GeoCode(lat, lng)

		duration = time.Since(startTime).Seconds()
		GetExternalApiDuration.WithLabelValues("geo_code").Observe(duration)

		if err != nil {
			return nil, err
		}

		err = g.cache.Set(context.Background(), "geo_code:"+lat+lng, addresses)
		if err != nil {
			return nil, err
		}

	} else {
		return nil, err
	}

	return addresses, nil
}

func (g *GeoService) AddressSearch(input string) ([]*models.Address, error) {
	startTime := time.Now()

	jsonCashe, err := g.cache.Get(context.Background(), "address_search:"+input)
	var addresses []*models.Address
	json.Unmarshal([]byte(jsonCashe.(string)), &addresses)

	duration := time.Since(startTime).Seconds()
	GetCacheDuration.WithLabelValues("address_search").Observe(duration)

	if err == nil {
		return addresses, nil
	}

	if err.Error() == "redis: nil" {
		startTime = time.Now()

		addresses, err = g.geoProvider.AddressSearch(input)

		duration = time.Since(startTime).Seconds()
		GetExternalApiDuration.WithLabelValues("address_search").Observe(duration)
		if err != nil {
			return nil, err
		}

		err = g.cache.Set(context.Background(), "address_search:"+input, addresses)
		if err != nil {
			return nil, err
		}

	} else {
		return nil, err
	}

	return addresses, nil
}
