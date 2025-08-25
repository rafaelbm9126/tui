package configpkg

import (
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Text map[string]map[string]map[string]string `yaml:"config"`
}

func GetEnv() {
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}
}

func LoadConfig() *Config {
	GetEnv()

	data, err := os.ReadFile("src/config.yaml")
	if err != nil {
		panic(err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		panic(err)
	}

	return &cfg
}
