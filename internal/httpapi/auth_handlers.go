package httpapi

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/anagarcia3174/talkie-backend/internal/auth"
)

type AuthHandler struct {
	svc    *auth.Service
	logger *slog.Logger
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

