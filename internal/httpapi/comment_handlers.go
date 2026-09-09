package httpapi

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/anagarcia3174/talkie-backend/internal/comments"
)

type CommentHandler struct {
	svc    *comments.Service
	logger *slog.Logger
}

func NewCommentHandler(svc *comments.Service, logger *slog.Logger) *CommentHandler {
	return &CommentHandler{svc: svc, logger: logger}
}

type createCommentRequest struct {
	TMDBMediaType    string `json:"tmdb_media_type"`
	TMDBID           int32  `json:"tmdb_id"`
	SeasonNumber     *int32 `json:"season_number"`
	EpisodeNumber    *int32 `json:"episode_number"`
	TimestampSeconds int32  `json:"timestamp_seconds"`
	Body             string `json:"body"`
}

type updateCommentRequest struct {
	TimestampSeconds int32  `json:"timestamp_seconds"`
	Body             string `json:"body"`
}

func (h *CommentHandler) Create(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		h.logger.Error("userID missing from context on protected route")
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	var req createCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input := comments.CreateCommentInput{
		UserID:           userID,
		TMDBMediaType:    comments.MediaType(req.TMDBMediaType),
		TMDBID:           req.TMDBID,
		SeasonNumber:     req.SeasonNumber,
		EpisodeNumber:    req.EpisodeNumber,
		TimestampSeconds: req.TimestampSeconds,
		Body:             req.Body,
	}

	comment, err := h.svc.Create(r.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, comments.ErrInvalidInput):
			writeError(w, http.StatusBadRequest, "invalid input")
		default:
			h.logger.Error("create comment failed", "error", err)
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusCreated, comment)
}

func (h *CommentHandler) ListByMovie(w http.ResponseWriter, r *http.Request) {
	tmdbID, err := strconv.Atoi(chi.URLParam(r, "tmdb_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tmdb_id")
		return
	}

	list, err := h.svc.ListByMovie(r.Context(), int32(tmdbID))
	if err != nil {
		h.logger.Error("list movie comments failed", "error", err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, list)
}

func (h *CommentHandler) ListByEpisode(w http.ResponseWriter, r *http.Request) {
	tmdbID, err := strconv.Atoi(chi.URLParam(r, "tmdb_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tmdb_id")
		return
	}
	season, err := strconv.Atoi(chi.URLParam(r, "season"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid season number")
		return
	}
	episode, err := strconv.Atoi(chi.URLParam(r, "episode"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid episode number")
		return
	}

	list, err := h.svc.ListByEpisode(r.Context(), int32(tmdbID), int32(season), int32(episode))
	if err != nil {
		h.logger.Error("list episode comments failed", "error", err,
			"tmdb_id", tmdbID, "season", season, "episode", episode)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, list)
}

func (h *CommentHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		h.logger.Error("userID missing from context on protected route")
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	commentID := chi.URLParam(r, "comment_id")

	var req updateCommentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	input := comments.UpdateCommentInput{
		TimestampSeconds: req.TimestampSeconds,
		Body:             req.Body,
	}

	comment, err := h.svc.Update(r.Context(), commentID, userID, input)
	if err != nil {
		switch {
		case errors.Is(err, comments.ErrInvalidInput):
			writeError(w, http.StatusBadRequest, "invalid input")
		case errors.Is(err, comments.ErrNotFound):
			writeError(w, http.StatusNotFound, "comment not found")
		case errors.Is(err, comments.ErrForbidden):
			writeError(w, http.StatusForbidden, "not your comment")
		default:
			h.logger.Error("update comment failed", "error", err)
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	writeJSON(w, http.StatusOK, comment)
}

func (h *CommentHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, ok := UserIDFromContext(r.Context())
	if !ok {
		h.logger.Error("userID missing from context on protected route")
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	commentID := chi.URLParam(r, "comment_id")

	if err := h.svc.Delete(r.Context(), commentID, userID); err != nil {
		switch {
		case errors.Is(err, comments.ErrNotFound):
			writeError(w, http.StatusNotFound, "comment not found")
		case errors.Is(err, comments.ErrForbidden):
			writeError(w, http.StatusForbidden, "not your comment")
		default:
			h.logger.Error("delete comment failed", "error", err)
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}