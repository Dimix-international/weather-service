package db

import (
	"context"

	"github.com/Dimix-international/weather-service/internal/config"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type DB struct {
	DB *sqlx.DB
}

func NewConnect(ctx context.Context, cfg *config.Config) (*DB, error) {
	db, err := sqlx.Connect("postgres", cfg.Database.URI+"?sslmode=disable")
	if err != nil {
		return nil, err
	}

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return &DB{DB: db}, nil
}

func (d *DB) Close() error {
	return d.DB.Close()
}
