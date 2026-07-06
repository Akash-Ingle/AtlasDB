package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/atlasdb/atlasdb/internal/ai"
	"github.com/atlasdb/atlasdb/pkg/models"
)

type AIHandler struct {
	queryEngine   *ai.QueryEngine
	investigator  *ai.Investigator
	anomalyDet    *ai.AnomalyDetector
	logger        zerolog.Logger
}

func NewAIHandler(
	queryEngine *ai.QueryEngine,
	investigator *ai.Investigator,
	anomalyDet *ai.AnomalyDetector,
	logger zerolog.Logger,
) *AIHandler {
	return &AIHandler{
		queryEngine:  queryEngine,
		investigator: investigator,
		anomalyDet:   anomalyDet,
		logger:       logger,
	}
}

// POST /api/v1/ai/query
func (h *AIHandler) Query(w http.ResponseWriter, r *http.Request) {
	var req ai.NLQueryRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, models.NewAPIError("bad_request", "Invalid request body"))
		return
	}

	if req.Question == "" {
		writeError(w, http.StatusBadRequest, models.NewAPIError("bad_request", "Question is required"))
		return
	}

	// Translate the question into a structured query
	query, err := h.queryEngine.TranslateQuery(r.Context(), req.Question)
	if err != nil {
		h.logger.Error().Err(err).Str("question", req.Question).Msg("Query translation failed")
		writeError(w, http.StatusInternalServerError, models.NewAPIError("ai_error", "Failed to translate query"))
		return
	}

	// For now, return the translated query and let the summarizer provide the answer
	// In a full implementation, we'd execute the query against the store and then summarize
	resp := &ai.NLQueryResponse{
		Answer:     fmt.Sprintf("Generated query of type '%s' for your question.", query.Type),
		Query:      query,
		Confidence: 0.8,
	}

	writeJSON(w, http.StatusOK, resp)
}

// POST /api/v1/ai/investigate
func (h *AIHandler) Investigate(w http.ResponseWriter, r *http.Request) {
	var req ai.InvestigationRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, models.NewAPIError("bad_request", "Invalid request body"))
		return
	}

	if req.Question == "" {
		writeError(w, http.StatusBadRequest, models.NewAPIError("bad_request", "Question is required"))
		return
	}

	if req.Stream {
		h.investigateStream(w, r, req)
		return
	}

	result, err := h.investigator.Investigate(r.Context(), req)
	if err != nil {
		h.logger.Error().Err(err).Str("question", req.Question).Msg("Investigation failed")
		writeError(w, http.StatusInternalServerError, models.NewAPIError("ai_error", "Investigation failed"))
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *AIHandler) investigateStream(w http.ResponseWriter, r *http.Request, req ai.InvestigationRequest) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, models.NewAPIError("sse_error", "Streaming not supported"))
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	steps, err := h.investigator.InvestigateStream(r.Context(), req)
	if err != nil {
		data, _ := json.Marshal(map[string]string{"error": err.Error()})
		fmt.Fprintf(w, "event: error\ndata: %s\n\n", data)
		flusher.Flush()
		return
	}

	for step := range steps {
		data, _ := json.Marshal(step)
		fmt.Fprintf(w, "event: step\ndata: %s\n\n", data)
		flusher.Flush()
	}

	fmt.Fprintf(w, "event: done\ndata: {}\n\n")
	flusher.Flush()
}

// GET /api/v1/ai/anomalies
func (h *AIHandler) Anomalies(w http.ResponseWriter, r *http.Request) {
	report, err := h.anomalyDet.DetectAnomalies(r.Context())
	if err != nil {
		h.logger.Error().Err(err).Msg("Anomaly detection failed")
		writeError(w, http.StatusInternalServerError, models.NewAPIError("ai_error", "Anomaly detection failed"))
		return
	}

	writeJSON(w, http.StatusOK, report)
}
