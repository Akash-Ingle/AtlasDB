package handlers

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"

	"github.com/atlasdb/atlasdb/internal/search"
	"github.com/atlasdb/atlasdb/pkg/models"
)

type SearchHandler struct {
	engine *search.Engine
	logger zerolog.Logger
}

func NewSearchHandler(engine *search.Engine, logger zerolog.Logger) *SearchHandler {
	return &SearchHandler{engine: engine, logger: logger}
}

// GET /api/v1/search
func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	if query == "" {
		writeError(w, http.StatusBadRequest, models.NewAPIError("bad_request", "Query parameter 'q' is required"))
		return
	}

	req := search.SearchRequest{
		Query:  query,
		Source: r.URL.Query().Get("source"),
		Severity: r.URL.Query().Get("severity"),
		Limit:  queryInt(r, "limit", models.DefaultPageSize),
		Cursor: r.URL.Query().Get("cursor"),
	}

	if v := r.URL.Query().Get("start_time"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			req.StartTime = &t
		}
	}

	if v := r.URL.Query().Get("end_time"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			req.EndTime = &t
		}
	}

	results, err := h.engine.Search(r.Context(), req)
	if err != nil {
		h.logger.Error().Err(err).Str("query", query).Msg("Search failed")
		writeError(w, http.StatusInternalServerError, models.ErrInternalServer)
		return
	}

	writeJSON(w, http.StatusOK, results)
}
