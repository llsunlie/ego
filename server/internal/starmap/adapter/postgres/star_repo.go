package postgres

import (
	"context"
	"errors"

	"ego-server/internal/platform/postgres/sqlc"
	"ego-server/internal/starmap/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type StarRepository struct {
	queries *sqlc.Queries
}

func NewStarRepository(queries *sqlc.Queries) *StarRepository {
	return &StarRepository{queries: queries}
}

func (r *StarRepository) Create(ctx context.Context, star *domain.Star) error {
	uid, err := uuid.Parse(star.ID)
	if err != nil {
		return err
	}
	userID, err := uuid.Parse(star.UserID)
	if err != nil {
		return err
	}
	traceID, err := uuid.Parse(star.TraceID)
	if err != nil {
		return err
	}

	return r.queries.CreateStar(ctx, sqlc.CreateStarParams{
		ID:        pgtype.UUID{Bytes: [16]byte(uid), Valid: true},
		UserID:    pgtype.UUID{Bytes: [16]byte(userID), Valid: true},
		TraceID:   pgtype.UUID{Bytes: [16]byte(traceID), Valid: true},
		Topic:     star.Topic,
		CreatedAt: pgtype.Timestamptz{Time: star.CreatedAt, Valid: true},
	})
}

func (r *StarRepository) FindByTraceID(ctx context.Context, traceID string) (*domain.Star, error) {
	uid, err := uuid.Parse(traceID)
	if err != nil {
		return nil, err
	}

	row, err := r.queries.GetStarByTraceID(ctx, pgtype.UUID{Bytes: [16]byte(uid), Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrStarNotFound
		}
		return nil, err
	}

	return toDomainStar(row), nil
}

func (r *StarRepository) FindByIDs(ctx context.Context, ids []string) ([]domain.Star, error) {
	pgIDs := make([]pgtype.UUID, len(ids))
	for i, id := range ids {
		uid, err := uuid.Parse(id)
		if err != nil {
			return nil, err
		}
		pgIDs[i] = pgtype.UUID{Bytes: [16]byte(uid), Valid: true}
	}

	rows, err := r.queries.ListStarsByIDs(ctx, pgIDs)
	if err != nil {
		return nil, err
	}

	result := make([]domain.Star, len(rows))
	for i, row := range rows {
		s := toDomainStar(row)
		result[i] = *s
	}
	return result, nil
}

func toDomainStar(row sqlc.Star) *domain.Star {
	id, _ := uuid.FromBytes(row.ID.Bytes[:])
	userID, _ := uuid.FromBytes(row.UserID.Bytes[:])
	traceID, _ := uuid.FromBytes(row.TraceID.Bytes[:])

	return &domain.Star{
		ID:        id.String(),
		UserID:    userID.String(),
		TraceID:   traceID.String(),
		Topic:     row.Topic,
		CreatedAt: row.CreatedAt.Time,
	}
}
