package db

import (
	"context"

	"github.com/Dimix-international/weather-service/internal/models"
)

type WeatheStorage interface {
	GetByCityName(ctx context.Context, city string) (models.Weather, error)
	Create(ctx context.Context, data *models.Weather) error
	Update(ctx context.Context, data *models.Weather) error
}
