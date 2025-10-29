package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/jackc/pgx/v5"
	httpadapter "github.com/romanovevgeniy/weather-service/internal/adapter/http"
	pgadapter "github.com/romanovevgeniy/weather-service/internal/adapter/postgres"
	"github.com/romanovevgeniy/weather-service/internal/client/http/geocoding"
	"github.com/romanovevgeniy/weather-service/internal/client/http/open_meteo"
	"github.com/romanovevgeniy/weather-service/internal/config"
	httpdelivery "github.com/romanovevgeniy/weather-service/internal/delivery/http"
	"github.com/romanovevgeniy/weather-service/internal/usecase"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	conn, err := pgx.Connect(ctx, cfg.PGConnString())
	if err != nil {
		log.Fatalf("db connect error: %v", err)
	}
	defer conn.Close(ctx)

	if err := ensureSchema(ctx, conn); err != nil {
		log.Fatalf("ensure schema error: %v", err)
	}

	repo := pgadapter.NewReadingRepository(conn, usecase.SystemClock{})

	httpClient := &http.Client{Timeout: 10 * time.Second}
	geoClient := geocoding.NewClient(*httpClient)
	weatherClient := open_meteo.NewClient(*httpClient)

	geoSvc := httpadapter.NewGeocodingAdapter(geoClient)
	weatherSvc := httpadapter.NewWeatherAdapter(weatherClient)

	getLatest := usecase.NewGetLatestReading(repo)
	ingester := usecase.NewIngestWeather(repo, geoSvc, weatherSvc, usecase.SystemClock{}, cfg.DefaultCity)

	srv := httpdelivery.NewServer(getLatest)

	scheduler, err := gocron.NewScheduler()
	if err != nil {
		log.Fatalf("scheduler error: %v", err)
	}

	if _, err := initJobs(ctx, scheduler, ingester, cfg.JobInterval); err != nil {
		log.Fatalf("init jobs error: %v", err)
	}

	httpSrv := &http.Server{
		Addr:              ":" + cfg.HTTPPort,
		Handler:           srv.Router(),
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("Server listening on :%s", cfg.HTTPPort)
		if err := httpSrv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http serve error: %v", err)
		}
	}()

	go scheduler.Start()

	<-ctx.Done()
	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelShutdown()
	_ = httpSrv.Shutdown(shutdownCtx)
}

func initJobs(ctx context.Context, scheduler gocron.Scheduler, ingester *usecase.IngestWeather, interval time.Duration) ([]gocron.Job, error) {
	j, err := scheduler.NewJob(
		gocron.DurationJob(interval),
		gocron.NewTask(func() {
			if err := ingester.Execute(ctx); err != nil {
				log.Println("ingest error:", err)
			} else {
				log.Printf("%v updated data for city", time.Now())
			}
		}),
	)
	if err != nil {
		return nil, err
	}
	return []gocron.Job{j}, nil
}

func ensureSchema(ctx context.Context, conn *pgx.Conn) error {
	_, err := conn.Exec(ctx, `
CREATE TABLE IF NOT EXISTS reading (
  name text NOT NULL,
  timestamp timestamptz NOT NULL,
  temperature double precision NOT NULL
);
CREATE INDEX IF NOT EXISTS reading_name_timestamp_idx
  ON reading (name, timestamp DESC);
`)
	return err
}
