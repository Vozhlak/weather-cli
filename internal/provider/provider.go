package provider

import (
	"context"
	"weather-cli/internal/domain"
)

type WeatherProvider interface {
	GetToday(ctx context.Context, city string) (domain.Today, error)
	GetHourly(ctx context.Context, city string, hours int) ([]domain.HourlyEntry, error)
	GetDaily(ctx context.Context, city string, days int) ([]domain.DailyEntry, error)
}
