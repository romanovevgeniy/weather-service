package ports

import (
	"context"
	"time"

	"github.com/romanovevgeniy/weather-service/internal/domain"
)

type ReadingRepository interface {
	Insert(ctx context.Context, reading domain.Reading) error
	GetLatestByCity(ctx context.Context, city string) (domain.Reading, error)
}

type Clock interface {
	Now() time.Time
}
