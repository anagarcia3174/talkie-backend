package reports

import (
	"context"
	"errors"
	"strings"
	"time"
	"unicode/utf8"

	db "github.com/anagarcia3174/talkie-backend/internal/db/generated"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const MaxDetailLength = 500

type Reason string

const (
	ReasonSpam          Reason = "spam"
	ReasonHarassment    Reason = "harassment"
	ReasonInappropriate Reason = "inappropriate"
	ReasonSpoiler       Reason = "spoiler"
	ReasonOther         Reason = "other"
)

type Report struct {
	ID        string    `json:"id"`
	CommentID string    `json:"comment_id"`
	Reason    Reason    `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateReportInput struct {
	CommentID string
	UserID    string
	Reason    Reason
	Detail    string
}

var (
	ErrInvalidInput    = errors.New("invalid input")
	ErrNotFound        = errors.New("comment not found")
	ErrOwnComment      = errors.New("cannot report your own comment")
	ErrAlreadyReported = errors.New("comment already reported")
)

type Service struct {
	q db.Querier
}

func NewService(q db.Querier) *Service {
	return &Service{q: q}
}

func (s *Service) Create(ctx context.Context, in CreateReportInput) (Report, error) {
	if !validReason(in.Reason) {
		return Report{}, ErrInvalidInput
	}
	detail := strings.TrimSpace(in.Detail)
	if utf8.RuneCountInString(detail) > MaxDetailLength {
		return Report{}, ErrInvalidInput
	}

	reporterID, err := uuid.Parse(in.UserID)
	if err != nil {
		return Report{}, ErrInvalidInput
	}
	commentID, err := uuid.Parse(in.CommentID)
	if err != nil {
		return Report{}, ErrNotFound
	}

	comment, err := s.q.GetCommentByID(ctx, commentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Report{}, ErrNotFound
		}
		return Report{}, err
	}
	if comment.UserID == reporterID {
		return Report{}, ErrOwnComment
	}

	row, err := s.q.CreateReport(ctx, db.CreateReportParams{
		CommentID:      commentID,
		ReporterUserID: reporterID,
		Reason:         db.ReportReason(in.Reason),
		ReasonDetail:   pgtype.Text{String: detail, Valid: detail != ""},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Report{}, ErrAlreadyReported
		}
		return Report{}, err
	}

	return Report{
		ID:        row.ID.String(),
		CommentID: row.CommentID.String(),
		Reason:    Reason(row.Reason),
		CreatedAt: row.CreatedAt,
	}, nil
}

func validReason(r Reason) bool {
	switch r {
	case ReasonSpam, ReasonHarassment, ReasonInappropriate, ReasonSpoiler, ReasonOther:
		return true
	}
	return false
}
