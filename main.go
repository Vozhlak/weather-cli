package main

import (
	"context"
	"fmt"
	"time"
	"weather-cli/internal/config"
	"weather-cli/internal/provider/openmeteo"
	"weather-cli/internal/ui"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Default city: %s\n", cfg.DefaultCity)

	cfg.DefaultCity = "Сочи"
	err = config.Save(cfg)
	if err != nil {
		panic(err)
	}

	client := openmeteo.NewClient()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	weather, err := client.GetToday(ctx, cfg.DefaultCity)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	weatherHourly, err := client.GetHourly(ctx, cfg.DefaultCity, 12)
	if err != nil {
		fmt.Printf("Error get hourly: %v\n", err)
		return
	}

	weatherDaily, err := client.GetDaily(ctx, cfg.DefaultCity, 7)
	if err != nil {
		fmt.Printf("Error get daily: %v\n", err)
		return
	}

	fmt.Println(ui.Header(weather.City, false, time.Now().Add(-3*time.Minute)))
	//fmt.Println(ui.RenderToday(weather))

	fmt.Printf("\nПочасовой прогноз (%d часов):\n%s", len(weatherHourly), ui.RenderHourly(weatherHourly))
	fmt.Println()

	fmt.Printf("\n Прогноз на неделю:\n%s", ui.RenderDaily(weatherDaily))
	fmt.Println()

	fmt.Printf("\n%s", ui.RenderMenu())
}
