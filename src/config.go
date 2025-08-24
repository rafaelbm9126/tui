package main

import (
	"encoding/json"
	"os"
)

type Config struct {
	Text struct {
		En struct {
			Comand struct {
				Help string `json:"help"`
			} `json:"comand"`
		} `json:"en"`
	} `json:"text"`
}

func LoadConfig() *Config {
	data, err := os.ReadFile("src/config.json")
	if err != nil {
		panic(err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		panic(err)
	}

	return &cfg
}
