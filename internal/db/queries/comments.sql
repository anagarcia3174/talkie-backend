-- name: CreateComment :one
INSERT INTO comments (
    user_id, tmdb_media_type, tmdb_id, season_number, episode_number,
    timestamp_seconds, body
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: ListCommentsByMovie :many
SELECT * FROM comments
WHERE tmdb_media_type = 'movie'
  AND tmdb_id = $1
ORDER BY timestamp_seconds ASC;

-- name: ListCommentsByEpisode :many
SELECT * FROM comments
WHERE tmdb_media_type = 'tv'
  AND tmdb_id = $1
  AND season_number = $2
  AND episode_number = $3
ORDER BY timestamp_seconds ASC;

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

