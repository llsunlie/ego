-- name: CreateEcho :exec
INSERT INTO echos (id, moment_id, user_id, matched_moment_ids, similarities, created_at)
VALUES ($1, $2, $3, $4, $5, $6);

-- name: GetEchoByMomentID :one
SELECT id, moment_id, user_id, matched_moment_ids, similarities, created_at
FROM echos WHERE moment_id = $1;

-- name: ListEchosByMomentIDs :many
SELECT id, moment_id, user_id, matched_moment_ids, similarities, created_at
FROM echos WHERE moment_id = ANY($1::uuid[]);
