package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"ego-server/internal/platform/logging"
	"ego-server/internal/platform/postgres/sqlc"
	"ego-server/internal/writing/domain"
)

const searchMomentIDsSQL = `
SELECT id, similarity(content, $1) AS score
FROM moments
WHERE user_id = $2
  AND id != $3
  AND trace_id != $4
  AND content % $1
ORDER BY similarity(content, $1) DESC
LIMIT $5
`

// MomentSparseSearch implements both MomentSearchIndexer and EchoSparseCandidateReader
// using PostgreSQL pg_trgm trigram similarity instead of Elasticsearch.
type MomentSparseSearch struct {
	db sqlc.DBTX
}

func NewMomentSparseSearch(_ *sqlc.Queries, db sqlc.DBTX) *MomentSparseSearch {
	return &MomentSparseSearch{db: db}
}

// IndexMoment is a no-op: the content is already stored in the moments table,
// and the GIN trigram index automatically covers it.
func (s *MomentSparseSearch) IndexMoment(_ context.Context, _ domain.Moment) error {
	return nil
}

// SearchMomentIDs returns moment IDs ranked by pg_trgm similarity to the
// current moment's content. Excludes the current moment and other moments
// from the same trace.
func (s *MomentSparseSearch) SearchMomentIDs(ctx context.Context, current domain.Moment, limit int32) ([]string, error) {
	logger := logging.FromContext(ctx)
	if s == nil || s.db == nil || limit <= 0 {
		return nil, nil
	}

	uid, err := uuid.Parse(current.UserID)
	if err != nil {
		return nil, fmt.Errorf("parse user_id: %w", err)
	}
	mid, err := uuid.Parse(current.ID)
	if err != nil {
		return nil, fmt.Errorf("parse moment_id: %w", err)
	}
	tid, err := uuid.Parse(current.TraceID)
	if err != nil {
		return nil, fmt.Errorf("parse trace_id: %w", err)
	}

	start := time.Now()
	rows, err := s.db.Query(ctx, searchMomentIDsSQL,
		current.Content,
		pgtype.UUID{Bytes: [16]byte(uid), Valid: true},
		pgtype.UUID{Bytes: [16]byte(mid), Valid: true},
		pgtype.UUID{Bytes: [16]byte(tid), Valid: true},
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("sparse search moments: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var (
			id    pgtype.UUID
			score float64
		)
		if err := rows.Scan(&id, &score); err != nil {
			return nil, fmt.Errorf("scan sparse search row: %w", err)
		}
		parsed, err := uuid.FromBytes(id.Bytes[:])
		if err != nil {
			continue
		}
		ids = append(ids, parsed.String())
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("sparse search rows: %w", err)
	}

	logger.DebugContext(ctx, "MomentSparseSearch: pg_trgm candidates loaded",
		"user_id", current.UserID,
		"moment_id", current.ID,
		"candidate_count", len(ids),
		"limit", limit,
		"elapsed_ms", time.Since(start).Milliseconds(),
	)
	return ids, nil
}

var _ domain.MomentSearchIndexer = (*MomentSparseSearch)(nil)
var _ domain.EchoSparseCandidateReader = (*MomentSparseSearch)(nil)
