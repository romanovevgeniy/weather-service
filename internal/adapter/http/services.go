package httpadapter

import (
	"context"

	geocoding "github.com/romanovevgeniy/weather-service/internal/client/http/geocoding"
	openmeteo "github.com/romanovevgeniy/weather-service/internal/client/http/open_meteo"
	"github.com/romanovevgeniy/weather-service/internal/ports"
)

type GeocodingAdapter struct{ c geocoder }

type geocoder interface {
	GetCoords(ctx context.Context, city string) (geocoding.Response, error)
}

func NewGeocodingAdapter(c geocoder) ports.GeocodingService { return &GeocodingAdapter{c: c} }

func (g *GeocodingAdapter) GetCoords(ctx context.Context, city string) (float64, float64, error) {
	res, err := g.c.GetCoords(ctx, city)
	if err != nil {
		return 0, 0, err
	}
	return res.Latitude, res.Longitude, nil
}

type WeatherAdapter struct{ c weather }

type weather interface {
	GetTemperature(ctx context.Context, lat, long float64) (openmeteo.Response, error)
}

func NewWeatherAdapter(c weather) ports.WeatherService { return &WeatherAdapter{c: c} }

func (w *WeatherAdapter) GetTemperature(ctx context.Context, lat, lon float64) (float64, string, error) {
	res, err := w.c.GetTemperature(ctx, lat, lon)
	if err != nil {
		return 0, "", err
	}
	return res.Current.Temperature2m, res.Current.Time, nil
}
