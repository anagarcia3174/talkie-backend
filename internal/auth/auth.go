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

// imports needed: "context", "errors", "time" (already have these from
// password.go/jwt.go in this same package), plus:
// - your internal/db package (for db.Querier, db.CreateUserParams, etc.)
// - "github.com/jackc/pgx/v5/pgconn" (for unique-violation inspection)

// --- Types ---
type User struct {
	ID        string
	Email     string
	Username  string
	CreatedAt time.Time
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
	jwtSecret string
}

func NewService(users db.Querier, jwtSecret string) *Service {
	return &Service{users: users, jwtSecret: jwtSecret}
}

func (s *Service) Signup(ctx context.Context, in SignupInput) (User, error) {
	if in.Email == "" || in.Username == "" || len(in.Password) < 12 {
		return User{}, ErrInvalidInput
	}
	hash, err := HashPassword(in.Password)
	if err != nil {
		return User{}, err
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
				return User{}, ErrEmailTaken
			case "users_username_key":
				return User{}, ErrUsernameTaken
			}
		}
		return User{}, err
	}
	return toUser(row), nil
}

func (s *Service) Login(ctx context.Context, in LoginInput) (User, string, error) {
	row, err := s.users.GetUserByEmail(ctx, in.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows){
			return User{}, "", ErrInvalidCredentials
		}
		return  User{}, "", err
	}

	match, err := CheckPassword(row.PasswordHash, in.Password)
	if err != nil {
		return User{}, "", err
	}
	if !match {
		return User{}, "", ErrInvalidCredentials
	}
	user := toUser(row)
	token, err := GenerateToken(user.ID, user.Email, s.jwtSecret)
	if err != nil {
		return User{}, "", err
	}
	return user, token, nil
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

func toUser(row db.User) User {
	return User{
		ID:        row.ID.String(),
		Email:     row.Email,
		Username:  row.Username,
		CreatedAt: row.CreatedAt,
	}
}
