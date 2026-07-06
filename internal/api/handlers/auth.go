package handlers

import (
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog"

	"github.com/atlasdb/atlasdb/internal/auth"
	"github.com/atlasdb/atlasdb/internal/storage/postgres"
	"github.com/atlasdb/atlasdb/pkg/models"
)

type AuthHandler struct {
	userStore  *postgres.UserStore
	jwtManager *auth.JWTManager
	logger     zerolog.Logger
}

func NewAuthHandler(
	userStore *postgres.UserStore,
	jwtManager *auth.JWTManager,
	logger zerolog.Logger,
) *AuthHandler {
	return &AuthHandler{
		userStore:  userStore,
		jwtManager: jwtManager,
		logger:     logger,
	}
}

type registerRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// POST /api/v1/auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, models.NewAPIError("bad_request", "Invalid request body"))
		return
	}

	if req.Email == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, models.NewAPIError("bad_request", "Email and password required"))
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusBadRequest, models.NewAPIError("bad_request", err.Error()))
		return
	}

	user, err := h.userStore.Create(r.Context(), req.Email, hash, "viewer")
	if err != nil {
		h.logger.Error().Err(err).Str("email", req.Email).Msg("Failed to create user")
		writeError(w, http.StatusConflict, models.NewAPIError("conflict", "Email already registered"))
		return
	}

	tokens, err := h.jwtManager.GenerateTokenPair(user.UserID, user.Email, user.Role)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to generate tokens")
		writeError(w, http.StatusInternalServerError, models.ErrInternalServer)
		return
	}

	writeJSON(w, http.StatusCreated, tokens)
}

// POST /api/v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, models.NewAPIError("bad_request", "Invalid request body"))
		return
	}

	user, err := h.userStore.GetByEmail(r.Context(), req.Email)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, http.StatusUnauthorized, models.NewAPIError("unauthorized", "Invalid credentials"))
			return
		}
		h.logger.Error().Err(err).Msg("Failed to look up user")
		writeError(w, http.StatusInternalServerError, models.ErrInternalServer)
		return
	}

	if !auth.CheckPassword(req.Password, user.PasswordHash) {
		writeError(w, http.StatusUnauthorized, models.NewAPIError("unauthorized", "Invalid credentials"))
		return
	}

	tokens, err := h.jwtManager.GenerateTokenPair(user.UserID, user.Email, user.Role)
	if err != nil {
		h.logger.Error().Err(err).Msg("Failed to generate tokens")
		writeError(w, http.StatusInternalServerError, models.ErrInternalServer)
		return
	}

	writeJSON(w, http.StatusOK, tokens)
}

// POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req refreshRequest
	if err := readJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, models.NewAPIError("bad_request", "Invalid request body"))
		return
	}

	userID, err := h.jwtManager.ValidateRefreshToken(req.RefreshToken)
	if err != nil {
		writeError(w, http.StatusUnauthorized, models.NewAPIError("unauthorized", "Invalid refresh token"))
		return
	}

	user, err := h.userStore.GetByID(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusUnauthorized, models.NewAPIError("unauthorized", "User not found"))
		return
	}

	tokens, err := h.jwtManager.GenerateTokenPair(user.UserID, user.Email, user.Role)
	if err != nil {
		writeError(w, http.StatusInternalServerError, models.ErrInternalServer)
		return
	}

	writeJSON(w, http.StatusOK, tokens)
}
