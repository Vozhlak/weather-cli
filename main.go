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

	cfg.DefaultCity = "Sochi"
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

	fmt.Println(ui.Header(weather.City, false, time.Now().Add(-3*time.Minute)))
	fmt.Println(ui.RenderToday(weather))
	fmt.Println(ui.RenderMenu())
}
