package config

import (
	"fmt"
	"log"
	"sync"

	"github.com/caarlos0/env/v10"
	"github.com/joho/godotenv"
)

var (
	cfg  Config
	once sync.Once
)

type Config struct {
	Port           string `env:"APP_PORT"`
	OpenWeatherKey string `env:"OPEN_WEATHER_KEY"`
	OpenWeatherURL string `env:"OPEN_WEATHER_URL"`
	GeoURL         string `env:"GEO_URL"`
	Database       DB
}

type DB struct {
	URI  string `env:"PG_URI,required,notEmpty"`
	Name string `env:"PG_NAME,required,notEmpty"`
}

func MustLoadConfig() Config {
	if err := godotenv.Load(); err != nil {
		log.Fatal("error loading .env file")
	}

	once.Do(func() {
		if err := env.Parse(&cfg); err != nil {
			fmt.Printf("%+v\n", err)
		}
	})

	return cfg
}
