package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/Dimix-international/weather-service/db"
	"github.com/Dimix-international/weather-service/internal/client/http/geocoding"
	"github.com/Dimix-international/weather-service/internal/client/http/meteo"
	"github.com/Dimix-international/weather-service/internal/config"
	"github.com/Dimix-international/weather-service/internal/models"

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

	client, err := db.NewConnect(dbCtx, &cfg)
	if err != nil {
		fmt.Println(fmt.Sprintf("failed to connect to db: %v", err))
		return
	}

	httClient := &http.Client{Timeout: time.Second * 10}

	weatherStorage := db.NewWeatherStorage(client.DB)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/{city}", func(w http.ResponseWriter, r *http.Request) {
		city := chi.URLParam(r, "city")

		weatherData, err := weatherStorage.GetByCityName(r.Context(), city)
		if err != nil {
			w.Write([]byte(err.Error()))
			return
		}

		raw, err := json.Marshal(weatherData)
		if err != nil {
			w.Write([]byte("not found"))
			return
		}

		w.Write([]byte(raw))
	})

	wg := sync.WaitGroup{}
	wg.Add(2)

	go func(cfg *config.Config, httpClient *http.Client, weatherStorage *db.WeatherStorage) {
		defer wg.Done()
		runCron(cfg, httpClient, weatherStorage)
	}(&cfg, httClient, weatherStorage)

	go func() {
		defer wg.Done()
		fmt.Println("start server")
		http.ListenAndServe(httpPort, r)
	}()

	wg.Wait()
}

func initWeatherJobs(s gocron.Scheduler, geo geocoding.GeoStore, weather meteo.WeatherStore, weatherStorage *db.WeatherStorage) ([]gocron.Job, error) {
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

				modelWeaher := &models.Weather{
					Name:     "minsk",
					Kelvin:   dataWeather.Main.Kelvin,
					Pressure: dataWeather.Main.Pressure,
					Humidity: dataWeather.Main.Humidity,
				}

				ctx := context.Background()

				err = weatherStorage.Update(ctx, modelWeaher)
				if err != nil {
					if errors.Is(err, models.ErrCityWeatherNotFound) {
						log.Print("create weather for city")
						weatherStorage.Create(ctx, modelWeaher)
					}
				}

				log.Print("update weather for city")
			},
		),
	)
	if err != nil {
		return nil, err
	}

	return []gocron.Job{j}, nil
}

func runCron(cfg *config.Config, httpClient *http.Client, weatherStorage *db.WeatherStorage) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return
	}

	geocodingClient := geocoding.NewClient(cfg, httpClient)
	meteoClient := meteo.NewClient(cfg, httpClient)

	jobs, err := initWeatherJobs(s, geocodingClient, meteoClient, weatherStorage)
	if err != nil {
		return
	}

	fmt.Printf("Started first job id %v\n", jobs[0].ID())

	s.Start()
}
