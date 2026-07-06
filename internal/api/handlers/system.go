package handlers

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type SystemHandler struct {
	pgPool *pgxpool.Pool
	redis  *redis.Client
}

func NewSystemHandler(pgPool *pgxpool.Pool, redis *redis.Client) *SystemHandler {
	return &SystemHandler{pgPool: pgPool, redis: redis}
}

// GET /healthz — liveness
func (h *SystemHandler) Healthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// GET /readyz — readiness
func (h *SystemHandler) Readyz(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	checks := map[string]string{}
	healthy := true

	if err := h.pgPool.Ping(ctx); err != nil {
		checks["postgres"] = "unhealthy: " + err.Error()
		healthy = false
	} else {
		checks["postgres"] = "healthy"
	}

	if err := h.redis.Ping(ctx).Err(); err != nil {
		checks["redis"] = "unhealthy: " + err.Error()
		healthy = false
	} else {
		checks["redis"] = "healthy"
	}

	status := http.StatusOK
	if !healthy {
		status = http.StatusServiceUnavailable
	}

	writeJSON(w, status, map[string]interface{}{
		"status": map[bool]string{true: "ready", false: "not_ready"}[healthy],
		"checks": checks,
	})
}
