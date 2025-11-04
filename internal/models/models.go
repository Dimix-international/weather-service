package models

import (
	"errors"
	"time"
)

const (
	UniqueCodeError = "23505"
)

var (
	ErrCityWeatherNotFound = errors.New("weather city not found")
	ErrCityWeatherExist    = errors.New("weather city already exist")
)

type Weather struct {
	Id        int       `json:"id" db:"id" `
	Name      string    `json:"name" db:"name" `
	Kelvin    float64   `json:"temp" db:"temp"`
	Pressure  int       `json:"pressure" db:"pressure"`
	Humidity  float64   `json:"humidity" db:"humidity"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
