package main

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Text map[string]map[string]map[string]string `yaml:"config"`
}

func LoadConfig() (*Config, string) {
	data, err := os.ReadFile("src/config.yaml")
	if err != nil {
		panic(err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		panic(err)
	}

	return &cfg, string(data)
}
