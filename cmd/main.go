package main

import (
	"flag"
	"os"

	"github.com/JonnyShabli/23.07.2025/config"
	"github.com/JonnyShabli/23.07.2025/pkg/logster"
	"github.com/joho/godotenv"
)

const localConfig = "config/local/config_local.yaml"

func main() {
	var appConfig config.Config
	var configFile string

	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	// читаем флаги запуска
	flag.StringVar(&configFile, "config", localConfig, "Path to the config file")
	flag.Parse()
	err = config.LoadConfig(configFile, &appConfig)
	if err != nil {
		panic(err)
	}

	// Создаем логер
	logger := logster.New(os.Stdout, appConfig.Logger)
	defer func() { _ = logger.Sync() }()
}
