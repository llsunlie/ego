package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"ego-server/internal/platform/postgres/sqlc"
	"ego-server/internal/starmap/domain"
)

type TraceProfileRepository struct {
	db           sqlc.DBTX
	embeddingDim int
}

func NewTraceProfileRepository(db sqlc.DBTX, embeddingDim int) *TraceProfileRepository {
	return &TraceProfileRepository{db: db, embeddingDim: embeddingDim}
}

const upsertTraceProfileSQL = `
INSERT INTO trace_profiles (
  trace_id, user_id, topic, summary, keywords, emotions, scenes,
  central_pattern, pattern_tags, representative_moment_id, profile_text,
  status, retry_count, last_error, created_at, updated_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7,
  $8, $9, $10, $11,
  $12, $13, $14, $15, $16
)
ON CONFLICT (trace_id) DO UPDATE SET
  user_id = EXCLUDED.user_id,
  topic = EXCLUDED.topic,
  summary = EXCLUDED.summary,
  keywords = EXCLUDED.keywords,
  emotions = EXCLUDED.emotions,
  scenes = EXCLUDED.scenes,
  central_pattern = EXCLUDED.central_pattern,
  pattern_tags = EXCLUDED.pattern_tags,
  representative_moment_id = EXCLUDED.representative_moment_id,
  profile_text = EXCLUDED.profile_text,
  status = EXCLUDED.status,
  retry_count = EXCLUDED.retry_count,
  last_error = EXCLUDED.last_error,
  updated_at = EXCLUDED.updated_at
`

const upsertTraceProfileVectorSQL = `
INSERT INTO trace_profile_vectors (
  trace_id, user_id, model, dim, embedding, created_at, updated_at
) VALUES (
  $1, $2, $3, $4, $5::vector, $6, $7
)
ON CONFLICT (trace_id) DO UPDATE SET
  user_id = EXCLUDED.user_id,
  model = EXCLUDED.model,
  dim = EXCLUDED.dim,
  embedding = EXCLUDED.embedding,
  updated_at = EXCLUDED.updated_at
`

func (r *TraceProfileRepository) Upsert(ctx context.Context, profile *domain.TraceProfile, vector *domain.TraceProfileVector) error {
	if r.db == nil {
		return fmt.Errorf("db is required for trace profile upsert")
	}
	if profile == nil {
		return fmt.Errorf("trace profile is nil")
	}

	traceID, err := uuid.Parse(profile.TraceID)
	if err != nil {
		return fmt.Errorf("parse trace_id: %w", err)
	}
	userID, err := uuid.Parse(profile.UserID)
	if err != nil {
		return fmt.Errorf("parse user_id: %w", err)
	}
	representativeMomentID, err := nullableUUID(profile.RepresentativeMomentID)
	if err != nil {
		return fmt.Errorf("parse representative_moment_id: %w", err)
	}
	keywords, err := json.Marshal(profile.Keywords)
	if err != nil {
		return fmt.Errorf("marshal keywords: %w", err)
	}
	emotions, err := json.Marshal(profile.Emotions)
	if err != nil {
		return fmt.Errorf("marshal emotions: %w", err)
	}
	scenes, err := json.Marshal(profile.Scenes)
	if err != nil {
		return fmt.Errorf("marshal scenes: %w", err)
	}
	patternTags, err := json.Marshal(profile.PatternTags)
	if err != nil {
		return fmt.Errorf("marshal pattern tags: %w", err)
	}

	if _, err := r.db.Exec(ctx, upsertTraceProfileSQL,
		pgtype.UUID{Bytes: [16]byte(traceID), Valid: true},
		pgtype.UUID{Bytes: [16]byte(userID), Valid: true},
		profile.Topic,
		profile.Summary,
		keywords,
		emotions,
		scenes,
		profile.CentralPattern,
		patternTags,
		representativeMomentID,
		profile.ProfileText,
		profile.Status,
		int32(profile.RetryCount),
		profile.LastError,
		pgtype.Timestamptz{Time: profile.CreatedAt, Valid: true},
		pgtype.Timestamptz{Time: profile.UpdatedAt, Valid: true},
	); err != nil {
		return fmt.Errorf("upsert trace profile: %w", err)
	}

	if vector == nil {
		return nil
	}
	literal, err := vectorLiteral(vector.Embedding, r.embeddingDim)
	if err != nil {
		return fmt.Errorf("trace profile vector literal: %w", err)
	}
	if vector.Model == "" {
		return fmt.Errorf("trace profile vector model is empty")
	}

	if _, err := r.db.Exec(ctx, upsertTraceProfileVectorSQL,
		pgtype.UUID{Bytes: [16]byte(traceID), Valid: true},
		pgtype.UUID{Bytes: [16]byte(userID), Valid: true},
		vector.Model,
		int32(vector.Dim),
		literal,
		pgtype.Timestamptz{Time: vector.CreatedAt, Valid: true},
		pgtype.Timestamptz{Time: vector.UpdatedAt, Valid: true},
	); err != nil {
		return fmt.Errorf("upsert trace profile vector: %w", err)
	}

	return nil
}

func nullableUUID(value string) (pgtype.UUID, error) {
	if value == "" {
		return pgtype.UUID{}, nil
	}
	parsed, err := uuid.Parse(value)
	if err != nil {
		return pgtype.UUID{}, err
	}
	return pgtype.UUID{Bytes: [16]byte(parsed), Valid: true}, nil
}

func vectorLiteral(values []float32, expectedDim int) (string, error) {
	if expectedDim > 0 && len(values) != expectedDim {
		return "", fmt.Errorf("embedding dimension mismatch: got %d, want %d", len(values), expectedDim)
	}
	if len(values) == 0 {
		return "", fmt.Errorf("embedding is empty")
	}

	var b strings.Builder
	b.WriteByte('[')
	for i, v := range values {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(strconv.FormatFloat(float64(v), 'g', -1, 32))
	}
	b.WriteByte(']')
	return b.String(), nil
}

var _ domain.TraceProfileRepository = (*TraceProfileRepository)(nil)
