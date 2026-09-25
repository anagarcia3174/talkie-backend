package httpapi

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/anagarcia3174/talkie-backend/internal/tmdb"
	"github.com/go-chi/chi/v5"
)

type MediaHandler struct {
	client *tmdb.Client
	logger *slog.Logger
}

func NewMediaHandler(client *tmdb.Client, logger *slog.Logger) *MediaHandler {
	return &MediaHandler{client: client, logger: logger}
}

func (h *MediaHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" {
		writeError(w, http.StatusBadRequest, "query parameter is required")
		return
	}

	results, err := h.client.SearchMulti(r.Context(), query)
	if err != nil {
		h.logger.Error("tmdb search failed", "error", err, "query", query)
		writeError(w, http.StatusBadGateway, "failed to search media")
		return
	}

	writeJSON(w, http.StatusOK, results)
}

func (h *MediaHandler) GetMovie(w http.ResponseWriter, r *http.Request) {
	tmdbID, err := strconv.Atoi(chi.URLParam(r, "tmdb_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tmdb_id")
		return
	}

	details, err := h.client.GetMovieDetails(r.Context(), tmdbID)
	if err != nil {
		h.logger.Error("tmdb movie details failed", "error", err, "tmdb_id", tmdbID)
		writeError(w, http.StatusBadGateway, "failed to fetch movie details")
		return
	}

	writeJSON(w, http.StatusOK, details)
}

func (h *MediaHandler) GetTVShow(w http.ResponseWriter, r *http.Request) {
	tmdbID, err := strconv.Atoi(chi.URLParam(r, "tmdb_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid tmdb_id")
		return
	}

	details, err := h.client.GetTVShowDetails(r.Context(), tmdbID)
	if err != nil {
		h.logger.Error("tmdb tv show details failed", "error", err, "tmdb_id", tmdbID)
		writeError(w, http.StatusBadGateway, "failed to fetch show details")
		return
	}

	writeJSON(w, http.StatusOK, details)
}

func (h *MediaHandler) GetEpisode(w http.ResponseWriter, r *http.Request) {
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

	details, err := h.client.GetEpisodeDetails(r.Context(), tmdbID, season, episode)
	if err != nil {
		h.logger.Error("tmdb episode details failed", "error", err,
			"tmdb_id", tmdbID, "season", season, "episode", episode)
		writeError(w, http.StatusBadGateway, "failed to fetch episode details")
		return
	}

	writeJSON(w, http.StatusOK, details)
}

func (h *MediaHandler) Trending(w http.ResponseWriter, r *http.Request) {
	results, err := h.client.GetTrending(r.Context())
	if err != nil {
		h.logger.Error("tmdb trending failed", "error", err)
		writeError(w, http.StatusBadGateway, "failed to fetch trending media")
		return
	}

	writeJSON(w, http.StatusOK, results)
}
