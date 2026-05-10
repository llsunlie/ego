-- name: CreateMoment :exec
INSERT INTO moments (id, trace_id, user_id, content, embeddings, created_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetMomentByID :one
SELECT id, trace_id, user_id, content, embeddings, created_at
FROM moments WHERE id = $1;

-- name: ListMomentsByTraceID :many
SELECT id, trace_id, user_id, content, embeddings, created_at
FROM moments WHERE trace_id = $1 ORDER BY created_at ASC;

-- name: ListMomentsByUserID :many
SELECT id, trace_id, user_id, content, embeddings, created_at
FROM moments WHERE user_id = $1 ORDER BY created_at DESC;

-- name: ListMomentsByUserIDCursor :many
SELECT id, trace_id, user_id, content, embeddings, created_at
FROM moments
WHERE user_id = $1 AND created_at < sqlc.arg(cursor_time)::timestamptz
ORDER BY created_at DESC
LIMIT $2;

-- name: RandomMomentsByUserID :many
SELECT id, trace_id, user_id, content, embeddings, created_at
FROM moments WHERE user_id = $1 ORDER BY random() LIMIT $2;

-- name: CountMomentsByUserID :one
SELECT COUNT(*) FROM moments WHERE user_id = $1;

-- name: FirstMomentsByTraceIDs :many
SELECT DISTINCT ON (trace_id) id, trace_id, user_id, content, embeddings, created_at
FROM moments WHERE trace_id = ANY($1::UUID[]) ORDER BY trace_id, created_at ASC;

-- name: ListMomentsByIDs :many
SELECT id, trace_id, user_id, content, embeddings, created_at
FROM moments WHERE id = ANY($1::UUID[]);
