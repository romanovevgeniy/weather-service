package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-co-op/gocron/v2"
	"github.com/jackc/pgx/v5"
	"github.com/romanovevgeniy/weather-service/internal/client/http/geocoding"
	"github.com/romanovevgeniy/weather-service/internal/client/http/open_meteo"
)

const (
	HttpPort = ":3000"
	city     = "moscow"
)

type Reading struct {
	Name        string    `db:"name"`
	Timestamp   time.Time `db:"timestamp"`
	Temperature float64   `db:"temperature"`
}

type Storage struct {
	data map[string][]Reading
	mu   sync.RWMutex
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	ctx := context.Background()

	conn, err := pgx.Connect(ctx, "postgresql://admin:password@localhost:54321/weather")
	if err != nil {
		panic(err)
	}
	defer conn.Close(ctx)

	r.Get("/{city}", func(w http.ResponseWriter, r *http.Request) {

		cityName := chi.URLParam(r, "city")
		fmt.Printf("Requested city: %s\n", cityName)

		var reading Reading

		err = conn.QueryRow(
			ctx,
			"SELECT name, timestamp, temperature FROM reading WHERE name = $1 ORDER BY timestamp DESC LIMIT 1", city,
		).Scan(&reading.Name, &reading.Timestamp, &reading.Temperature)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("internal error"))
		}

		var raw []byte
		raw, err := json.Marshal(reading)
		if err != nil {
			log.Println(err)
		}

		_, err = w.Write(raw)
		if err != nil {
			log.Println(err)
		}
	})

	s, err := gocron.NewScheduler()
	if err != nil {
		panic(err)
	}

	jobs, err := initJobs(ctx, s, conn)
	if err != nil {
		panic(err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		fmt.Println("Server starting on port" + HttpPort)
		err := http.ListenAndServe(HttpPort, r)
		if err != nil {
			panic(err)
		}
	}()

	go func() {
		defer wg.Done()
		fmt.Printf("Starting job: %v\n", jobs[0].ID())
		s.Start()
	}()

	wg.Wait()
}

func initJobs(ctx context.Context, sheduler gocron.Scheduler, conn *pgx.Conn) ([]gocron.Job, error) {

	httpClient := &http.Client{Timeout: 10 * time.Second}
	geoCodingClient := geocoding.NewClient(*httpClient)
	openMeteoClient := open_meteo.NewClient(*httpClient)

	j, err := sheduler.NewJob(
		gocron.DurationJob(
			10*time.Second,
		),
		gocron.NewTask(
			func() {
				geoRes, err := geoCodingClient.GetCoords(city)
				if err != nil {
					log.Println(err)
					return
				}

				openMetRes, err := openMeteoClient.GetTemperature(geoRes.Latitude, geoRes.Longitude)
				if err != nil {
					log.Println(err)
					return
				}

				timestamp, err := time.Parse("2006-01-02T15:04", openMetRes.Current.Time)

				_, err = conn.Exec(
					ctx,
					"INSERT INTO reading (name, temperature, timestamp) VALUES ($1, $2, $3)",
					city, openMetRes.Current.Temperature2m, timestamp,
				)
				if err != nil {
					log.Println(err)
					return
				}

				fmt.Printf("%v Update data for city: %s\n", time.Now(), city)
			},
		),
	)
	if err != nil {
		return nil, err
	}

	return []gocron.Job{j}, nil
}

func runCron() {

	select {
	case <-time.After(time.Minute):
	}

	//err = s.Shutdown()
	//if err != nil {
	//
	//}
}
