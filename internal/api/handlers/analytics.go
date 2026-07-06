package handlers

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"

	"github.com/atlasdb/atlasdb/internal/analytics"
	"github.com/atlasdb/atlasdb/pkg/models"
)

type AnalyticsHandler struct {
	engine *analytics.Engine
	logger zerolog.Logger
}

func NewAnalyticsHandler(engine *analytics.Engine, logger zerolog.Logger) *AnalyticsHandler {
	return &AnalyticsHandler{engine: engine, logger: logger}
}

// GET /api/v1/analytics/summary
func (h *AnalyticsHandler) Summary(w http.ResponseWriter, r *http.Request) {
	start, end := parseTimeRange(r)

	summary, err := h.engine.Summary(r.Context(), start, end)
	if err != nil {
		h.logger.Error().Err(err).Msg("Analytics summary failed")
		writeError(w, http.StatusInternalServerError, models.ErrInternalServer)
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

// GET /api/v1/analytics/timeseries
func (h *AnalyticsHandler) TimeSeries(w http.ResponseWriter, r *http.Request) {
	start, end := parseTimeRange(r)

	req := analytics.TimeSeriesRequest{
		StartTime:  start,
		EndTime:    end,
		Resolution: r.URL.Query().Get("resolution"),
		Source:     r.URL.Query().Get("source"),
		Severity:   r.URL.Query().Get("severity"),
		GroupBy:    r.URL.Query().Get("group_by"),
	}

	if req.Resolution == "" {
		// Auto-select resolution based on range
		duration := end.Sub(start)
		if duration > 24*time.Hour {
			req.Resolution = "1h"
		} else {
			req.Resolution = "1m"
		}
	}

	data, err := h.engine.TimeSeries(r.Context(), req)
	if err != nil {
		h.logger.Error().Err(err).Msg("Analytics timeseries failed")
		writeError(w, http.StatusInternalServerError, models.ErrInternalServer)
		return
	}
	if data == nil {
		data = []analytics.TimeSeriesPoint{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":       data,
		"resolution": req.Resolution,
		"start_time": start,
		"end_time":   end,
	})
}

// GET /api/v1/analytics/top
func (h *AnalyticsHandler) TopN(w http.ResponseWriter, r *http.Request) {
	start, end := parseTimeRange(r)

	req := analytics.TopNRequest{
		StartTime: start,
		EndTime:   end,
		GroupBy:   r.URL.Query().Get("group_by"),
		Limit:     queryInt(r, "limit", 10),
	}
	if req.GroupBy == "" {
		req.GroupBy = "source"
	}

	data, err := h.engine.TopN(r.Context(), req)
	if err != nil {
		h.logger.Error().Err(err).Msg("Analytics top-N failed")
		writeError(w, http.StatusInternalServerError, models.ErrInternalServer)
		return
	}
	if data == nil {
		data = []analytics.TopNResult{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"data":     data,
		"group_by": req.GroupBy,
	})
}

func parseTimeRange(r *http.Request) (start, end time.Time) {
	end = time.Now().UTC()
	start = end.Add(-1 * time.Hour)

	if v := r.URL.Query().Get("start_time"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			start = t
		}
	}
	if v := r.URL.Query().Get("end_time"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			end = t
		}
	}

	// Shorthand ranges
	if v := r.URL.Query().Get("range"); v != "" {
		switch v {
		case "15m":
			start = end.Add(-15 * time.Minute)
		case "1h":
			start = end.Add(-1 * time.Hour)
		case "6h":
			start = end.Add(-6 * time.Hour)
		case "24h":
			start = end.Add(-24 * time.Hour)
		case "7d":
			start = end.Add(-7 * 24 * time.Hour)
		case "30d":
			start = end.Add(-30 * 24 * time.Hour)
		}
	}

	return start, end
}
