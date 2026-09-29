-- name: GetProfileByID :one
SELECT id, username, created_at, (terms_accepted_at IS NOT NULL)::boolean AS terms_accepted
FROM profiles
WHERE id = $1;

-- name: IsUsernameTaken :one
SELECT EXISTS (
    SELECT 1 FROM profiles WHERE username = $1
) AS taken;

-- name: SetUsername :one
UPDATE profiles
SET username = $2
WHERE id = $1
RETURNING id, username, created_at, (terms_accepted_at IS NOT NULL)::boolean AS terms_accepted;

-- name: IsUserBanned :one
SELECT EXISTS (
    SELECT 1 FROM profiles WHERE id = $1 AND banned_at IS NOT NULL
) AS banned;

-- name: HasAcceptedTerms :one
SELECT EXISTS (
    SELECT 1 FROM profiles WHERE id = $1 AND terms_accepted_at IS NOT NULL
) AS accepted;

-- name: AcceptTerms :one
UPDATE profiles
SET terms_accepted_at = now()
WHERE id = $1
RETURNING id, username, created_at, (terms_accepted_at IS NOT NULL)::boolean AS terms_accepted;
