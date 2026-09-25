package httpapi

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/anagarcia3174/talkie-backend/internal/auth"
	"github.com/google/uuid"
)

type contextKey string

const (
	userIDContextKey     contextKey = "userID"
	lastSignInContextKey contextKey = "lastSignIn"
)

func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDContextKey).(string)
	return userID, ok
}

func AuthMiddleware(verifier *auth.Verifier) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			token, ok := strings.CutPrefix(header, "Bearer ")
			if !ok || token == "" {
				writeError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			claims, err := verifier.Verify(token)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			ctx := context.WithValue(r.Context(), userIDContextKey, claims.Subject)
			ctx = context.WithValue(ctx, lastSignInContextKey, claims.LastSignInAt())
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func OriginMiddleware(allowedOrigin string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" && origin != allowedOrigin {
				writeError(w, http.StatusForbidden, "invalid origin")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequireRecentSignIn(window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			lastSignIn, ok := r.Context().Value(lastSignInContextKey).(time.Time)
			if !ok || lastSignIn.IsZero() || time.Since(lastSignIn) > window {
				writeError(w, http.StatusForbidden, "recent sign-in required")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

type BanChecker interface {
	IsUserBanned(ctx context.Context, id uuid.UUID) (bool, error)
}

func RequireNotBanned(checker BanChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, ok := UserIDFromContext(r.Context())
			if !ok {
				writeError(w, http.StatusUnauthorized, "unauthorized")
				return
			}
			id, err := uuid.Parse(userID)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			banned, err := checker.IsUserBanned(r.Context(), id)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "internal server error")
				return
			}
			if banned {
				writeError(w, http.StatusForbidden, "account suspended")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
