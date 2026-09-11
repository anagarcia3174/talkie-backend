package auth

import (
	"context"
	"errors"
	"time"

	db "github.com/anagarcia3174/talkie-backend/internal/db/generated"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)



type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

type SignupInput struct {
	Email    string
	Username string
	Password string
}
type LoginInput struct {
	Email    string
	Password string
}

var (
	ErrInvalidInput       = errors.New("invalid input")
	ErrEmailTaken         = errors.New("email already taken")
	ErrUsernameTaken      = errors.New("username already taken")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

type Service struct {
	users     db.Querier
	sessions *SessionService
}

func NewService(users db.Querier, sessions *SessionService) *Service {
	return &Service{users: users, sessions: sessions}
}

func (s *Service) Signup(ctx context.Context, in SignupInput) (User, string, time.Time, error) {
	if in.Email == "" || in.Username == "" || len(in.Password) < 12 {
		return User{}, "", time.Time{}, ErrInvalidInput
	}
	hash, err := HashPassword(in.Password)
	if err != nil {
		return User{}, "", time.Time{}, err
	}
	row, err := s.users.CreateUser(ctx, db.CreateUserParams{
		Email:        in.Email,
		Username:     in.Username,
		PasswordHash: hash,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "users_email_key":
				return User{}, "", time.Time{}, ErrEmailTaken
			case "users_username_key":
				return User{}, "", time.Time{}, ErrUsernameTaken
			}
		}
		return User{}, "", time.Time{}, err
	}

	user := toUser(row)
	token, expiresAt, err := s.sessions.Create(ctx, user.ID)
	if err != nil {
		return User{}, "", time.Time{}, err
	}
	return user, token, expiresAt, nil
}

func (s *Service) Login(ctx context.Context, in LoginInput) (User, string, time.Time, error) {
	row, err := s.users.GetUserByEmail(ctx, in.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, "", time.Time{}, ErrInvalidCredentials
		}
		return User{}, "", time.Time{}, err
	}

	match, err := CheckPassword(row.PasswordHash, in.Password)
	if err != nil {
		return User{}, "", time.Time{}, err
	}
	if !match {
		return User{}, "", time.Time{}, ErrInvalidCredentials
	}

	user := toUser(row)
	token, expiresAt, err := s.sessions.Create(ctx, user.ID)
	if err != nil {
		return User{}, "", time.Time{}, err
	}
	return user, token, expiresAt, nil
}

func (s *Service) GetCurrentUser(ctx context.Context, userID string) (User, error) {
	id, err := uuid.Parse(userID)
	if err != nil {
		return User{}, ErrInvalidCredentials
	}

	row, err := s.users.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrInvalidCredentials
		}
		return User{}, err
	}

	return toUser(row), nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	return s.sessions.Revoke(ctx, token)
}

func toUser(row db.User) User {
	return User{
		ID:        row.ID.String(),
		Email:     row.Email,
		Username:  row.Username,
		CreatedAt: row.CreatedAt,
	}
}
