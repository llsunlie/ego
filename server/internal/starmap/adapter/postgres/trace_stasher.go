package postgres

import (
	"context"

	"ego-server/internal/platform/postgres/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type TraceStasher struct {
	queries *sqlc.Queries
}

func NewTraceStasher(queries *sqlc.Queries) *TraceStasher {
	return &TraceStasher{queries: queries}
}

func (s *TraceStasher) MarkStashed(ctx context.Context, traceID string) error {
	uid, err := uuid.Parse(traceID)
	if err != nil {
		return err
	}

	return s.queries.UpdateTrace(ctx, sqlc.UpdateTraceParams{
		ID:      pgtype.UUID{Bytes: [16]byte(uid), Valid: true},
		Stashed: true,
	})
}
