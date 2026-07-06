package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// Audit logs mutating API operations (POST, PUT, DELETE) with user context.
func Audit(logger zerolog.Logger) func(http.Handler) http.Handler {
	auditLog := logger.With().Str("component", "audit").Logger()

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Only audit mutating operations
			if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()

			// Capture response status
			wrapped := &statusWriter{ResponseWriter: w, status: 200}
			next.ServeHTTP(wrapped, r)

			// Extract user from context
			userID := "-"
			if claims := GetClaims(r.Context()); claims != nil {
				userID = claims.UserID.String()
			}

			auditLog.Info().
				Str("method", r.Method).
				Str("path", r.URL.Path).
				Str("user_id", userID).
				Str("remote_addr", extractIP(r)).
				Int("status", wrapped.status).
				Dur("duration", time.Since(start)).
				Str("user_agent", r.UserAgent()).
				Msg("audit")
		})
	}
}

func extractIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.SplitN(xff, ",", 2)
		return strings.TrimSpace(parts[0])
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	return r.RemoteAddr
}

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}
