package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/anagarcia3174/talkie-backend/internal/reports"
)

type ReportHandler struct {
	svc    *reports.Service
	logger *slog.Logger
}

func NewReportHandler(svc *reports.Service, logger *slog.Logger) *ReportHandler {
	return &ReportHandler{svc: svc, logger: logger}
}

type createReportRequest struct {
	Reason string `json:"reason"`
	Detail string `json:"detail"`
}

func (h *ReportHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		h.logger.Error("userID missing from context on protected route")
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	var req createReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	report, err := h.svc.Create(r.Context(), reports.CreateReportInput{
		CommentID: chi.URLParam(r, "comment_id"),
		UserID:    userID,
		Reason:    reports.Reason(req.Reason),
		Detail:    req.Detail,
	})
	if err != nil {
		switch {
		case errors.Is(err, reports.ErrInvalidInput):
			writeError(w, http.StatusBadRequest, "invalid input")
		case errors.Is(err, reports.ErrNotFound):
			writeError(w, http.StatusNotFound, "comment not found")
		case errors.Is(err, reports.ErrOwnComment):
			writeError(w, http.StatusForbidden, "you can't report your own comment")
		case errors.Is(err, reports.ErrAlreadyReported):
			writeError(w, http.StatusConflict, "you already reported this comment")
		default:
			h.logger.Error("create report failed", "error", err)
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusCreated, report)
}
