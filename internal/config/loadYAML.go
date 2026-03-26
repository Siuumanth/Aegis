package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

func Load() (*Config, error) {
	data, err := os.ReadFile("./aegis.yaml")
	if err != nil {
		return nil, err
	}
	fmt.Println("Config file:", string(data))

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
