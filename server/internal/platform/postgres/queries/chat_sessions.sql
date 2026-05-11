-- name: CreateChatSession :exec
INSERT INTO chat_sessions (id, user_id, star_id, created_at)
VALUES ($1, $2, $3, $4);

-- name: GetChatSessionByID :one
SELECT id, user_id, star_id, created_at
FROM chat_sessions WHERE id = $1;
