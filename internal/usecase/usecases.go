package usecase

import (
	"context"
	"time"

	"github.com/romanovevgeniy/weather-service/internal/domain"
	"github.com/romanovevgeniy/weather-service/internal/ports"
)

type GetLatestReading struct {
	repo ports.ReadingRepository
}

func NewGetLatestReading(repo ports.ReadingRepository) *GetLatestReading {
	return &GetLatestReading{repo: repo}
}

func (u *GetLatestReading) Execute(ctx context.Context, city string) (domain.Reading, error) {
	return u.repo.GetLatestByCity(ctx, city)
}

type IngestWeather struct {
	repo     ports.ReadingRepository
	geo      ports.GeocodingService
	weather  ports.WeatherService
	clock    ports.Clock
	cityName string
}

func NewIngestWeather(repo ports.ReadingRepository, geo ports.GeocodingService, weather ports.WeatherService, clock ports.Clock, city string) *IngestWeather {
	return &IngestWeather{repo: repo, geo: geo, weather: weather, clock: clock, cityName: city}
}

func (u *IngestWeather) Execute(ctx context.Context) error {
	lat, lon, err := u.geo.GetCoords(ctx, u.cityName)
	if err != nil {
		return err
	}

	tempC, tsISO, err := u.weather.GetTemperature(ctx, lat, lon)
	if err != nil {
		return err
	}

	parsedAt, err := time.Parse("2006-01-02T15:04", tsISO)
	if err != nil {
		// fallback to now when parsing fails
		parsedAt = u.clock.Now().UTC()
	}

	return u.repo.Insert(ctx, domain.Reading{
		Name:        u.cityName,
		Timestamp:   parsedAt,
		Temperature: tempC,
	})
}
