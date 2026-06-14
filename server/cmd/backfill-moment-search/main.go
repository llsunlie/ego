package main

import (
	"context"
	"fmt"
	"os"

	"ego-server/internal/config"
	platformes "ego-server/internal/platform/elasticsearch"
	"ego-server/internal/platform/logging"
	"ego-server/internal/platform/postgres"
	"ego-server/internal/platform/postgres/sqlc"
	writinges "ego-server/internal/writing/adapter/elasticsearch"
	writingdomain "ego-server/internal/writing/domain"

	"github.com/google/uuid"
)

const listMomentsForSearchBackfillSQL = `
SELECT id, trace_id, user_id, content, embeddings, created_at
FROM moments
ORDER BY created_at ASC
`

const batchSize = 500

func main() {
	logger := logging.NewDefault()
	cfg := config.Load()

	pool, err := postgres.Connect(cfg.DatabaseURL)
	if err != nil {
		logger.Error("db connect failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	search := writinges.NewMomentSearch(platformes.NewClient(platformes.Config{
		URL:      cfg.ElasticsearchURL,
		Username: cfg.ElasticsearchUser,
		Password: cfg.ElasticsearchPass,
	}, logger), writinges.DefaultMomentIndex)

	ctx := context.Background()
	if err := search.EnsureIndex(ctx); err != nil {
		logger.Error("ensure search index failed", "error", err)
		os.Exit(1)
	}

	rows, err := pool.Query(ctx, listMomentsForSearchBackfillSQL)
	if err != nil {
		logger.Error("list moments failed", "error", err)
		os.Exit(1)
	}
	defer rows.Close()

	var scanned, indexed, failed int
	batch := make([]writingdomain.Moment, 0, batchSize)
	flush := func() {
		if len(batch) == 0 {
			return
		}
		if err := search.BulkIndexMoments(ctx, batch); err != nil {
			failed += len(batch)
			logger.Error("bulk index moments failed", "count", len(batch), "error", err)
		} else {
			indexed += len(batch)
			logger.Debug("bulk indexed moments", "count", len(batch))
		}
		batch = batch[:0]
	}

	for rows.Next() {
		scanned++
		var row sqlc.Moment
		if err := rows.Scan(&row.ID, &row.TraceID, &row.UserID, &row.Content, &row.Embeddings, &row.CreatedAt); err != nil {
			failed++
			logger.Error("scan moment failed", "error", err)
			continue
		}
		moment, err := backfillMomentFromRow(row)
		if err != nil {
			failed++
			logger.Error("parse moment failed", "error", err)
			continue
		}
		batch = append(batch, moment)
		if len(batch) >= batchSize {
			flush()
		}
	}
	flush()
	if err := rows.Err(); err != nil {
		logger.Error("iterate moments failed", "error", err)
		os.Exit(1)
	}

	logger.Info("moment search backfill completed",
		"scanned", scanned,
		"indexed", indexed,
		"failed", failed,
		"index", writinges.DefaultMomentIndex,
	)
	if failed > 0 {
		os.Exit(1)
	}
}

func backfillMomentFromRow(row sqlc.Moment) (writingdomain.Moment, error) {
	id, err := uuid.FromBytes(row.ID.Bytes[:])
	if err != nil {
		return writingdomain.Moment{}, err
	}
	traceID, err := uuid.FromBytes(row.TraceID.Bytes[:])
	if err != nil {
		return writingdomain.Moment{}, err
	}
	userID, err := uuid.FromBytes(row.UserID.Bytes[:])
	if err != nil {
		return writingdomain.Moment{}, err
	}
	if !row.CreatedAt.Valid {
		return writingdomain.Moment{}, fmt.Errorf("created_at is invalid for moment %s", id.String())
	}
	return writingdomain.Moment{
		ID:        id.String(),
		TraceID:   traceID.String(),
		UserID:    userID.String(),
		Content:   row.Content,
		CreatedAt: row.CreatedAt.Time,
	}, nil
}
