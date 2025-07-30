package main

import (
	"context"
	"errors"
	"flag"
	"os"

	"github.com/JonnyShabli/23.07.2025/config"
	"github.com/JonnyShabli/23.07.2025/internal/Service"
	d "github.com/JonnyShabli/23.07.2025/internal/Service/downloader"
	"github.com/JonnyShabli/23.07.2025/internal/Service/zipper"
	"github.com/JonnyShabli/23.07.2025/internal/controller"
	"github.com/JonnyShabli/23.07.2025/internal/repository"
	pkghttp "github.com/JonnyShabli/23.07.2025/pkg/http"
	"github.com/JonnyShabli/23.07.2025/pkg/logster"
	"github.com/JonnyShabli/23.07.2025/pkg/sig"
	"golang.org/x/sync/errgroup"
)

const localConfig = "config/config_local.yaml"

func main() {
	var appConfig config.Config
	var configFile string
	// читаем флаги запуска
	flag.StringVar(&configFile, "config", localConfig, "Path to the config file")
	flag.Parse()
	err := config.LoadConfig(configFile, &appConfig)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Создаем логер
	logger := logster.New(os.Stdout, appConfig.Logger)
	defer func() { _ = logger.Sync() }()

	logger.Infof("startins application with config %+v", appConfig)

	// создаем errgroup
	g, ctx := errgroup.WithContext(ctx)

	// Gracefully shutdown
	g.Go(func() error {
		return sig.ListenSignal(ctx, logger, cancel)
	})

	// создаем pool для скачивания файлов
	downloaderPool := d.NewDownloader(appConfig.Pool, logger)

	// создаем зависимости
	repo := repository.NewStorage(logger)
	// создаем zipper
	zipperMgr := zipper.NewZipper(appConfig.Zipper, logger, downloaderPool.Out, repo)

	service := Service.NewServiceObj(repo, logger, downloaderPool)
	handlerObj := controller.NewHandlers(service, logger, appConfig.HttpServer.Addr+":"+appConfig.HttpServer.Port)

	go downloaderPool.StartDownloader(ctx)

	_, err = os.Stat(appConfig.Zipper.ArchivePath)
	if os.IsNotExist(err) {
		err = os.MkdirAll(appConfig.Zipper.ArchivePath, 0775)
		if err != nil {
			panic(err)
		}
	}

	g.Go(func() error {
		return logster.LogIfError(
			logger, zipperMgr.Start(ctx, logger), "Zipper manager",
		)
	})

	// создаем хэндлер
	handler := pkghttp.NewHandler("/", pkghttp.WithLogger(logger), pkghttp.DefaultTechOptions(), controller.WithApiHandler(handlerObj))
	logger.Infof("Create and configure handler")

	// запускаем http server
	g.Go(func() error {
		return logster.LogIfError(
			logger, pkghttp.RunServer(ctx, appConfig.HttpServer.Addr+":"+appConfig.HttpServer.Port, logger, handler),
			"Api server",
		)
	})

	// ждем завершения
	err = g.Wait()
	if err != nil && !errors.Is(err, sig.ErrSignalReceived) {
		logger.WithError(err).Errorf("Exit reason")
	}
}
