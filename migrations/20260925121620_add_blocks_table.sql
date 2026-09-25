-- +goose Up
CREATE TABLE IF NOT EXISTS blocks (
    blocker_id UUID NOT NULL,
    blocked_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    -- one row per pair: a user can only block someone once
    PRIMARY KEY (blocker_id, blocked_id),
    CONSTRAINT blocks_no_self_block CHECK (blocker_id <> blocked_id),

    CONSTRAINT blocks_blocker_id_profiles_id_fk
        FOREIGN KEY (blocker_id)
        REFERENCES profiles(id)
        ON DELETE CASCADE,
    CONSTRAINT blocks_blocked_id_profiles_id_fk
        FOREIGN KEY (blocked_id)
        REFERENCES profiles(id)
        ON DELETE CASCADE
);

-- The primary key covers lookups by blocker_id; this covers the cascade
-- when a blocked user's profile is deleted.
CREATE INDEX IF NOT EXISTS idx_blocks_blocked_id ON blocks (blocked_id);

-- RLS with no policies: only the backend's database role can access these
-- tables, not the public Supabase Data API. profiles, comments, and reports
-- already have RLS on via the dashboard; this makes the migrations match.
ALTER TABLE blocks ENABLE ROW LEVEL SECURITY;
ALTER TABLE profiles ENABLE ROW LEVEL SECURITY;
ALTER TABLE comments ENABLE ROW LEVEL SECURITY;
ALTER TABLE reports ENABLE ROW LEVEL SECURITY;

-- +goose Down
-- RLS on profiles/comments/reports is left on: it was on before this migration.
DROP INDEX IF EXISTS idx_blocks_blocked_id;
DROP TABLE IF EXISTS blocks;
