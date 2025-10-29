package ports

import "context"

type GeocodingService interface {
	GetCoords(ctx context.Context, city string) (lat float64, lon float64, err error)
}

type WeatherService interface {
	GetTemperature(ctx context.Context, lat, lon float64) (temperatureC float64, timestampISO string, err error)
}
