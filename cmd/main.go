package main

import (
	"log/slog"
	"os"

	"github.com/dimakropachev/image_resizer_service/internal/app"
	"github.com/dimakropachev/image_resizer_service/internal/config"
	"github.com/dimakropachev/image_resizer_service/pkg/logger"
)

func main() {
	cfgPath := os.Getenv("CFG_PATH")
	if cfgPath == "" {
		cfgPath = "./config/config.yaml"
	}

	envPath := os.Getenv("ENV_PATH")
	if envPath == "" {
		envPath = "./config/.env.local"
	}

	cfg := config.MustLoad(envPath, cfgPath)

	l := logger.New(cfg.Env)
	slog.SetDefault(l)

	app, err := app.New(cfg)
	if err != nil {
		slog.Error("failed to create app", slog.String("error", err.Error()))
		os.Exit(1)
	}
	
	app.Start()
}
