package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"ego-server/internal/platform/postgres/sqlc"
	"ego-server/internal/writing/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
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

	embeddingsJSON, err := json.Marshal(moment.Embeddings)
	if err != nil {
		return err
	}

	return r.queries.CreateMoment(ctx, sqlc.CreateMomentParams{
		ID:         pgtype.UUID{Bytes: [16]byte(uid), Valid: true},
		TraceID:    pgtype.UUID{Bytes: [16]byte(traceID), Valid: true},
		UserID:     pgtype.UUID{Bytes: [16]byte(userID), Valid: true},
		Content:    moment.Content,
		Embeddings: embeddingsJSON,
		CreatedAt:  pgtype.Timestamptz{Time: now, Valid: true},
	})
}

func (r *MomentRepository) GetByIDs(ctx context.Context, ids []string) ([]domain.Moment, error) {
	if len(ids) == 0 {
		return nil, nil
	}

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

	var embeddings []domain.EmbeddingEntry
	if len(row.Embeddings) > 0 {
		json.Unmarshal(row.Embeddings, &embeddings)
	}
	if embeddings == nil {
		embeddings = []domain.EmbeddingEntry{}
	}

	return &domain.Moment{
		ID:         id.String(),
		TraceID:    traceID.String(),
		UserID:     userID.String(),
		Content:    row.Content,
		Embeddings: embeddings,
		CreatedAt:  row.CreatedAt.Time,
	}
}
