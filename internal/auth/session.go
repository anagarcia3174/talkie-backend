package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	db "github.com/anagarcia3174/talkie-backend/internal/db/generated"
	"github.com/google/uuid"
)

const SessionDuration = 7 * 24 * time.Hour

var ErrInvalidSession = errors.New("invalid session")

type SessionService struct {
	sessions db.Querier
}

func NewSessionService(sessions db.Querier) *SessionService {
	return &SessionService{sessions: sessions}
}

func (s *SessionService) Create(ctx context.Context, userID string) (string, time.Time, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return "", time.Time{}, ErrInvalidSession
	}

	token, err := generateSessionToken()
	if err != nil {
		return "", time.Time{}, err
	}

	expiresAt := time.Now().Add(SessionDuration)
	_, err = s.sessions.CreateSession(ctx, db.CreateSessionParams{
		UserID:    uid,
		TokenHash: hashSessionToken(token),
		ExpiresAt: expiresAt,
	})
	if err != nil {
		return "", time.Time{}, err
	}

	return token, expiresAt, nil
}

func (s *SessionService) Validate(ctx context.Context, token string) (string, error) {
	row, err := s.sessions.GetActiveSessionByTokenHash(ctx, hashSessionToken(token))
	if err != nil {
		return "", ErrInvalidSession
	}
	return row.UserID.String(), nil
}

func (s *SessionService) Revoke(ctx context.Context, token string) error {
	return s.sessions.RevokeSessionByTokenHash(ctx, hashSessionToken(token))
}

func (s *SessionService) RevokeAll(ctx context.Context, userID string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return ErrInvalidSession
	}
	return s.sessions.RevokeAllSessionsForUser(ctx, uid)
}

func generateSessionToken() (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func hashSessionToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}