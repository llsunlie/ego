package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"ego-server/internal/platform/logging"
	"ego-server/internal/platform/postgres/sqlc"
	"ego-server/internal/writing/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type MomentRepository struct {
	queries      *sqlc.Queries
	db           sqlc.DBTX
	embeddingDim int
}

func NewMomentRepository(queries *sqlc.Queries) *MomentRepository {
	return &MomentRepository{queries: queries}
}

func NewMomentRepositoryWithVector(queries *sqlc.Queries, db sqlc.DBTX, embeddingDim int) *MomentRepository {
	return &MomentRepository{queries: queries, db: db, embeddingDim: embeddingDim}
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

	if r.db != nil && len(moment.Embeddings) > 0 {
		return r.createWithVector(ctx, moment, uid, traceID, userID, embeddingsJSON)
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

const createMomentWithVectorSQL = `
WITH inserted AS (
  INSERT INTO moments (id, trace_id, user_id, content, embeddings, created_at)
  VALUES ($1, $2, $3, $4, $5, $6)
  RETURNING id, trace_id, user_id, created_at
)
INSERT INTO moment_embedding_vectors (moment_id, user_id, trace_id, model, dim, embedding, created_at)
SELECT id, user_id, trace_id, $7, $8, $9::vector, created_at
FROM inserted
`

const upsertMomentEmbeddingVectorSQL = `
INSERT INTO moment_embedding_vectors (moment_id, user_id, trace_id, model, dim, embedding, created_at)
VALUES ($1, $2, $3, $4, $5, $6::vector, $7)
ON CONFLICT (moment_id, model) DO UPDATE SET
  user_id = EXCLUDED.user_id,
  trace_id = EXCLUDED.trace_id,
  dim = EXCLUDED.dim,
  embedding = EXCLUDED.embedding,
  created_at = EXCLUDED.created_at
`

func (r *MomentRepository) createWithVector(ctx context.Context, moment *domain.Moment, uid uuid.UUID, traceID uuid.UUID, userID uuid.UUID, embeddingsJSON []byte) error {
	logger := logging.FromContext(ctx)
	entry := moment.Embeddings[0]
	literal, err := vectorLiteral(entry.Embedding, r.embeddingDim)
	if err != nil {
		return err
	}
	if entry.Model == "" {
		return fmt.Errorf("embedding model is empty")
	}
	logger.DebugContext(ctx, "MomentRepository: creating moment with vector",
		"moment_id", moment.ID,
		"trace_id", moment.TraceID,
		"user_id", moment.UserID,
		"model", entry.Model,
		"dim", len(entry.Embedding),
	)
	if _, err := r.db.Exec(ctx, createMomentWithVectorSQL,
		pgtype.UUID{Bytes: [16]byte(uid), Valid: true},
		pgtype.UUID{Bytes: [16]byte(traceID), Valid: true},
		pgtype.UUID{Bytes: [16]byte(userID), Valid: true},
		moment.Content,
		embeddingsJSON,
		pgtype.Timestamptz{Time: moment.CreatedAt, Valid: true},
		entry.Model,
		len(entry.Embedding),
		literal,
	); err != nil {
		return fmt.Errorf("create moment with vector: %w", err)
	}
	logger.DebugContext(ctx, "MomentRepository: moment vector created",
		"moment_id", moment.ID,
		"model", entry.Model,
		"dim", len(entry.Embedding),
	)
	return nil
}

func (r *MomentRepository) UpsertEmbeddingVector(ctx context.Context, moment domain.Moment, entry domain.EmbeddingEntry) error {
	logger := logging.FromContext(ctx)
	if r.db == nil {
		return fmt.Errorf("db is required for vector upsert")
	}
	mid, err := uuid.Parse(moment.ID)
	if err != nil {
		return err
	}
	tid, err := uuid.Parse(moment.TraceID)
	if err != nil {
		return err
	}
	uid, err := uuid.Parse(moment.UserID)
	if err != nil {
		return err
	}
	literal, err := vectorLiteral(entry.Embedding, r.embeddingDim)
	if err != nil {
		return err
	}
	if entry.Model == "" {
		return fmt.Errorf("embedding model is empty")
	}

	logger.DebugContext(ctx, "MomentRepository: upserting moment vector",
		"moment_id", moment.ID,
		"trace_id", moment.TraceID,
		"user_id", moment.UserID,
		"model", entry.Model,
		"dim", len(entry.Embedding),
	)
	_, err = r.db.Exec(ctx, upsertMomentEmbeddingVectorSQL,
		pgtype.UUID{Bytes: [16]byte(mid), Valid: true},
		pgtype.UUID{Bytes: [16]byte(uid), Valid: true},
		pgtype.UUID{Bytes: [16]byte(tid), Valid: true},
		entry.Model,
		len(entry.Embedding),
		literal,
		pgtype.Timestamptz{Time: moment.CreatedAt, Valid: true},
	)
	if err != nil {
		return fmt.Errorf("upsert moment embedding vector: %w", err)
	}
	logger.DebugContext(ctx, "MomentRepository: moment vector upserted",
		"moment_id", moment.ID,
		"model", entry.Model,
		"dim", len(entry.Embedding),
	)
	return nil
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
