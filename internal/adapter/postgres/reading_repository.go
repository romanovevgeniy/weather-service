package postgres

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/romanovevgeniy/weather-service/internal/domain"
	"github.com/romanovevgeniy/weather-service/internal/ports"
)

type ReadingRepository struct {
	conn *pgx.Conn
	clk  ports.Clock
}

func NewReadingRepository(conn *pgx.Conn, clk ports.Clock) *ReadingRepository {
	return &ReadingRepository{conn: conn, clk: clk}
}

func (r *ReadingRepository) Insert(ctx context.Context, reading domain.Reading) error {
	_, err := r.conn.Exec(
		ctx,
		"INSERT INTO reading (name, temperature, timestamp) VALUES ($1, $2, $3)",
		reading.Name, reading.Temperature, reading.Timestamp,
	)
	return err
}

func (r *ReadingRepository) GetLatestByCity(ctx context.Context, city string) (domain.Reading, error) {
	var out domain.Reading
	err := r.conn.QueryRow(
		ctx,
		"SELECT name, timestamp, temperature FROM reading WHERE name = $1 ORDER BY timestamp DESC LIMIT 1",
		city,
	).Scan(&out.Name, &out.Timestamp, &out.Temperature)
	return out, err
}
