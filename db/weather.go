package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/Dimix-international/weather-service/internal/models"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type WeatherStorage struct {
	db *sqlx.DB
}

const (
	usersEmailKeyConstraint = "reading_name_key"
)

func NewWeatherStorage(db *sqlx.DB) *WeatherStorage {
	return &WeatherStorage{db: db}
}

func (w *WeatherStorage) GetByCityName(ctx context.Context, city string) (models.Weather, error) {
	var weather models.Weather

	if err := w.db.QueryRowxContext(
		ctx,
		`SELECT * FROM reading WHERE upper(name)=$1;`,
		strings.ToUpper(city),
	).StructScan(&weather); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Weather{}, errors.New("weather not found")
		}
		return models.Weather{}, err
	}

	return weather, nil
}

func (w *WeatherStorage) Create(ctx context.Context, data *models.Weather) error {
	if _, err := w.db.ExecContext(
		ctx,
		`INSERT INTO reading (name, temp, pressure, humidity)
			VALUES($1, $2, $3, $4);`,
		data.Name,
		data.Kelvin,
		data.Pressure,
		data.Humidity,
	); err != nil {
		pqErr, ok := err.(*pq.Error)
		if ok && pqErr.Code == models.UniqueCodeError && pqErr.Constraint == usersEmailKeyConstraint {
			return models.ErrCityWeatherExist
		}
		return err
	}

	return nil
}

func (w *WeatherStorage) Update(ctx context.Context, data *models.Weather) error {
	res, err := w.db.ExecContext(
		ctx,
		`UPDATE reading SET 
			temp=$1, 
			pressure=$2, 
			humidity=$3
		WHERE upper(name)=$4;`,
		data.Kelvin,
		data.Pressure,
		data.Humidity,
		strings.ToUpper(data.Name),
	)
	if err != nil {
		fmt.Println("res update", err)
		if errors.Is(err, sql.ErrNoRows) {
			return models.ErrCityWeatherNotFound
		}

		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return models.ErrCityWeatherNotFound
	}

	return nil
}
