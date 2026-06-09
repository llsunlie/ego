-- name: InsertFeedback :exec
INSERT INTO feedbacks (id, user_id, content, created_at) VALUES ($1, $2, $3, $4);
