-- name: CreateChatMessage :exec
INSERT INTO chat_messages (id, user_id, session_id, role, content, referenced_moments, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: ListMessagesBySessionID :many
SELECT id, user_id, session_id, role, content, referenced_moments, created_at
FROM chat_messages WHERE session_id = $1 ORDER BY created_at;
