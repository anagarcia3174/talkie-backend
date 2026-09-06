-- +goose Up
SELECT 'up SQL query';
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE TYPE media_type AS ENUM ('movie', 'tv');
CREATE TYPE report_status AS ENUM ('open', 'reviewed', 'dismissed');
CREATE TYPE report_reason AS ENUM ('spam', 'harassment', 'inappropriate', 'spoiler', 'other');

CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT NOT NULL UNIQUE,
    username VARCHAR(30) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    tmdb_media_type media_type NOT NULL,
    tmdb_id INT NOT NULL,
    season_number SMALLINT,
    episode_number SMALLINT,
    timestamp_seconds INT NOT NULL CHECK (timestamp_seconds >= 0),
    body TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT comments_user_id_users_id_fk
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    comment_id UUID NOT NULL,
    reporter_user_id UUID NOT NULL,
    reason report_reason NOT NULL,
    reason_detail TEXT,
    report_status report_status DEFAULT 'open' NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(comment_id, reporter_user_id),

    CONSTRAINT reports_comment_id_comments_id_fk
        FOREIGN KEY (comment_id)
        REFERENCES comments(id)
        ON DELETE CASCADE,
    CONSTRAINT reports_reporter_user_id_users_id_fk
        FOREIGN KEY (reporter_user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_comments_media_timestamp
    ON comments (tmdb_media_type, tmdb_id, season_number, episode_number, timestamp_seconds);

CREATE INDEX IF NOT EXISTS idx_comments_user_id
    ON comments (user_id);

CREATE INDEX IF NOT EXISTS idx_reports_reporter_user_id
    ON reports (reporter_user_id);
-- +goose Down
SELECT 'down SQL query';

DROP INDEX IF EXISTS idx_comments_media_timestamp;
DROP INDEX IF EXISTS idx_comments_user_id;
DROP INDEX IF EXISTS idx_reports_reporter_user_id;

DROP TABLE IF EXISTS reports;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS media_type;
DROP TYPE IF EXISTS report_status;