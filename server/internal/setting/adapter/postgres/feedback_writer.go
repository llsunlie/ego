package postgres

import (
	"context"

	"ego-server/internal/platform/postgres/sqlc"
	"ego-server/internal/setting/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// FeedbackWriter implements domain.FeedbackWriter using sqlc.
type FeedbackWriter struct {
	queries *sqlc.Queries
}

func NewFeedbackWriter(queries *sqlc.Queries) *FeedbackWriter {
	return &FeedbackWriter{queries: queries}
}

func (w *FeedbackWriter) Save(ctx context.Context, fb *domain.Feedback) error {
	id, err := uuid.Parse(fb.ID)
	if err != nil {
		return err
	}
	userID, err := uuid.Parse(fb.UserID)
	if err != nil {
		return err
	}

	var idArr, userIDArr [16]byte
	copy(idArr[:], id[:])
	copy(userIDArr[:], userID[:])

	return w.queries.InsertFeedback(ctx, sqlc.InsertFeedbackParams{
		ID:        pgtype.UUID{Bytes: idArr, Valid: true},
		UserID:    pgtype.UUID{Bytes: userIDArr, Valid: true},
		Content:   fb.Content,
		CreatedAt: pgtype.Timestamptz{Time: fb.CreatedAt, Valid: true},
	})
}
