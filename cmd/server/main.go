package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-co-op/gocron/v2"
	"github.com/romanovevgeniy/weather-service/internal/client/http/geocoding"
	"github.com/romanovevgeniy/weather-service/internal/client/http/open_meteo"
)

const HttpPort = ":3000"

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	httpClient := &http.Client{Timeout: 10 * time.Second}
	geoCodingClient := geocoding.NewClient(*httpClient)
	openMeteoClient := open_meteo.NewClient(*httpClient)

	r.Get("/{city}", func(w http.ResponseWriter, r *http.Request) {

		city := chi.URLParam(r, "city")
		fmt.Printf("Requested city: %s\n", city)

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

		raw, err := json.Marshal(openMetRes)
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

	jobs, err := initJobs(s)
	if err != nil {
		panic(err)
	}

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		fmt.Println("Server starting on port " + HttpPort)
		err := http.ListenAndServe(HttpPort, r)
		if err != nil {
			panic(err)
		}
	}()

	go func() {
		defer wg.Done()
		fmt.Printf("Starting job: %v", jobs[0].ID())
		s.Start()
	}()

	wg.Wait()
}

func initJobs(sheduler gocron.Scheduler) ([]gocron.Job, error) {

	j, err := sheduler.NewJob(
		gocron.DurationJob(
			10*time.Second,
		),
		gocron.NewTask(
			func() {
				fmt.Println("Hello")
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
