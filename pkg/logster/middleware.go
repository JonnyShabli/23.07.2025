package logster

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

func LogsterMiddleware(logger Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)

			logger.WithField("method", r.Method).
				WithField("path", r.URL.Path).
				WithField("status", ww.Status()).
				WithField("duration", time.Since(start)).
				WithField("ip", r.RemoteAddr).Infof("HTTP request")

		})
	}
}
