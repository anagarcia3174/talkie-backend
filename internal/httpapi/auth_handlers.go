package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/anagarcia3174/talkie-backend/internal/auth"
)

type AuthHandler struct {
	svc    *auth.Service
	logger *slog.Logger
}

type updateMeRequest struct {
	Username string `json:"username"`
}

func NewAuthHandler(svc *auth.Service, logger *slog.Logger) *AuthHandler {
	return &AuthHandler{svc: svc, logger: logger}
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		h.logger.Error("userID missing from context on protected route")
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	user, err := h.svc.GetCurrentUser(r.Context(), userID)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		h.logger.Error("get current user failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (h *AuthHandler) UpdateMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		h.logger.Error("userID missing from context on protected route")
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	var req updateMeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	user, err := h.svc.SetUsername(r.Context(), userID, req.Username)
	if err != nil {
		switch {
		case errors.Is(err, auth.ErrInvalidCredentials):
			writeError(w, http.StatusUnauthorized, "invalid credentials")
		case errors.Is(err, auth.ErrInvalidUsername):
			writeError(w, http.StatusBadRequest, "invalid username")
		case errors.Is(err, auth.ErrUsernameTaken):
			writeError(w, http.StatusConflict, "username already taken")
		default:
			h.logger.Error("update username failed", "error", err)
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusOK, user)
}
