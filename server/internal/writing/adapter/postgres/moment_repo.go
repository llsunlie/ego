package postgres

import (
	"context"
	"errors"
	"time"

	"ego-server/internal/platform/postgres/sqlc"
	"ego-server/internal/writing/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pgvector/pgvector-go"
)

type MomentRepository struct {
	queries *sqlc.Queries
}

func NewMomentRepository(queries *sqlc.Queries) *MomentRepository {
	return &MomentRepository{queries: queries}
}

func (r *MomentRepository) Create(ctx context.Context, moment *domain.Moment) error {
	uid, err := uuid.Parse(moment.ID)
	if err != nil {
		return err
	}
	traceID, err := uuid.Parse(moment.TraceID)
	if err != nil {
		return err
	}
	userID, err := uuid.Parse(moment.UserID)
	if err != nil {
		return err
	}

	now := time.Now()
	moment.CreatedAt = now

	embedding := pgvector.NewVector(make([]float32, 0))
	if len(moment.Embedding) > 0 {
		embedding = pgvector.NewVector(moment.Embedding)
	}

	return r.queries.CreateMoment(ctx, sqlc.CreateMomentParams{
		ID:        pgtype.UUID{Bytes: [16]byte(uid), Valid: true},
		TraceID:   pgtype.UUID{Bytes: [16]byte(traceID), Valid: true},
		UserID:    pgtype.UUID{Bytes: [16]byte(userID), Valid: true},
		Content:   moment.Content,
		Embedding: embedding,
		Connected: false,
		CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
	})
}

func (r *MomentRepository) GetByID(ctx context.Context, id string) (*domain.Moment, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	row, err := r.queries.GetMomentByID(ctx, pgtype.UUID{Bytes: [16]byte(uid), Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrMomentNotFound
		}
		return nil, err
	}

	return toDomainMoment(row), nil
}

func (r *MomentRepository) ListByTraceID(ctx context.Context, traceID string) ([]domain.Moment, error) {
	tid, err := uuid.Parse(traceID)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.ListMomentsByTraceID(ctx, pgtype.UUID{Bytes: [16]byte(tid), Valid: true})
	if err != nil {
		return nil, err
	}

	moments := make([]domain.Moment, len(rows))
	for i, row := range rows {
		moments[i] = *toDomainMoment(row)
	}
	return moments, nil
}

func (r *MomentRepository) ListByUserID(ctx context.Context, userID string) ([]domain.Moment, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}

	rows, err := r.queries.ListMomentsByUserID(ctx, pgtype.UUID{Bytes: [16]byte(uid), Valid: true})
	if err != nil {
		return nil, err
	}

	moments := make([]domain.Moment, len(rows))
	for i, row := range rows {
		moments[i] = *toDomainMoment(row)
	}
	return moments, nil
}

func toDomainMoment(row sqlc.Moment) *domain.Moment {
	id, _ := uuid.FromBytes(row.ID.Bytes[:])
	traceID, _ := uuid.FromBytes(row.TraceID.Bytes[:])
	userID, _ := uuid.FromBytes(row.UserID.Bytes[:])

	var embedding []float32
	if vec := row.Embedding.Slice(); vec != nil {
		embedding = vec
	}

	return &domain.Moment{
		ID:        id.String(),
		TraceID:   traceID.String(),
		UserID:    userID.String(),
		Content:   row.Content,
		Embedding: embedding,
		Connected: row.Connected,
		CreatedAt: row.CreatedAt.Time,
	}
}
