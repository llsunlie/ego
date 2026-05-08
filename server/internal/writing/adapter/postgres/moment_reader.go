package postgres

import (
	"context"

	conversationdomain "ego-server/internal/conversation/domain"
	"ego-server/internal/platform/postgres/sqlc"
	"ego-server/internal/writing/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type ChatMomentReader struct {
	queries *sqlc.Queries
}

func NewChatMomentReader(queries *sqlc.Queries) *ChatMomentReader {
	return &ChatMomentReader{queries: queries}
}

var _ conversationdomain.MomentReader = (*ChatMomentReader)(nil)

func (r *ChatMomentReader) FindByIDs(ctx context.Context, ids []string) ([]domain.Moment, error) {
	pgIDs := make([]pgtype.UUID, len(ids))
	for i, id := range ids {
		uid, err := uuid.Parse(id)
		if err != nil {
			return nil, err
		}
		pgIDs[i] = pgtype.UUID{Bytes: [16]byte(uid), Valid: true}
	}

	rows, err := r.queries.ListMomentsByIDs(ctx, pgIDs)
	if err != nil {
		return nil, err
	}

	moments := make([]domain.Moment, len(rows))
	for i, row := range rows {
		moments[i] = *toDomainMoment(row)
	}
	return moments, nil
}
