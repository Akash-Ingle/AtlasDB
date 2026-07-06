package handlers

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"

	"github.com/atlasdb/atlasdb/internal/api/middleware"
	"github.com/atlasdb/atlasdb/internal/auth"
	"github.com/atlasdb/atlasdb/pkg/models"
)

type APIKeyHandler struct {
	store  *auth.APIKeyStore
	logger zerolog.Logger
}

func NewAPIKeyHandler(store *auth.APIKeyStore, logger zerolog.Logger) *APIKeyHandler {
	return &APIKeyHandler{store: store, logger: logger}
}

type createAPIKeyRequest struct {
	Name      string   `json:"name"`
	Scopes    []string `json:"scopes"`
	ExpiresIn *int     `json:"expires_in_days,omitempty"`
}

// POST /api/v1/auth/api-keys
func (h *APIKeyHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, models.ErrUnauthorized)
		return
	}

	var req createAPIKeyRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, models.NewAPIError("bad_request", "Invalid request body"))
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, models.NewAPIError("bad_request", "Name is required"))
		return
	}

	if len(req.Scopes) == 0 {
		req.Scopes = []string{"events:read", "events:write", "search:read"}
	}

	var expiresAt *time.Time
	if req.ExpiresIn != nil {
		t := time.Now().AddDate(0, 0, *req.ExpiresIn)
		expiresAt = &t
	}

	key, err := h.store.Create(r.Context(), claims.UserID, req.Name, req.Scopes, expiresAt)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to create API key")
		writeError(w, http.StatusInternalServerError, models.ErrInternalServer)
		return
	}

	writeJSON(w, http.StatusCreated, key)
}

// GET /api/v1/auth/api-keys
func (h *APIKeyHandler) List(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, models.ErrUnauthorized)
		return
	}

	keys, err := h.store.ListByUser(r.Context(), claims.UserID)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to list API keys")
		writeError(w, http.StatusInternalServerError, models.ErrInternalServer)
		return
	}

	if keys == nil {
		keys = []auth.APIKey{}
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"api_keys": keys})
}

// DELETE /api/v1/auth/api-keys/{id}
func (h *APIKeyHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		writeError(w, http.StatusUnauthorized, models.ErrUnauthorized)
		return
	}

	keyID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, models.NewAPIError("bad_request", "Invalid key ID"))
		return
	}

	err = h.store.Delete(r.Context(), keyID, claims.UserID)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, http.StatusNotFound, models.ErrNotFound)
			return
		}
		h.logger.Error().Err(err).Msg("Failed to delete API key")
		writeError(w, http.StatusInternalServerError, models.ErrInternalServer)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
