package httpapi

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
)

type RouterDeps struct {
	Logger      *slog.Logger
	AuthHandler *AuthHandler
	MediaHandler *MediaHandler
	CommentHandler *CommentHandler
	JWTSecret   string
}

func NewRouter(deps RouterDeps) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.ClientIPFromRemoteAddr)

	r.Get("/health", healthHandler)

	r.Group(func(r chi.Router) {
		r.Use(httprate.LimitBy(5, time.Minute, clientIPKey))
		r.Post("/signup", deps.AuthHandler.Signup)
		r.Post("/login", deps.AuthHandler.Login)
	})

	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware(deps.JWTSecret))
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
		r.Use(AuthMiddleware(deps.JWTSecret))
		r.Post("/comments", deps.CommentHandler.Create)
		r.Patch("/comments/{comment_id}", deps.CommentHandler.Update)
		r.Delete("/comments/{comment_id}", deps.CommentHandler.Delete)
	})

	return r
}

func clientIPKey(r *http.Request) (string, error) {
	return httprate.CanonicalizeIP(middleware.GetClientIP(r.Context())), nil
}