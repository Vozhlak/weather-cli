package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	DefaultCity string `yaml:"default_city"`
}

const configPath = "config.yaml"

func Load() (Config, error) {
	var cfg Config

	data, err := os.ReadFile(configPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return cfg, nil
		}

		return cfg, err
	}

	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return cfg, err
	}

	return cfg, nil
}

func Save(cfg Config) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	err = os.WriteFile(configPath, data, 0644)
	if err != nil {
		return err
	}

	return nil
}
