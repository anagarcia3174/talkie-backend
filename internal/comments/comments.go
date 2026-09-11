package comments

import (
	"context"
	"errors"
	"time"

	db "github.com/anagarcia3174/talkie-backend/internal/db/generated"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type MediaType string

const (
	MediaTypeMovie MediaType = "movie"
	MediaTypeTV    MediaType = "tv"
)

type Comment struct {
	ID               string    `json:"id"`
	UserID           string    `json:"user_id"`
	TMDBMediaType    MediaType `json:"tmdb_media_type"`
	TMDBID           int32     `json:"tmdb_id"`
	SeasonNumber     *int32    `json:"season_number"`
	EpisodeNumber    *int32    `json:"episode_number"`
	TimestampSeconds int32     `json:"timestamp_seconds"`
	Body             string    `json:"body"`
	CreatedAt        time.Time `json:"created_at"`
}

type CommentWithUsername struct {
	ID               string    `json:"id"`
	UserID           string    `json:"user_id"`
	Username         string    `json:"username"`
	TMDBMediaType    MediaType `json:"tmdb_media_type"`
	TMDBID           int32     `json:"tmdb_id"`
	SeasonNumber     *int32    `json:"season_number"`
	EpisodeNumber    *int32    `json:"episode_number"`
	TimestampSeconds int32     `json:"timestamp_seconds"`
	Body             string    `json:"body"`
	CreatedAt        time.Time `json:"created_at"`
}

type CreateCommentInput struct {
	UserID           string
	TMDBMediaType    MediaType
	TMDBID           int32
	SeasonNumber     *int32
	EpisodeNumber    *int32
	TimestampSeconds int32
	Body             string
}

type UpdateCommentInput struct {
	TimestampSeconds int32
	Body             string
}

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrNotFound     = errors.New("comment not found")
	ErrForbidden    = errors.New("not your comment")
)

type Service struct {
	comments db.Querier
}

func NewService(comments db.Querier) *Service {
	return &Service{comments: comments}
}

func (s *Service) Create(ctx context.Context, in CreateCommentInput) (Comment, error) {
	if !validMediaReference(in.TMDBMediaType, in.TMDBID, in.SeasonNumber, in.EpisodeNumber) {
		return Comment{}, ErrInvalidInput
	}
	if in.TimestampSeconds < 0 || in.Body == "" {
		return Comment{}, ErrInvalidInput
	}

	userID, err := uuid.Parse(in.UserID)
	if err != nil {
		return Comment{}, ErrInvalidInput
	}

	row, err := s.comments.CreateComment(ctx, db.CreateCommentParams{
		UserID:           userID,
		TmdbMediaType:    db.MediaType(in.TMDBMediaType),
		TmdbID:           in.TMDBID,
		SeasonNumber:     in.SeasonNumber,
		EpisodeNumber:    in.EpisodeNumber,
		TimestampSeconds: in.TimestampSeconds,
		Body:             in.Body,
	})
	if err != nil {
		return Comment{}, err
	}
	return toComment(row), nil
}

func (s *Service) ListByMovie(ctx context.Context, tmdbID int32) ([]CommentWithUsername, error) {
	rows, err := s.comments.ListCommentsByMovie(ctx, tmdbID)
	if err != nil {
		return nil, err
	}
	out := make([]CommentWithUsername, 0, len(rows))
	for _, row := range rows {
		out = append(out, CommentWithUsername{
			ID:               row.ID.String(),
			UserID:           row.UserID.String(),
			Username:         row.Username,
			TMDBMediaType:    MediaType(row.TmdbMediaType),
			TMDBID:           row.TmdbID,
			SeasonNumber:     row.SeasonNumber,
			EpisodeNumber:    row.EpisodeNumber,
			TimestampSeconds: row.TimestampSeconds,
			Body:             row.Body,
			CreatedAt:        row.CreatedAt,
		})
	}
	return out, nil
}

func (s *Service) ListByEpisode(ctx context.Context, tmdbID, season, episode int32) ([]CommentWithUsername, error) {
	rows, err := s.comments.ListCommentsByEpisode(ctx, db.ListCommentsByEpisodeParams{
		TmdbID:        tmdbID,
		SeasonNumber:  &season,
		EpisodeNumber: &episode,
	})
	if err != nil {
		return nil, err
	}
	out := make([]CommentWithUsername, 0, len(rows))
	for _, row := range rows {
		out = append(out, CommentWithUsername{
			ID:               row.ID.String(),
			UserID:           row.UserID.String(),
			Username:         row.Username,
			TMDBMediaType:    MediaType(row.TmdbMediaType),
			TMDBID:           row.TmdbID,
			SeasonNumber:     row.SeasonNumber,
			EpisodeNumber:    row.EpisodeNumber,
			TimestampSeconds: row.TimestampSeconds,
			Body:             row.Body,
			CreatedAt:        row.CreatedAt,
		})
	}
	return out, nil
}

func (s *Service) Update(ctx context.Context, commentID, userID string, in UpdateCommentInput) (Comment, error) {
	if in.TimestampSeconds < 0 || in.Body == "" {
		return Comment{}, ErrInvalidInput
	}

	id, err := s.checkOwnership(ctx, commentID, userID)
	if err != nil {
		return Comment{}, err
	}

	row, err := s.comments.UpdateComment(ctx, db.UpdateCommentParams{
		ID:               id,
		Body:             in.Body,
		TimestampSeconds: in.TimestampSeconds,
	})
	if err != nil {
		return Comment{}, err
	}
	return toComment(row), nil
}

func (s *Service) Delete(ctx context.Context, commentID, userID string) error {
	id, err := s.checkOwnership(ctx, commentID, userID)
	if err != nil {
		return err
	}
	return s.comments.DeleteComment(ctx, id)
}

func (s *Service) checkOwnership(ctx context.Context, commentID, userID string) (uuid.UUID, error) {
	id, err := uuid.Parse(commentID)
	if err != nil {
		return uuid.UUID{}, ErrNotFound
	}

	row, err := s.comments.GetCommentByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.UUID{}, ErrNotFound
		}
		return uuid.UUID{}, err
	}

	if row.UserID.String() != userID {
		return uuid.UUID{}, ErrForbidden
	}

	return id, nil
}

func toComment(row db.Comment) Comment {
	return Comment{
		ID:               row.ID.String(),
		UserID:           row.UserID.String(),
		TMDBMediaType:    MediaType(row.TmdbMediaType),
		TMDBID:           row.TmdbID,
		SeasonNumber:     row.SeasonNumber,
		EpisodeNumber:    row.EpisodeNumber,
		TimestampSeconds: row.TimestampSeconds,
		Body:             row.Body,
		CreatedAt:        row.CreatedAt,
	}
}

func toComments(rows []db.Comment) []Comment {
	comments := make([]Comment, 0, len(rows))
	for _, row := range rows {
		comments = append(comments, toComment(row))
	}
	return comments
}

func validMediaReference(mediaType MediaType, tmdbID int32, season, episode *int32) bool {
	if mediaType != MediaTypeMovie && mediaType != MediaTypeTV {
		return false
	}
	if tmdbID <= 0 {
		return false
	}
	switch mediaType {
	case MediaTypeTV:
		if season == nil || episode == nil || *season < 0 || *episode < 0 {
			return false
		}
	case MediaTypeMovie:
		if season != nil || episode != nil {
			return false
		}
	}
	return true
}
