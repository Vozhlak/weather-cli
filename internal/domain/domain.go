package domain

import "time"

type Today struct {
	City             string
	TemperatureC     float64
	FeelsLikeC       float64
	Condition        string
	WindSpeedMS      float64
	WindDirectionDeg int
	HumidityPercent  int
	PressureHPa      int
	VisibilityKm     float64
	PrecipitationMm  float64
	UpdatedAt        time.Time
}

type HourlyEntry struct {
	Time         time.Time
	TemperatureC float64
	POPPercent   int
	WindSpeedMS  float64
}

type DailyEntry struct {
	Date        time.Time
	TempMinC    float64
	TempMaxC    float64
	POPPercent  int
	WindSpeedMS float64
	Condition   string
}
