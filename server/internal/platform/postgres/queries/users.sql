-- name: GetUserByPhone :one
SELECT id, password_hash FROM users WHERE phone = $1;

-- name: CreateUser :exec
INSERT INTO users (id, phone, password_hash, created_at) VALUES ($1, $2, $3, $4);

-- name: GetUserByID :one
SELECT phone, created_at FROM users WHERE id = $1;

-- name: UpdateUserPassword :exec
UPDATE users SET password_hash = $2 WHERE id = $1;
