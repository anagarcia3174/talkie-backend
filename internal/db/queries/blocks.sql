-- name: BlockUser :exec
INSERT INTO blocks (blocker_id, blocked_id)
VALUES ($1, $2)
ON CONFLICT (blocker_id, blocked_id) DO NOTHING;

-- name: UnblockUser :exec
DELETE FROM blocks
WHERE blocker_id = $1 AND blocked_id = $2;

-- name: ListBlockedUsers :many
SELECT p.id, p.username, b.created_at
FROM blocks b
JOIN profiles p ON p.id = b.blocked_id
WHERE b.blocker_id = $1
ORDER BY b.created_at DESC;
