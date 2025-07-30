package sig

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/JonnyShabli/23.07.2025/pkg/logster"
)

var ErrSignalReceived = errors.New("operating system signal")

func ListenSignal(ctx context.Context, logger logster.Logger, cancel context.CancelFunc) error {
	sigquit := make(chan os.Signal, 1)
	signal.Ignore(syscall.SIGHUP, syscall.SIGPIPE)
	signal.Notify(sigquit, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-ctx.Done():
		return nil
	case sig := <-sigquit:
		cancel()
		logger.WithField("signal", sig).Infof("Captured signal")
		logger.Infof("Gracefully shutting down server...")
		return ErrSignalReceived
	}
}
