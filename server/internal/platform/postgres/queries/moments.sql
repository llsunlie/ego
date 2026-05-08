-- name: CreateMoment :exec
INSERT INTO moments (id, trace_id, user_id, content, embedding, connected, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetMomentByID :one
SELECT id, trace_id, user_id, content, embedding, connected, created_at
FROM moments WHERE id = $1;

-- name: ListMomentsByTraceID :many
SELECT id, trace_id, user_id, content, embedding, connected, created_at
FROM moments WHERE trace_id = $1 ORDER BY created_at ASC;

-- name: ListMomentsByUserID :many
SELECT id, trace_id, user_id, content, embedding, connected, created_at
FROM moments WHERE user_id = $1 ORDER BY created_at DESC;

-- name: ListMomentsByUserIDCursor :many
SELECT id, trace_id, user_id, content, embedding, connected, created_at
FROM moments
WHERE user_id = $1 AND created_at < sqlc.arg(cursor_time)::timestamptz
ORDER BY created_at DESC
LIMIT $2;

-- name: RandomMomentsByUserID :many
SELECT id, trace_id, user_id, content, embedding, connected, created_at
FROM moments WHERE user_id = $1 ORDER BY random() LIMIT $2;

-- name: CountMomentsByUserID :one
SELECT COUNT(*) FROM moments WHERE user_id = $1;
