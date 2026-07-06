package api

import (
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"

	"github.com/atlasdb/atlasdb/internal/api/handlers"
	"github.com/atlasdb/atlasdb/internal/api/middleware"
	"github.com/atlasdb/atlasdb/internal/auth"
	"github.com/atlasdb/atlasdb/internal/telemetry"
)

type RouterDeps struct {
	EventHandler      *handlers.EventHandler
	AuthHandler       *handlers.AuthHandler
	APIKeyHandler     *handlers.APIKeyHandler
	SearchHandler     *handlers.SearchHandler
	AnalyticsHandler  *handlers.AnalyticsHandler
	AlertHandler      *handlers.AlertHandler
	AIHandler         *handlers.AIHandler
	AdminHandler      *handlers.AdminHandler
	SystemHandler     *handlers.SystemHandler
	WebSocketHandler  *handlers.WebSocketHandler
	JWTManager        *auth.JWTManager
	Metrics           *telemetry.Metrics
	Logger            zerolog.Logger
}

func NewRouter(deps RouterDeps) *chi.Mux {
	r := chi.NewRouter()

	// Global middleware stack
	r.Use(middleware.Tracing("atlas-api"))
	r.Use(middleware.RequestID)
	r.Use(middleware.Recovery(deps.Logger))
	r.Use(middleware.Logger(deps.Logger))
	r.Use(middleware.Metrics(deps.Metrics))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:8080"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health endpoints (no auth)
	r.Get("/healthz", deps.SystemHandler.Healthz)
	r.Get("/readyz", deps.SystemHandler.Readyz)
	r.Handle("/metrics", promhttp.Handler())

	// Public auth routes
	r.Route("/api/v1/auth", func(r chi.Router) {
		r.Use(httprate.LimitByIP(10, 1*time.Minute))
		r.Post("/register", deps.AuthHandler.Register)
		r.Post("/login", deps.AuthHandler.Login)
		r.Post("/refresh", deps.AuthHandler.Refresh)
	})

	// Protected API routes
	r.Route("/api/v1", func(r chi.Router) {
		r.Use(middleware.Auth(deps.JWTManager))
		r.Use(middleware.Audit(deps.Logger))

		// Events
		r.Route("/events", func(r chi.Router) {
			r.Post("/", deps.EventHandler.Ingest)
			r.Get("/", deps.EventHandler.List)
			r.Get("/{id}", deps.EventHandler.Get)
		})

		// WebSocket (live stream)
		if deps.WebSocketHandler != nil {
			r.Get("/events/stream", deps.WebSocketHandler.Stream)
		}

		// Search
		r.Get("/search", deps.SearchHandler.Search)

		// Analytics
		if deps.AnalyticsHandler != nil {
			r.Route("/analytics", func(r chi.Router) {
				r.Get("/summary", deps.AnalyticsHandler.Summary)
				r.Get("/timeseries", deps.AnalyticsHandler.TimeSeries)
				r.Get("/top", deps.AnalyticsHandler.TopN)
			})
		}

		// Alerts
		if deps.AlertHandler != nil {
			r.Route("/alerts", func(r chi.Router) {
				r.Route("/rules", func(r chi.Router) {
					r.Post("/", deps.AlertHandler.CreateRule)
					r.Get("/", deps.AlertHandler.ListRules)
					r.Put("/{id}", deps.AlertHandler.UpdateRule)
					r.Delete("/{id}", deps.AlertHandler.DeleteRule)
				})
				r.Get("/history", deps.AlertHandler.History)
			})
		}

		// AI
		if deps.AIHandler != nil {
			r.Route("/ai", func(r chi.Router) {
				r.Post("/query", deps.AIHandler.Query)
				r.Post("/investigate", deps.AIHandler.Investigate)
				r.Get("/anomalies", deps.AIHandler.Anomalies)
			})
		}

		// Admin
		if deps.AdminHandler != nil {
			r.Route("/admin", func(r chi.Router) {
				r.Get("/dlq/stats", deps.AdminHandler.DLQStats)
				r.Get("/dlq/messages", deps.AdminHandler.DLQMessages)
				r.Post("/dlq/retry", deps.AdminHandler.DLQRetry)
				r.Delete("/dlq/purge", deps.AdminHandler.DLQPurge)
			})
		}

		// API Keys
		if deps.APIKeyHandler != nil {
			r.Route("/auth/api-keys", func(r chi.Router) {
				r.Post("/", deps.APIKeyHandler.Create)
				r.Get("/", deps.APIKeyHandler.List)
				r.Delete("/{id}", deps.APIKeyHandler.Delete)
			})
		}
	})

	return r
}
