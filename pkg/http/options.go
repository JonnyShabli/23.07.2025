package http

import (
	"github.com/JonnyShabli/23.07.2025/pkg/logster"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func DefaultTechOptions() RouterOption {
	return RouterOptions(
		WithRecover(),
		WithDebugHandler(),
	)
}

func RouterOptions(options ...RouterOption) func(chi.Router) {
	return func(r chi.Router) {
		for _, option := range options {
			option(r)
		}
	}
}

type RouterOption func(chi.Router)

func WithDebugHandler() RouterOption {
	return func(r chi.Router) {
		r.Mount("/debug", middleware.Profiler())
	}
}

func WithRecover() RouterOption {
	return func(r chi.Router) {
		r.Use(middleware.Recoverer)
	}
}

func WithLogger(loger logster.Logger) RouterOption {
	return func(r chi.Router) {
		r.Use(logster.LogsterMiddleware(loger))
	}
}
