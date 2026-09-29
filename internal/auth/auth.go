package auth

import (
	"context"
	"errors"
	"log/slog"
	"regexp"
	"strings"
	"time"

	db "github.com/anagarcia3174/talkie-backend/internal/db/generated"
	"github.com/anagarcia3174/talkie-backend/internal/moderation"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type User struct {
	ID            string    `json:"id"`
	Username      *string   `json:"username"`
	CreatedAt     time.Time `json:"created_at"`
	TermsAccepted bool      `json:"terms_accepted"`
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidUsername    = errors.New("invalid username")
	ErrUsernameTaken      = errors.New("username taken")
	ErrUsernameNotAllowed = errors.New("username not allowed")
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]{3,30}$`)

type UserDeleter interface {
	DeleteUser(ctx context.Context, userID string) error
}
type AppleRevoker interface {
	RevokeWithAuthorizationCode(ctx context.Context, code string) error
}

type Service struct {
	profiles db.Querier
	users    UserDeleter
	apple    AppleRevoker
	logger   *slog.Logger
}

func NewService(profiles db.Querier, users UserDeleter, apple AppleRevoker, logger *slog.Logger) *Service {
	return &Service{profiles: profiles, users: users, apple: apple, logger: logger}
}

func (s *Service) DeleteAccount(ctx context.Context, userID, appleAuthCode string) error {
	id, err := uuid.Parse(userID)
	if err != nil {
		return ErrInvalidCredentials
	}

	if appleAuthCode != "" {
		if err := s.apple.RevokeWithAuthorizationCode(ctx, appleAuthCode); err != nil {
			s.logger.Error("apple token revocation failed", "error", err, "user_id", id.String())
		}
	}

	return s.users.DeleteUser(ctx, id.String())
}

func (s *Service) GetCurrentUser(ctx context.Context, userID string) (User, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return User{}, ErrInvalidCredentials
	}

	row, err := s.profiles.GetProfileByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrInvalidCredentials
		}
		return User{}, err
	}

	return User{ID: row.ID.String(), Username: row.Username, CreatedAt: row.CreatedAt, TermsAccepted: row.TermsAccepted}, nil
}

func (s *Service) SetUsername(ctx context.Context, userID, username string) (User, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return User{}, ErrInvalidCredentials
	}

	username = strings.TrimSpace(username)
	if !usernamePattern.MatchString(username) {
		return User{}, ErrInvalidUsername
	}
	if moderation.ContainsObjectionable(username) {
		return User{}, ErrUsernameNotAllowed
	}

	taken, err := s.profiles.IsUsernameTaken(ctx, &username)
	if err != nil {
		return User{}, err
	}
	if taken {
		return User{}, ErrUsernameTaken
	}

	row, err := s.profiles.SetUsername(ctx, db.SetUsernameParams{
		ID:       id,
		Username: &username,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrInvalidCredentials
		}
		return User{}, err
	}

	return User{ID: row.ID.String(), Username: row.Username, CreatedAt: row.CreatedAt, TermsAccepted: row.TermsAccepted}, nil
}

func (s *Service) AcceptTerms(ctx context.Context, userID string) (User, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return User{}, ErrInvalidCredentials
	}

	row, err := s.profiles.AcceptTerms(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrInvalidCredentials
		}
		return User{}, err
	}

	return User{ID: row.ID.String(), Username: row.Username, CreatedAt: row.CreatedAt, TermsAccepted: row.TermsAccepted}, nil
}
