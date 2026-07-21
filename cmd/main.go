package main

import (
	"fmt"
	"os"

	"github.com/dimakropachev/image_resizer_service/internal/config"
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

	fmt.Println(*cfg)
	// ...
}
