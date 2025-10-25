package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Dimix-international/weather-service/internal/client/http/geocoding"
	"github.com/Dimix-international/weather-service/internal/client/http/meteo"
	"github.com/Dimix-international/weather-service/internal/config"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-co-op/gocron/v2"
)

const httpPort = ":3000"

func main() {
	cfg := config.MustLoadConfig()

	httClient := &http.Client{Timeout: time.Second * 10}

	geocodingClient := geocoding.NewClient(&cfg, httClient)
	meteoClient := meteo.NewClient(&cfg, httClient)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/{city}", func(w http.ResponseWriter, r *http.Request) {
		var err error
		city := chi.URLParam(r, "city")

		resp, err := geocodingClient.GetCoords(city)
		if err != nil {
			w.Write([]byte("Error"))
			return
		}

		var dataWeather meteo.Weather

		if resp.Name != "" {
			dataWeather, err = meteoClient.GetWeatherByCoord(resp.Latitude, resp.Longitude)
		} else {
			dataWeather, err = meteoClient.GetWeatherByCity(city)
		}

		if err != nil {
			w.Write([]byte("Error"))
			return
		}

		weather, err := json.Marshal(dataWeather)
		if err != nil {
			w.Write([]byte("Error"))
			return
		}

		w.Write([]byte(weather))
	})

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		runCron()
	}()

	go func() {
		defer wg.Done()
		fmt.Println("start server")
		http.ListenAndServe(httpPort, r)
	}()

	wg.Wait()
}

func initJobs(s gocron.Scheduler) ([]gocron.Job, error) {
	j, err := s.NewJob(
		gocron.DurationJob(
			10*time.Second,
		),
		gocron.NewTask(
			func() {
				fmt.Println("hello")
			},
		),
	)
	if err != nil {
		return nil, err
	}

	return []gocron.Job{j}, nil
}

func runCron() {
	s, err := gocron.NewScheduler()
	if err != nil {
		return
	}

	jobs, err := initJobs(s)
	if err != nil {
		return
	}

	fmt.Printf("Started first job id %v\n", jobs[0].ID())

	s.Start()
}
