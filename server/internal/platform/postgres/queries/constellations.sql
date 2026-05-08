-- name: CreateConstellation :exec
INSERT INTO constellations (id, user_id, name, constellation_insight, star_ids, topic_prompts, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8);

-- name: UpdateConstellation :exec
UPDATE constellations SET name = $2, constellation_insight = $3, star_ids = $4, topic_prompts = $5, updated_at = $6
WHERE id = $1;

-- name: ListConstellationsByUserID :many
SELECT id, user_id, name, constellation_insight, star_ids, topic_prompts, created_at, updated_at
FROM constellations WHERE user_id = $1 ORDER BY updated_at DESC;

-- name: GetConstellationByID :one
SELECT id, user_id, name, constellation_insight, star_ids, topic_prompts, created_at, updated_at
FROM constellations WHERE id = $1;

-- name: GetConstellationByStarID :one
SELECT id, user_id, name, constellation_insight, star_ids, topic_prompts, created_at, updated_at
FROM constellations WHERE star_ids @> ARRAY[$1]::UUID[];
