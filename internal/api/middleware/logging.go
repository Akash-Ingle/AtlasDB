package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
)

func Logger(logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)

			duration := time.Since(start)
			status := ww.Status()

			event := logger.Info()
			if status >= 500 {
				event = logger.Error()
			} else if status >= 400 {
				event = logger.Warn()
			}

			event.
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Int("status", status).
				Dur("duration", duration).
				Int("bytes", ww.BytesWritten()).
				Str("remote_addr", r.RemoteAddr).
				Str("request_id", GetRequestID(r.Context())).
				Msg("HTTP request")
		})
	}
}
