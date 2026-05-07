-- name: GetUserByAccount :one
SELECT id, password_hash FROM users WHERE account = $1;

-- name: CreateUser :exec
INSERT INTO users (id, account, password_hash, created_at) VALUES ($1, $2, $3, $4);
