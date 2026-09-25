package blocks

import (
	"context"
	"errors"
	"time"

	db "github.com/anagarcia3174/talkie-backend/internal/db/generated"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type BlockedUser struct {
	ID        string    `json:"id"`
	Username  *string   `json:"username"`
	BlockedAt time.Time `json:"blocked_at"`
}

var (
	ErrInvalidInput    = errors.New("invalid input")
	ErrUserNotFound    = errors.New("user not found")
	ErrCannotBlockSelf = errors.New("cannot block yourself")
)

type Service struct {
	q db.Querier
}

func NewService(q db.Querier) *Service {
	return &Service{q: q}
}

// Block hides blockedID's comments from blockerID. Blocking someone who is
// already blocked succeeds without changes.
func (s *Service) Block(ctx context.Context, blockerID, blockedID string) error {
	blocker, blocked, err := parseIDs(blockerID, blockedID)
	if err != nil {
		return err
	}
	if blocker == blocked {
		return ErrCannotBlockSelf
	}

	if _, err := s.q.GetProfileByID(ctx, blocked); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrUserNotFound
		}
		return err
	}

	return s.q.BlockUser(ctx, db.BlockUserParams{BlockerID: blocker, BlockedID: blocked})
}

// Unblock removes a block. Unblocking someone who isn't blocked succeeds.
func (s *Service) Unblock(ctx context.Context, blockerID, blockedID string) error {
	blocker, blocked, err := parseIDs(blockerID, blockedID)
	if err != nil {
		return err
	}
	return s.q.UnblockUser(ctx, db.UnblockUserParams{BlockerID: blocker, BlockedID: blocked})
}

func (s *Service) List(ctx context.Context, blockerID string) ([]BlockedUser, error) {
	blocker, err := uuid.Parse(blockerID)
	if err != nil {
		return nil, ErrInvalidInput
	}

	rows, err := s.q.ListBlockedUsers(ctx, blocker)
	if err != nil {
		return nil, err
	}
	out := make([]BlockedUser, 0, len(rows))
	for _, row := range rows {
		out = append(out, BlockedUser{
			ID:        row.ID.String(),
			Username:  row.Username,
			BlockedAt: row.CreatedAt,
		})
	}
	return out, nil
}

func parseIDs(blockerID, blockedID string) (uuid.UUID, uuid.UUID, error) {
	blocker, err := uuid.Parse(blockerID)
	if err != nil {
		return uuid.UUID{}, uuid.UUID{}, ErrInvalidInput
	}
	blocked, err := uuid.Parse(blockedID)
	if err != nil {
		return uuid.UUID{}, uuid.UUID{}, ErrUserNotFound
	}
	return blocker, blocked, nil
}
