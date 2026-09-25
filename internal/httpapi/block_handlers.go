package httpapi

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/anagarcia3174/talkie-backend/internal/blocks"
)

type BlockHandler struct {
	svc    *blocks.Service
	logger *slog.Logger
}

func NewBlockHandler(svc *blocks.Service, logger *slog.Logger) *BlockHandler {
	return &BlockHandler{svc: svc, logger: logger}
}

func (h *BlockHandler) Block(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		h.logger.Error("userID missing from context on protected route")
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if err := h.svc.Block(r.Context(), userID, chi.URLParam(r, "user_id")); err != nil {
		h.writeBlockError(w, err, "block user failed")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *BlockHandler) Unblock(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		h.logger.Error("userID missing from context on protected route")
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if err := h.svc.Unblock(r.Context(), userID, chi.URLParam(r, "user_id")); err != nil {
		h.writeBlockError(w, err, "unblock user failed")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *BlockHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		h.logger.Error("userID missing from context on protected route")
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	list, err := h.svc.List(r.Context(), userID)
	if err != nil {
		h.writeBlockError(w, err, "list blocked users failed")
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *BlockHandler) writeBlockError(w http.ResponseWriter, err error, logMsg string) {
	switch {
	case errors.Is(err, blocks.ErrInvalidInput):
		writeError(w, http.StatusUnauthorized, "invalid credentials")
	case errors.Is(err, blocks.ErrCannotBlockSelf):
		writeError(w, http.StatusBadRequest, "cannot block yourself")
	case errors.Is(err, blocks.ErrUserNotFound):
		writeError(w, http.StatusNotFound, "user not found")
	default:
		h.logger.Error(logMsg, "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}
