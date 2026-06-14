package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"ego-server/internal/config"
	"ego-server/internal/platform/logging"
	"ego-server/internal/platform/postgres"
	"ego-server/internal/platform/postgres/sqlc"
	writingpostgres "ego-server/internal/writing/adapter/postgres"
	writingdomain "ego-server/internal/writing/domain"

	"github.com/google/uuid"
)

const listMomentsForVectorBackfillSQL = `
SELECT id, trace_id, user_id, content, embeddings, created_at
FROM moments
ORDER BY created_at ASC
`

func main() {
	logger := logging.NewDefault()
	cfg := config.Load()
	if cfg.AIEmbeddingModel == "" {
		logger.Error("AI_EMBEDDING_MODEL is required")
		os.Exit(1)
	}
	embeddingDim, err := strconv.Atoi(cfg.AIEmbeddingDim)
	if err != nil {
		logger.Error("invalid AI_EMBEDDING_DIM", "error", err)
		os.Exit(1)
	}
	logger.Debug("moment vector backfill starting",
		"model", cfg.AIEmbeddingModel,
		"dim", embeddingDim,
	)

	pool, err := postgres.Connect(cfg.DatabaseURL)
	if err != nil {
		logger.Error("db connect failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	ctx := context.Background()
	queries := sqlc.New(pool)
	repo := writingpostgres.NewMomentRepositoryWithVector(queries, pool, embeddingDim)

	rows, err := pool.Query(ctx, listMomentsForVectorBackfillSQL)
	if err != nil {
		logger.Error("list moments failed", "error", err)
		os.Exit(1)
	}
	defer rows.Close()

	var scanned, upserted, skipped, failed int
	for rows.Next() {
		scanned++
		var row sqlc.Moment
		if err := rows.Scan(&row.ID, &row.TraceID, &row.UserID, &row.Content, &row.Embeddings, &row.CreatedAt); err != nil {
			failed++
			logger.Error("scan moment failed", "error", err)
			continue
		}

		moment, entries, err := backfillMomentFromRow(row)
		if err != nil {
			failed++
			logger.Error("parse moment failed", "error", err)
			continue
		}
		entry, ok := selectEmbedding(entries, cfg.AIEmbeddingModel)
		if !ok {
			skipped++
			logger.Warn("embedding model not found, skipping", "moment_id", moment.ID, "model", cfg.AIEmbeddingModel)
			continue
		}
		if len(entry.Embedding) != embeddingDim {
			skipped++
			logger.Warn("embedding dimension mismatch, skipping", "moment_id", moment.ID, "got", len(entry.Embedding), "want", embeddingDim)
			continue
		}

		logger.Debug("upserting moment vector",
			"moment_id", moment.ID,
			"trace_id", moment.TraceID,
			"user_id", moment.UserID,
			"model", entry.Model,
			"dim", len(entry.Embedding),
		)
		if err := repo.UpsertEmbeddingVector(ctx, moment, entry); err != nil {
			failed++
			logger.Error("upsert vector failed", "moment_id", moment.ID, "error", err)
			continue
		}
		upserted++
	}
	if err := rows.Err(); err != nil {
		logger.Error("iterate moments failed", "error", err)
		os.Exit(1)
	}

	logger.Info("moment vector backfill completed",
		"scanned", scanned,
		"upserted", upserted,
		"skipped", skipped,
		"failed", failed,
		"model", cfg.AIEmbeddingModel,
		"dim", embeddingDim,
	)

	if failed > 0 {
		os.Exit(1)
	}
}

func backfillMomentFromRow(row sqlc.Moment) (writingdomain.Moment, []writingdomain.EmbeddingEntry, error) {
	id, err := uuid.FromBytes(row.ID.Bytes[:])
	if err != nil {
		return writingdomain.Moment{}, nil, err
	}
	traceID, err := uuid.FromBytes(row.TraceID.Bytes[:])
	if err != nil {
		return writingdomain.Moment{}, nil, err
	}
	userID, err := uuid.FromBytes(row.UserID.Bytes[:])
	if err != nil {
		return writingdomain.Moment{}, nil, err
	}
	if !row.CreatedAt.Valid {
		return writingdomain.Moment{}, nil, fmt.Errorf("created_at is invalid for moment %s", id.String())
	}

	var entries []writingdomain.EmbeddingEntry
	if len(row.Embeddings) > 0 {
		if err := json.Unmarshal(row.Embeddings, &entries); err != nil {
			return writingdomain.Moment{}, nil, err
		}
	}

	return writingdomain.Moment{
		ID:         id.String(),
		TraceID:    traceID.String(),
		UserID:     userID.String(),
		Content:    row.Content,
		Embeddings: entries,
		CreatedAt:  row.CreatedAt.Time,
	}, entries, nil
}

func selectEmbedding(entries []writingdomain.EmbeddingEntry, model string) (writingdomain.EmbeddingEntry, bool) {
	for _, entry := range entries {
		if entry.Model == model {
			return entry, true
		}
	}
	return writingdomain.EmbeddingEntry{}, false
}
