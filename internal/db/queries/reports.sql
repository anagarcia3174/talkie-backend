-- name: CreateReport :one
INSERT INTO reports (comment_id, reporter_user_id, reason, reason_detail)
VALUES ($1, $2, $3, $4)
ON CONFLICT (comment_id, reporter_user_id) DO NOTHING
RETURNING *;
