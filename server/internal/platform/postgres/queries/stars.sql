-- name: CreateStar :exec
INSERT INTO stars (id, user_id, trace_id, topic, created_at)
VALUES ($1, $2, $3, $4, $5);

-- name: GetStarByID :one
SELECT id, user_id, trace_id, topic, created_at
FROM stars WHERE id = $1;

-- name: GetStarByTraceID :one
SELECT id, user_id, trace_id, topic, created_at
FROM stars WHERE trace_id = $1;

-- name: ListStarsByIDs :many
SELECT id, user_id, trace_id, topic, created_at
FROM stars WHERE id = ANY($1::UUID[]);
