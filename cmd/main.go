package main

import (
	"context"
	"errors"
	"flag"
	"os"

	"github.com/JonnyShabli/23.07.2025/config"
	"github.com/JonnyShabli/23.07.2025/internal/Service"
	"github.com/JonnyShabli/23.07.2025/internal/controller"
	"github.com/JonnyShabli/23.07.2025/internal/repository"
	sig "github.com/JonnyShabli/23.07.2025/pkg"
	pkghttp "github.com/JonnyShabli/23.07.2025/pkg/http"
	"github.com/JonnyShabli/23.07.2025/pkg/logster"
	"github.com/joho/godotenv"
	"golang.org/x/sync/errgroup"
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

	// создаем errgroup
	g, ctx := errgroup.WithContext(context.Background())

	// Gracefully shutdown
	g.Go(func() error {
		return sig.ListenSignal(ctx, logger)
	})

	repo := repository.NewStorage(logger)
	service := Service.NewZipper(repo, logger)
	handlerObj := controller.NewHandlers(service, logger)

	// создаем хэндлер
	handler := pkghttp.NewHandler("/", pkghttp.WithLogger(logger), pkghttp.DefaultTechOptions(), controller.WithApiHandler(handlerObj))
	logger.Infof("Create and configure handler")

	// запускаем http server
	g.Go(func() error {
		return logster.LogIfError(
			logger, pkghttp.RunServer(ctx, appConfig.Addr+":"+appConfig.Port, logger, handler),
			"Api server",
		)
	})

	// ждем завершения
	err = g.Wait()
	if err != nil && !errors.Is(err, sig.ErrSignalReceived) {
		logger.WithError(err).Errorf("Exit reason")
	}
}
