package main

import (
	"context"
	"fmt"
	"time"
	"weather-cli/internal/config"
	"weather-cli/internal/provider/openmeteo"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Default city: %s\n", cfg.DefaultCity)

	cfg.DefaultCity = "Sochi"
	err = config.Save(cfg)
	if err != nil {
		panic(err)
	}

	client := openmeteo.NewClient()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	weather, err := client.GetToday(ctx, cfg.DefaultCity)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("City: %s\n", weather.City)
	fmt.Printf("Coordinates: %.4f, %.4f\n", 0.0, 0.0) // координаты не возвращаются напрямую
	fmt.Printf("Temperature: %.1f°C\n", weather.TemperatureC)
	fmt.Printf("Feels like: %.1f°C\n", weather.FeelsLikeC)
	fmt.Printf("Condition: %s\n", weather.Condition)
	fmt.Printf("Wind: %.1f m/s, %d°\n", weather.WindSpeedMS, weather.WindDirectionDeg)
	fmt.Printf("Humidity: %d%%\n", weather.HumidityPercent)
	fmt.Printf("Pressure: %d hPa\n", weather.PressureHPa)
	fmt.Printf("Visibility: %.1f km\n", weather.VisibilityKm)
	fmt.Printf("Precipitation: %.1f mm\n", weather.PrecipitationMm)
	fmt.Printf("Updated at: %s\n", weather.UpdatedAt.Format("15:04 02.01.2006"))
}
