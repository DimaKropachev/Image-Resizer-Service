package config

import (
	"fmt"
	"log/slog"

	"github.com/dimakropachev/image_resizer_service/internal/transport/http"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Size struct {
	Name   string `yaml:"name"`
	Width  int    `yaml:"width"`
	Height int    `yaml:"height"`
}

type Config struct {
	Env   string       `yaml:"env" env:"ENV" env-default:"prod"`
	HTTP  http.Config `yaml:"http"`
	Sizes []Size       `yaml:"sizes"`
}

func MustLoad(envPath, cfgPath string) *Config {
	err := godotenv.Load(envPath)
	if err != nil {
		slog.Warn("couldn't load env",
			slog.String("path", envPath),
			slog.String("error", err.Error()),
		)
	}

	cfg := &Config{}
	if err := cleanenv.ReadConfig(cfgPath, cfg); err != nil {
		panic(fmt.Errorf("couldn't read config file: %w", err))
	}

	return cfg
}
