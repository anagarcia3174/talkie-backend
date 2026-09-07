-- name: CreateUser :one
INSERT INTO users (email, username, password_hash) VALUES ($1, $2, $3) RETURNING id, email, username, password_hash, created_at;

-- name: GetUserByEmail :one
SELECT id, email, username, password_hash, created_at FROM users WHERE email = $1;

-- name: GetUserByUsername :one
SELECT id, email, username, password_hash, created_at FROM users WHERE username = $1;

-- name: GetUserByID :one
SELECT id, email, username, password_hash, created_at FROM users WHERE id = $1;