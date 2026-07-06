package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/atlasdb/atlasdb/internal/alerts"
	"github.com/atlasdb/atlasdb/internal/api/middleware"
	"github.com/atlasdb/atlasdb/pkg/models"
)

type AlertHandler struct {
	store  *alerts.Store
	logger zerolog.Logger
}

func NewAlertHandler(store *alerts.Store, logger zerolog.Logger) *AlertHandler {
	return &AlertHandler{store: store, logger: logger}
}

type createAlertRuleRequest struct {
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Condition       alerts.AlertCondition  `json:"condition"`
	Severity        string                 `json:"severity"`
	Channels        json.RawMessage        `json:"channels"`
	Enabled         *bool                  `json:"enabled"`
	CooldownSeconds int                    `json:"cooldown_seconds"`
}

// POST /api/v1/alerts/rules
func (h *AlertHandler) CreateRule(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())

	var req createAlertRuleRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, models.NewAPIError("bad_request", "Invalid request body"))
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, models.NewAPIError("bad_request", "Name is required"))
		return
	}
	if req.Condition.Metric == "" {
		writeError(w, http.StatusBadRequest, models.NewAPIError("bad_request", "Condition metric is required"))
		return
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	if req.Severity == "" {
		req.Severity = "warning"
	}
	if req.CooldownSeconds == 0 {
		req.CooldownSeconds = 300
	}
	if req.Channels == nil {
		req.Channels = json.RawMessage(`[]`)
	}

	var createdBy *uuid.UUID
	if claims != nil {
		createdBy = &claims.UserID
	}

	rule := &alerts.AlertRule{
		RuleID:          uuid.New(),
		Name:            req.Name,
		Description:     req.Description,
		Condition:       req.Condition,
		Severity:        req.Severity,
		Channels:        req.Channels,
		Enabled:         enabled,
		CooldownSeconds: req.CooldownSeconds,
		CreatedBy:       createdBy,
	}

	if err := h.store.CreateRule(r.Context(), rule); err != nil {
		h.logger.Error().Err(err).Msg("Failed to create alert rule")
		writeError(w, http.StatusInternalServerError, models.ErrInternalServer)
		return
	}

	writeJSON(w, http.StatusCreated, rule)
}

// GET /api/v1/alerts/rules
func (h *AlertHandler) ListRules(w http.ResponseWriter, r *http.Request) {
	rules, err := h.store.ListRules(r.Context())
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to list alert rules")
		writeError(w, http.StatusInternalServerError, models.ErrInternalServer)
		return
	}
	if rules == nil {
		rules = []alerts.AlertRule{}
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{"rules": rules})
}

// PUT /api/v1/alerts/rules/{id}
func (h *AlertHandler) UpdateRule(w http.ResponseWriter, r *http.Request) {
	ruleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, models.NewAPIError("bad_request", "Invalid rule ID"))
		return
	}

	existing, err := h.store.GetRule(r.Context(), ruleID)
	if err != nil {
		writeError(w, http.StatusNotFound, models.ErrNotFound)
		return
	}

	var req createAlertRuleRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, models.NewAPIError("bad_request", "Invalid request body"))
		return
	}

	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Description != "" {
		existing.Description = req.Description
	}
	if req.Condition.Metric != "" {
		existing.Condition = req.Condition
	}
	if req.Severity != "" {
		existing.Severity = req.Severity
	}
	if req.Channels != nil {
		existing.Channels = req.Channels
	}
	if req.Enabled != nil {
		existing.Enabled = *req.Enabled
	}
	if req.CooldownSeconds > 0 {
		existing.CooldownSeconds = req.CooldownSeconds
	}

	if err := h.store.UpdateRule(r.Context(), existing); err != nil {
		h.logger.Error().Err(err).Msg("Failed to update alert rule")
		writeError(w, http.StatusInternalServerError, models.ErrInternalServer)
		return
	}

	writeJSON(w, http.StatusOK, existing)
}

// DELETE /api/v1/alerts/rules/{id}
func (h *AlertHandler) DeleteRule(w http.ResponseWriter, r *http.Request) {
	ruleID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, models.NewAPIError("bad_request", "Invalid rule ID"))
		return
	}

	if err := h.store.DeleteRule(r.Context(), ruleID); err != nil {
		writeError(w, http.StatusNotFound, models.ErrNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// GET /api/v1/alerts/history
func (h *AlertHandler) History(w http.ResponseWriter, r *http.Request) {
	limit := queryInt(r, "limit", 50)

	events, err := h.store.ListAlertEvents(r.Context(), limit)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to list alert events")
		writeError(w, http.StatusInternalServerError, models.ErrInternalServer)
		return
	}
	if events == nil {
		events = []alerts.AlertEvent{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"alerts": events})
}
