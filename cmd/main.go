package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/Dimix-international/weather-service/db"
	"github.com/Dimix-international/weather-service/internal/client/http/geocoding"
	"github.com/Dimix-international/weather-service/internal/client/http/meteo"
	"github.com/Dimix-international/weather-service/internal/config"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-co-op/gocron/v2"
)

const httpPort = ":3000"

type Reading struct {
	Timestamp   time.Time
	Temperature float64
}

type Storage struct {
	data map[string][]Reading
	mu   sync.RWMutex
}

func main() {
	cfg := config.MustLoadConfig()

	dbCtx, dbCancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer dbCancel()

	_, err := db.NewConnect(dbCtx, &cfg)
	if err != nil {
		fmt.Println(fmt.Sprintf("failed to connect to db: %v", err))
		return
	}

	httClient := &http.Client{Timeout: time.Second * 10}

	storage := &Storage{
		data: make(map[string][]Reading),
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/{city}", func(w http.ResponseWriter, r *http.Request) {
		city := chi.URLParam(r, "city")

		storage.mu.RLock()

		defer storage.mu.RUnlock()

		reading, ok := storage.data[city]
		if !ok {
			w.Write([]byte("not found"))
			return
		}

		raw, err := json.Marshal(reading)
		if err != nil {
			w.Write([]byte("not found"))
			return
		}

		w.Write([]byte(raw))
	})

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func(cfg *config.Config, httpClient *http.Client, storage *Storage) {
		defer wg.Done()
		runCron(cfg, httpClient, storage)
	}(&cfg, httClient, storage)

	go func() {
		defer wg.Done()
		fmt.Println("start server")
		http.ListenAndServe(httpPort, r)
	}()

	wg.Wait()
}

func initWeatherJobs(s gocron.Scheduler, geo geocoding.GeoStore, weather meteo.WeatherStore, storage *Storage) ([]gocron.Job, error) {
	j, err := s.NewJob(
		gocron.DurationJob(
			10*time.Second,
		),
		gocron.NewTask(
			func() {
				var dataWeather meteo.Weather

				geoResp, err := geo.GetCoords("minsk")
				if err != nil {
					return
				}

				if geoResp.Name != "" {
					dataWeather, err = weather.GetWeatherByCoord(geoResp.Latitude, geoResp.Longitude)
				} else {
					dataWeather, err = weather.GetWeatherByCity("minsk")
				}

				if err != nil {
					return
				}

				storage.mu.Lock()

				storage.data["minsk"] = append(storage.data["minsk"], Reading{
					Timestamp:   time.Now().UTC(),
					Temperature: dataWeather.Main.Kelvin - 273.15,
				})

				defer storage.mu.Unlock()

				fmt.Println(dataWeather)
			},
		),
	)
	if err != nil {
		return nil, err
	}

	return []gocron.Job{j}, nil
}

func runCron(cfg *config.Config, httpClient *http.Client, storage *Storage) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return
	}

	geocodingClient := geocoding.NewClient(cfg, httpClient)
	meteoClient := meteo.NewClient(cfg, httpClient)

	jobs, err := initWeatherJobs(s, geocodingClient, meteoClient, storage)
	if err != nil {
		return
	}

	fmt.Printf("Started first job id %v\n", jobs[0].ID())

	s.Start()
}
