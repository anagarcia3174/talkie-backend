package httpapi

import (
	"encoding/json"
	"errors"
	"io"
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

type deleteMeRequest struct {
	AppleAuthorizationCode string `json:"apple_authorization_code"`
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
		case errors.Is(err, auth.ErrUsernameNotAllowed):
			writeError(w, http.StatusUnprocessableEntity, "objectionable content")
		default:
			h.logger.Error("update username failed", "error", err)
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (h *AuthHandler) AcceptTerms(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		h.logger.Error("userID missing from context on protected route")
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	user, err := h.svc.AcceptTerms(r.Context(), userID)
	if err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		h.logger.Error("accept terms failed", "error", err, "user_id", userID)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (h *AuthHandler) DeleteMe(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		h.logger.Error("userID missing from context on protected route")
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	var req deleteMeRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<13)).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.svc.DeleteAccount(r.Context(), userID, req.AppleAuthorizationCode); err != nil {
		if errors.Is(err, auth.ErrInvalidCredentials) {
			writeError(w, http.StatusUnauthorized, "invalid credentials")
			return
		}
		h.logger.Error("delete account failed", "error", err, "user_id", userID)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.logger.Info("account deleted", "user_id", userID)
	w.WriteHeader(http.StatusNoContent)
}
