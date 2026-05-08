package postgres

import (
	"context"
	"errors"

	conversationdomain "ego-server/internal/conversation/domain"
	"ego-server/internal/platform/postgres/sqlc"
	starmapdomain "ego-server/internal/starmap/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type StarReader struct {
	queries *sqlc.Queries
}

func NewStarReader(queries *sqlc.Queries) *StarReader {
	return &StarReader{queries: queries}
}

var _ conversationdomain.StarReader = (*StarReader)(nil)

func (r *StarReader) FindByID(ctx context.Context, id string) (*starmapdomain.Star, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	row, err := r.queries.GetStarByID(ctx, pgtype.UUID{Bytes: [16]byte(uid), Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, conversationdomain.ErrStarNotFound
		}
		return nil, err
	}

	return toDomainStar(row), nil
}
