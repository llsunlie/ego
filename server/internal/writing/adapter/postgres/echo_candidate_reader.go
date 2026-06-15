package postgres

import (
	"context"
	"fmt"
	"time"

	"ego-server/internal/platform/logging"
	"ego-server/internal/platform/postgres/sqlc"
	"ego-server/internal/writing/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

const findNearestMomentsSQL = `
SELECT
  m.id,
  m.trace_id,
  m.user_id,
  m.content,
  m.embeddings,
  m.created_at
FROM moment_embedding_vectors mev
JOIN moments m ON m.id = mev.moment_id
WHERE mev.user_id = $1
  AND mev.model = $2
  AND mev.moment_id <> $3
ORDER BY mev.embedding <=> $4::vector
LIMIT $5
`

type EchoCandidateReader struct {
	db           sqlc.DBTX
	embeddingDim int
}

func NewEchoCandidateReader(_ *sqlc.Queries, db sqlc.DBTX, embeddingDim int) *EchoCandidateReader {
	return &EchoCandidateReader{db: db, embeddingDim: embeddingDim}
}

func (r *EchoCandidateReader) FindNearestMoments(ctx context.Context, userID string, currentMomentID string, model string, embedding []float32, limit int32) ([]domain.Moment, error) {
	logger := logging.FromContext(ctx)
	if limit <= 0 {
		logger.DebugContext(ctx, "EchoCandidateReader: skip nearest query because limit is non-positive",
			"user_id", userID,
			"moment_id", currentMomentID,
			"limit", limit,
		)
		return nil, nil
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		return nil, err
	}
	mid, err := uuid.Parse(currentMomentID)
	if err != nil {
		return nil, err
	}
	literal, err := vectorLiteral(embedding, r.embeddingDim)
	if err != nil {
		return nil, err
	}

	start := time.Now()
	rows, err := r.db.Query(ctx, findNearestMomentsSQL,
		pgtype.UUID{Bytes: [16]byte(uid), Valid: true},
		model,
		pgtype.UUID{Bytes: [16]byte(mid), Valid: true},
		literal,
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("find nearest moments: %w", err)
	}
	defer rows.Close()

	var moments []domain.Moment
	for rows.Next() {
		var row sqlc.Moment
		if err := rows.Scan(&row.ID, &row.TraceID, &row.UserID, &row.Content, &row.Embeddings, &row.CreatedAt); err != nil {
			return nil, err
		}
		moments = append(moments, *toDomainMoment(row))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	logger.DebugContext(ctx, "EchoCandidateReader: nearest moments loaded",
		"user_id", userID,
		"moment_id", currentMomentID,
		"model", model,
		"candidate_count", len(moments),
		"top_k", limit,
		"elapsed_ms", time.Since(start).Milliseconds(),
	)
	return moments, nil
}

var _ domain.EchoCandidateReader = (*EchoCandidateReader)(nil)
