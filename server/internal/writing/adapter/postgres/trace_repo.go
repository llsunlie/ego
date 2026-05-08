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
)

type TraceRepository struct {
	queries *sqlc.Queries
}

func NewTraceRepository(queries *sqlc.Queries) *TraceRepository {
	return &TraceRepository{queries: queries}
}

func (r *TraceRepository) Create(ctx context.Context, trace *domain.Trace) error {
	uid, err := uuid.Parse(trace.ID)
	if err != nil {
		return err
	}
	userID, err := uuid.Parse(trace.UserID)
	if err != nil {
		return err
	}

	now := time.Now()
	trace.CreatedAt = now
	trace.UpdatedAt = now

	var topic pgtype.Text
	if trace.Topic != "" {
		topic = pgtype.Text{String: trace.Topic, Valid: true}
	}

	return r.queries.CreateTrace(ctx, sqlc.CreateTraceParams{
		ID:        pgtype.UUID{Bytes: [16]byte(uid), Valid: true},
		UserID:    pgtype.UUID{Bytes: [16]byte(userID), Valid: true},
		Topic:     topic,
		CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
	})
}

func (r *TraceRepository) GetByID(ctx context.Context, id string) (*domain.Trace, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}

	row, err := r.queries.GetTraceByID(ctx, pgtype.UUID{Bytes: [16]byte(uid), Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTraceNotFound
		}
		return nil, err
	}

	return toDomainTrace(row), nil
}

func (r *TraceRepository) Delete(ctx context.Context, id string) error {
	uid, err := uuid.Parse(id)
	if err != nil {
		return err
	}
	return r.queries.DeleteTrace(ctx, pgtype.UUID{Bytes: [16]byte(uid), Valid: true})
}

func (r *TraceRepository) Update(ctx context.Context, trace *domain.Trace) error {
	uid, err := uuid.Parse(trace.ID)
	if err != nil {
		return err
	}

	now := time.Now()
	trace.UpdatedAt = now

	var topic pgtype.Text
	if trace.Topic != "" {
		topic = pgtype.Text{String: trace.Topic, Valid: true}
	}

	return r.queries.UpdateTrace(ctx, sqlc.UpdateTraceParams{
		ID:        pgtype.UUID{Bytes: [16]byte(uid), Valid: true},
		Topic:     topic,
		UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
	})
}

func toDomainTrace(row sqlc.Trace) *domain.Trace {
	id, _ := uuid.FromBytes(row.ID.Bytes[:])
	userID, _ := uuid.FromBytes(row.UserID.Bytes[:])

	topic := ""
	if row.Topic.Valid {
		topic = row.Topic.String
	}

	return &domain.Trace{
		ID:        id.String(),
		UserID:    userID.String(),
		Topic:     topic,
		CreatedAt: row.CreatedAt.Time,
		UpdatedAt: row.UpdatedAt.Time,
	}
}
