-- name: CreateComment :one
INSERT INTO comments (
    user_id, tmdb_media_type, tmdb_id, season_number, episode_number,
    timestamp_seconds, body
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ListCommentsByMovie :many
SELECT c.id, c.user_id, c.tmdb_media_type, c.tmdb_id, c.season_number,
       c.episode_number, c.timestamp_seconds, c.body, c.created_at,
       p.username
FROM comments c
JOIN profiles p ON p.id = c.user_id
WHERE c.tmdb_media_type = 'movie'
  AND c.tmdb_id = sqlc.arg(tmdb_id)
  AND p.banned_at IS NULL
  -- auto-hide: 2+ non-dismissed reports (one per user, via the reports unique constraint)
  AND (
    SELECT count(*) FROM reports r
    WHERE r.comment_id = c.id AND r.report_status <> 'dismissed'
  ) < 2
  -- hide comments from users the viewer blocked (viewer_id is NULL when signed out)
  AND NOT EXISTS (
    SELECT 1 FROM blocks b
    WHERE b.blocker_id = sqlc.narg(viewer_id)::uuid
      AND b.blocked_id = c.user_id
  )
ORDER BY c.timestamp_seconds ASC;

-- name: ListCommentsByEpisode :many
SELECT c.id, c.user_id, c.tmdb_media_type, c.tmdb_id, c.season_number,
       c.episode_number, c.timestamp_seconds, c.body, c.created_at,
       p.username
FROM comments c
JOIN profiles p ON p.id = c.user_id
WHERE c.tmdb_media_type = 'tv'
  AND c.tmdb_id = sqlc.arg(tmdb_id)
  AND c.season_number = sqlc.arg(season_number)
  AND c.episode_number = sqlc.arg(episode_number)
  AND p.banned_at IS NULL
  -- auto-hide: 2+ non-dismissed reports (one per user, via the reports unique constraint)
  AND (
    SELECT count(*) FROM reports r
    WHERE r.comment_id = c.id AND r.report_status <> 'dismissed'
  ) < 2
  -- hide comments from users the viewer blocked (viewer_id is NULL when signed out)
  AND NOT EXISTS (
    SELECT 1 FROM blocks b
    WHERE b.blocker_id = sqlc.narg(viewer_id)::uuid
      AND b.blocked_id = c.user_id
  )
ORDER BY c.timestamp_seconds ASC;

-- name: GetCommentByID :one
SELECT * FROM comments
WHERE id = $1;

-- name: UpdateComment :one
UPDATE comments
SET body = $2, timestamp_seconds = $3
WHERE id = $1
RETURNING *;

-- name: DeleteComment :exec
DELETE FROM comments
WHERE id = $1;