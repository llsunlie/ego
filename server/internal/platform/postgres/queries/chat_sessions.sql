-- name: CreateChatSession :exec
INSERT INTO chat_sessions (id, user_id, star_id, context_moment_ids, created_at)
VALUES ($1, $2, $3, $4, $5);

-- name: GetChatSessionByID :one
SELECT id, user_id, star_id, context_moment_ids, created_at
FROM chat_sessions WHERE id = $1;
