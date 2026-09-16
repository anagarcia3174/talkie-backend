-- +goose Up
ALTER TABLE profiles ALTER COLUMN username DROP NOT NULL;

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION handle_new_user()
RETURNS TRIGGER AS $$
BEGIN
  INSERT INTO profiles (id)
  VALUES (NEW.id)
  ON CONFLICT (id) DO NOTHING;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
-- +goose StatementEnd

-- +goose Down
UPDATE profiles SET username = 'user_' || substr(id::text, 1, 8) WHERE username IS NULL;

ALTER TABLE profiles ALTER COLUMN username SET NOT NULL;

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION handle_new_user()
RETURNS TRIGGER AS $$
BEGIN
  INSERT INTO profiles (id, username)
  VALUES (NEW.id, COALESCE(split_part(NEW.email, '@', 1), 'user') || '_' || substr(NEW.id::text, 1, 8))
  ON CONFLICT (id) DO NOTHING;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
-- +goose StatementEnd