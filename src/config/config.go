package configpkg

import (
	"os"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Config struct {
		Messages struct {
			Commands struct {
				Title      string `yaml:"title"`
				Collection []struct {
					Command     string `yaml:"command"`
					Description string `yaml:"description"`
					Variants    []struct {
						Command     string `yaml:"command"`
						Description string `yaml:"description"`
					} `yaml:"variants"`
				} `yaml:"collection"`
			} `yaml:"commands"`
		} `yaml:"messages"`
	} `yaml:"config"`
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
