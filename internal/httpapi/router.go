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
	ReportHandler  *ReportHandler
	BlockHandler   *BlockHandler
	Verifier       *auth.Verifier
	BanChecker     BanChecker
	TermsChecker   TermsChecker
	AllowedOrigin  string
	ClientIPHeader string
}

func NewRouter(deps RouterDeps) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	if deps.ClientIPHeader != "" {
		r.Use(middleware.ClientIPFromHeader(deps.ClientIPHeader))
	} else {
		r.Use(middleware.ClientIPFromRemoteAddr)
	}
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
		r.Get("/me/blocks", deps.BlockHandler.List)
	})

	r.Group(func(r chi.Router) {
		r.Use(httprate.LimitBy(60, time.Minute, clientIPKey))
		r.Get("/search", deps.MediaHandler.Search)
		r.Get("/trending", deps.MediaHandler.Trending)
		r.Get("/movies/{tmdb_id}", deps.MediaHandler.GetMovie)
		r.Get("/tv/{tmdb_id}", deps.MediaHandler.GetTVShow)
		r.Get("/tv/{tmdb_id}/season/{season}/episode/{episode}", deps.MediaHandler.GetEpisode)
		// Optional auth: signed-in viewers don't see comments from users they blocked.
		r.With(OptionalAuthMiddleware(deps.Verifier)).Get("/movies/{tmdb_id}/comments", deps.CommentHandler.ListByMovie)
		r.With(OptionalAuthMiddleware(deps.Verifier)).Get("/tv/{tmdb_id}/season/{season}/episode/{episode}/comments", deps.CommentHandler.ListByEpisode)
	})

	r.Group(func(r chi.Router) {
		r.Use(OriginMiddleware(deps.AllowedOrigin))
		r.Use(AuthMiddleware(deps.Verifier))
		notBanned := RequireNotBanned(deps.BanChecker)
		termsAccepted := RequireTermsAccepted(deps.TermsChecker)
		r.With(
			httprate.LimitBy(15, time.Minute, clientIPKey),
			notBanned,
			termsAccepted,
		).Post("/comments", deps.CommentHandler.Create)
		r.With(notBanned, termsAccepted).Patch("/comments/{comment_id}", deps.CommentHandler.Update)
		r.Delete("/comments/{comment_id}", deps.CommentHandler.Delete)
		r.With(
			httprate.LimitBy(10, time.Minute, clientIPKey),
			notBanned,
		).Post("/comments/{comment_id}/reports", deps.ReportHandler.Create)
		r.With(notBanned).Patch("/me", deps.AuthHandler.UpdateMe)
		r.With(
			httprate.LimitBy(10, time.Minute, clientIPKey),
		).Post("/me/terms", deps.AuthHandler.AcceptTerms)
		r.With(
			httprate.LimitBy(30, time.Minute, clientIPKey),
		).Post("/users/{user_id}/block", deps.BlockHandler.Block)
		r.Delete("/users/{user_id}/block", deps.BlockHandler.Unblock)
		r.With(
			httprate.LimitBy(5, time.Minute, clientIPKey),
			RequireRecentSignIn(5*time.Minute),
		).Delete("/me", deps.AuthHandler.DeleteMe)
	})

	return r

}

func clientIPKey(r *http.Request) (string, error) {
	return httprate.CanonicalizeIP(middleware.GetClientIP(r.Context())), nil
}
