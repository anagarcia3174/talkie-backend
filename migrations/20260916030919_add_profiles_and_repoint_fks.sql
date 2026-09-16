-- +goose Up
SELECT 'up SQL query';

CREATE TABLE IF NOT EXISTS public.profiles (
    id UUID PRIMARY KEY REFERENCES auth.users(id) ON DELETE CASCADE,
    username VARCHAR(30) UNIQUE NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION public.handle_new_user()
RETURNS trigger AS $$
BEGIN
  INSERT INTO public.profiles (id, username)
  VALUES (
    NEW.id,
    COALESCE(NEW.raw_user_meta_data->>'preferred_username', split_part(NEW.email, '@', 1))
  )
  ON CONFLICT (id) DO NOTHING;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
-- +goose StatementEnd

CREATE TRIGGER on_auth_user_created
  AFTER INSERT ON auth.users
  FOR EACH ROW EXECUTE FUNCTION public.handle_new_user();

ALTER TABLE comments DROP CONSTRAINT comments_user_id_users_id_fk;
ALTER TABLE comments ADD CONSTRAINT comments_user_id_profiles_id_fk
    FOREIGN KEY (user_id) REFERENCES public.profiles(id) ON DELETE CASCADE;

ALTER TABLE reports DROP CONSTRAINT reports_reporter_user_id_users_id_fk;
ALTER TABLE reports ADD CONSTRAINT reports_reporter_user_id_profiles_id_fk
    FOREIGN KEY (reporter_user_id) REFERENCES public.profiles(id) ON DELETE CASCADE;

DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;

-- +goose Down
SELECT 'down SQL query';

-- Recreate the tables Up dropped, so the FKs below can point back at them.
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    username VARCHAR(30) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    token_hash TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,

    CONSTRAINT sessions_user_id_users_id_fk
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions (user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions (expires_at);

-- Repoint comments/reports back at the local users table.
ALTER TABLE comments DROP CONSTRAINT comments_user_id_profiles_id_fk;
ALTER TABLE comments ADD CONSTRAINT comments_user_id_users_id_fk
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;

ALTER TABLE reports DROP CONSTRAINT reports_reporter_user_id_profiles_id_fk;
ALTER TABLE reports ADD CONSTRAINT reports_reporter_user_id_users_id_fk
    FOREIGN KEY (reporter_user_id) REFERENCES users(id) ON DELETE CASCADE;

DROP TRIGGER IF EXISTS on_auth_user_created ON auth.users;
DROP FUNCTION IF EXISTS public.handle_new_user();
DROP TABLE IF EXISTS public.profiles;