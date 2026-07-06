package handlers

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"

	"github.com/atlasdb/atlasdb/internal/ingest"
	"github.com/atlasdb/atlasdb/internal/storage/postgres"
	"github.com/atlasdb/atlasdb/pkg/models"
)

type EventHandler struct {
	ingester   *ingest.Ingester
	eventStore *postgres.EventStore
	logger     zerolog.Logger
}

func NewEventHandler(
	ingester *ingest.Ingester,
	eventStore *postgres.EventStore,
	logger zerolog.Logger,
) *EventHandler {
	return &EventHandler{
		ingester:   ingester,
		eventStore: eventStore,
		logger:     logger,
	}
}

// POST /api/v1/events
func (h *EventHandler) Ingest(w http.ResponseWriter, r *http.Request) {
	var req models.IngestRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, models.NewAPIError("bad_request", "Invalid request body: "+err.Error()))
		return
	}

	// Support single event sent as object (wrap in array)
	if len(req.Events) == 0 {
		var single models.EventInput
		if err := readJSON(r, &single); err == nil && single.Source != "" {
			req.Events = []models.EventInput{single}
		}
	}

	if len(req.Events) == 0 {
		writeError(w, http.StatusBadRequest, models.NewAPIError("bad_request", "No events provided"))
		return
	}

	resp, err := h.ingester.Ingest(r.Context(), req.Events)
	if err != nil {
		if isBackpressure(err) {
			w.Header().Set("Retry-After", "5")
			writeError(w, http.StatusTooManyRequests, models.NewAPIError("rate_limited", err.Error()))
			return
		}
		h.logger.Error().Err(err).Msg("Ingest failed")
		writeError(w, http.StatusInternalServerError, models.ErrInternalServer)
		return
	}

	writeJSON(w, http.StatusAccepted, resp)
}

// GET /api/v1/events
func (h *EventHandler) List(w http.ResponseWriter, r *http.Request) {
	q := models.EventQuery{
		Source:    r.URL.Query().Get("source"),
		EventType: r.URL.Query().Get("event_type"),
		Severity:  models.Severity(r.URL.Query().Get("severity")),
		Limit:    queryInt(r, "limit", models.DefaultPageSize),
		Cursor:   r.URL.Query().Get("cursor"),
	}

	if v := r.URL.Query().Get("start_time"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			q.StartTime = &t
		}
	}

	if v := r.URL.Query().Get("end_time"); v != "" {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			q.EndTime = &t
		}
	}

	resp, err := h.eventStore.Query(r.Context(), q)
	if err != nil {
		h.logger.Error().Err(err).Msg("Query events failed")
		writeError(w, http.StatusInternalServerError, models.ErrInternalServer)
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// GET /api/v1/events/{id}
func (h *EventHandler) Get(w http.ResponseWriter, r *http.Request) {
	eventID := chi.URLParam(r, "id")
	if eventID == "" {
		writeError(w, http.StatusBadRequest, models.NewAPIError("bad_request", "Event ID required"))
		return
	}

	event, err := h.eventStore.GetByID(r.Context(), eventID)
	if err != nil {
		writeError(w, http.StatusNotFound, models.ErrNotFound)
		return
	}

	writeJSON(w, http.StatusOK, event)
}

func isBackpressure(err error) bool {
	return err != nil && len(err.Error()) > 0 && err.Error()[:5] == "queue"
}
