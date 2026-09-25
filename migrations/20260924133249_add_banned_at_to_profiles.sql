-- +goose Up
-- NULL = active account; a timestamp = banned at that time.
ALTER TABLE profiles ADD COLUMN banned_at TIMESTAMPTZ;

-- +goose Down
ALTER TABLE profiles DROP COLUMN IF EXISTS banned_at;
