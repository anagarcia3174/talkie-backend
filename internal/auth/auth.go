package auth

import (
	"context"
	"errors"
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

var ErrInvalidCredentials = errors.New("invalid credentials")

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