-- +goose Up
-- NULL = hasn't agreed to the Terms of Service yet; a timestamp = when they agreed.
-- Required before posting or editing comments (App Store guideline 1.2).
ALTER TABLE profiles ADD COLUMN terms_accepted_at TIMESTAMPTZ;

-- +goose Down
ALTER TABLE profiles DROP COLUMN IF EXISTS terms_accepted_at;
