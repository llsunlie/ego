-- name: CreateTrace :exec
INSERT INTO traces (id, user_id, motivation, stashed, created_at)
VALUES ($1, $2, $3, $4, $5);

-- name: GetTraceByID :one
SELECT id, user_id, motivation, stashed, created_at
FROM traces WHERE id = $1;

-- name: UpdateTrace :exec
UPDATE traces SET stashed = $2 WHERE id = $1;

-- name: DeleteTrace :exec
DELETE FROM traces WHERE id = $1;

-- name: ListTracesByUserIDCursor :many
SELECT id, user_id, motivation, stashed, created_at
FROM traces
WHERE user_id = $1 AND created_at < sqlc.arg(cursor_time)::timestamptz
ORDER BY created_at DESC
LIMIT $2;
