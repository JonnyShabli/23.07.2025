package http

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/JonnyShabli/23.07.2025/pkg/logster"
	"github.com/go-chi/chi/v5"
)

const (
	defaultReadHeaderTimeout = 5 * time.Second
	shutdownTimeout          = 5 * time.Second
)

type HTTPClient struct {
	Timeout time.Duration `yaml:"timeout"`
	Addr    string        `yaml:"addr"`
	Port    string        `yaml:"port"`
}

func NewHandler(basePath string, opts ...RouterOption) http.Handler {
	baseRouter := chi.NewRouter()
	baseRouter.Route(basePath, func(r chi.Router) {
		for _, opt := range opts {
			opt(r)
		}
	})
	return baseRouter
}

func NewServer(addr string, logger logster.Logger, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ErrorLog:          log.New(logger, "http_", 0),
		ReadHeaderTimeout: defaultReadHeaderTimeout,
	}
}

func RunServer(ctx context.Context, addr string, logger logster.Logger, handler http.Handler) error {
	logger.WithField("address", addr).Infof("Starting http server")
	server := NewServer(addr, logger, handler)
	errListen := make(chan error, 1)
	go func() {
		errListen <- server.ListenAndServe()
	}()
	select {
	case <-ctx.Done():
		ctxShutdown, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		err := server.Shutdown(ctxShutdown)
		if err != nil {
			return fmt.Errorf("can't shutdown server: %w", err)
		}
		return nil
	case err := <-errListen:
		return fmt.Errorf("can't run server: %w", err)
	}
}
