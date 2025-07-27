package controller

import (
	pkghttp "github.com/JonnyShabli/23.07.2025/pkg/http"
	"github.com/go-chi/chi/v5"
)

func WithApiHandler(api HandlerInterface) pkghttp.RouterOption {
	return func(r chi.Router) {
		r.Route("/api/zipper", func(r chi.Router) {
			r.Get("/", api.AddTask)
			r.Post("/", api.AddLinks)
			r.Get("/status/{task_id}", api.GetStatus)

		})
	}
}
