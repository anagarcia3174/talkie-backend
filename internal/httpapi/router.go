package httpapi

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/anagarcia3174/talkie-backend/internal/auth"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
)

type RouterDeps struct {
	Logger         *slog.Logger
	AuthHandler    *AuthHandler
	MediaHandler   *MediaHandler
	CommentHandler *CommentHandler
	Verifier       *auth.Verifier
	AllowedOrigin  string
}

func NewRouter(deps RouterDeps) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.ClientIPFromRemoteAddr)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{deps.AllowedOrigin},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/health", healthHandler)

	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware(deps.Verifier))
		r.Get("/me", deps.AuthHandler.Me)
	})

	r.Group(func(r chi.Router) {
		r.Use(httprate.LimitBy(60, time.Minute, clientIPKey))
		r.Get("/search", deps.MediaHandler.Search)
		r.Get("/trending", deps.MediaHandler.Trending)
		r.Get("/movies/{tmdb_id}", deps.MediaHandler.GetMovie)
		r.Get("/tv/{tmdb_id}", deps.MediaHandler.GetTVShow)
		r.Get("/tv/{tmdb_id}/season/{season}/episode/{episode}", deps.MediaHandler.GetEpisode)
		r.Get("/movies/{tmdb_id}/comments", deps.CommentHandler.ListByMovie)
		r.Get("/tv/{tmdb_id}/season/{season}/episode/{episode}/comments", deps.CommentHandler.ListByEpisode)
	})

	r.Group(func(r chi.Router) {
		r.Use(OriginMiddleware(deps.AllowedOrigin))
		r.Use(AuthMiddleware(deps.Verifier))
		r.Post("/comments", deps.CommentHandler.Create)
		r.Patch("/comments/{comment_id}", deps.CommentHandler.Update)
		r.Delete("/comments/{comment_id}", deps.CommentHandler.Delete)
	})

	return r

}

func clientIPKey(r *http.Request) (string, error) {
	return httprate.CanonicalizeIP(middleware.GetClientIP(r.Context())), nil
}
