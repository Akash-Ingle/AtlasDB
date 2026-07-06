package handlers

import (
	"net/http"
	"strings"

	"github.com/rs/zerolog"

	"github.com/atlasdb/atlasdb/internal/queue"
	"github.com/atlasdb/atlasdb/pkg/models"
)

type AdminHandler struct {
	dlq        *queue.DLQManager
	partitions int
	logger     zerolog.Logger
}

func NewAdminHandler(dlq *queue.DLQManager, partitions int, logger zerolog.Logger) *AdminHandler {
	return &AdminHandler{dlq: dlq, partitions: partitions, logger: logger}
}

// GET /api/v1/admin/dlq/stats
func (h *AdminHandler) DLQStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.dlq.Stats(r.Context(), h.partitions)
	if err != nil {
		writeError(w, http.StatusInternalServerError, models.ErrInternalServer)
		return
	}
	if stats == nil {
		stats = []queue.DLQStats{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"streams": stats})
}

// GET /api/v1/admin/dlq/messages?stream=dlq:0&count=50
func (h *AdminHandler) DLQMessages(w http.ResponseWriter, r *http.Request) {
	stream := r.URL.Query().Get("stream")
	if stream == "" || !strings.HasPrefix(stream, "dlq:") {
		writeError(w, http.StatusBadRequest, models.NewAPIError("bad_request", "stream parameter required (e.g. dlq:0)"))
		return
	}

	count := int64(queryInt(r, "count", 50))
	msgs, err := h.dlq.List(r.Context(), stream, count)
	if err != nil {
		h.logger.Error().Err(err).Msg("DLQ list failed")
		writeError(w, http.StatusInternalServerError, models.ErrInternalServer)
		return
	}
	if msgs == nil {
		msgs = []queue.DLQMessage{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"stream":   stream,
		"messages": msgs,
		"count":    len(msgs),
	})
}

// POST /api/v1/admin/dlq/retry
func (h *AdminHandler) DLQRetry(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Stream     string   `json:"stream"`
		MessageIDs []string `json:"message_ids"`
		Partition  int      `json:"partition"`
	}
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, models.NewAPIError("bad_request", "Invalid request"))
		return
	}

	retried, err := h.dlq.Retry(r.Context(), req.Stream, req.MessageIDs, req.Partition)
	if err != nil {
		writeError(w, http.StatusInternalServerError, models.ErrInternalServer)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"retried": retried})
}

// DELETE /api/v1/admin/dlq/purge?stream=dlq:0
func (h *AdminHandler) DLQPurge(w http.ResponseWriter, r *http.Request) {
	stream := r.URL.Query().Get("stream")
	if stream == "" {
		writeError(w, http.StatusBadRequest, models.NewAPIError("bad_request", "stream parameter required"))
		return
	}

	purged, err := h.dlq.Purge(r.Context(), stream)
	if err != nil {
		writeError(w, http.StatusInternalServerError, models.ErrInternalServer)
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"purged": purged})
}
