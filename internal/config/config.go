package config

import (
	"log/slog"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type HTTP struct {
	Port    int           `yaml:"port" env:"HTTP_PORT" env-default:"8088"`
	Timeout time.Duration `yaml:"timeout" env:"HTTP_TIMEOUT" env-default:"5s"`
}

type Size struct {
	Name   string `yaml:"name"`
	Width  int    `yaml:"width"`
	Height int    `yaml:"height"`
}

type Storage struct {
	UploadPath    string `yaml:"upload_dir" env:"UPLOAD_DIR"`
	ProcessedPath string `yaml:"processed_dir" env:"PROCESSED_DIR"`
}

type Config struct {
	Env     string  `yaml:"env" env:"ENV" env-default:"prod"`
	HTTP    HTTP    `yaml:"http"`
	Sizes   []Size  `yaml:"sizes"`
	Workers int     `yaml:"worker_num"`
	Storage Storage `yaml:"storage"`
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
		slog.Error("failed to read config file", slog.String("error", err.Error()))
		os.Exit(1)
	}

	return cfg
}
