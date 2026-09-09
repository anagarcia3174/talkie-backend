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
       u.username
FROM comments c
JOIN users u ON u.id = c.user_id
WHERE c.tmdb_media_type = 'movie'
  AND c.tmdb_id = $1
ORDER BY c.timestamp_seconds ASC;

-- name: ListCommentsByEpisode :many
SELECT c.id, c.user_id, c.tmdb_media_type, c.tmdb_id, c.season_number,
       c.episode_number, c.timestamp_seconds, c.body, c.created_at,
       u.username
FROM comments c
JOIN users u ON u.id = c.user_id
WHERE c.tmdb_media_type = 'tv'
  AND c.tmdb_id = $1
  AND c.season_number = $2
  AND c.episode_number = $3
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

