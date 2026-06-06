-- name: GetUserByPhone :one
SELECT id, password_hash FROM users WHERE phone = $1;

-- name: CreateUser :exec
INSERT INTO users (id, phone, password_hash, created_at) VALUES ($1, $2, $3, $4);
