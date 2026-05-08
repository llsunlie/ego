-- name: CreateTrace :exec
INSERT INTO traces (id, user_id, topic, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5);

-- name: GetTraceByID :one
SELECT id, user_id, topic, created_at, updated_at
FROM traces WHERE id = $1;

-- name: UpdateTrace :exec
UPDATE traces SET topic = $2, updated_at = $3 WHERE id = $1;

-- name: DeleteTrace :exec
DELETE FROM traces WHERE id = $1;
