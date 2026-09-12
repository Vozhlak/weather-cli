package main

import (
	"fmt"
	"weather-cli/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	fmt.Printf("Default city: %s\n", cfg.DefaultCity)

	cfg.DefaultCity = "Novosibirsk"
	err = config.Save(cfg)
	if err != nil {
		panic(err)
	}
}
