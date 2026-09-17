package auth

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	db "github.com/anagarcia3174/talkie-backend/internal/db/generated"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type User struct {
	ID        string    `json:"id"`
	Username  *string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidUsername    = errors.New("invalid username")
	ErrUsernameTaken      = errors.New("username taken")
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]{3,30}$`)

type Service struct {
	profiles db.Querier
}

func NewService(profiles db.Querier) *Service {
	return &Service{profiles: profiles}
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

	return User{ID: row.ID.String(), Username: row.Username, CreatedAt: row.CreatedAt}, nil
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

	return User{ID: row.ID.String(), Username: row.Username, CreatedAt: row.CreatedAt}, nil
}