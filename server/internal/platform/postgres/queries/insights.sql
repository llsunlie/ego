-- name: CreateInsight :exec
INSERT INTO insights (id, user_id, moment_id, echo_id, text, related_moment_ids, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7);

-- name: GetInsightByMomentID :one
SELECT id, user_id, moment_id, echo_id, text, related_moment_ids, created_at
FROM insights WHERE moment_id = $1;

-- name: GetInsightByEchoID :one
SELECT id, user_id, moment_id, echo_id, text, related_moment_ids, created_at
FROM insights WHERE echo_id = $1;
