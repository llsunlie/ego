package ai

import (
	"context"
	"fmt"

	platformai "ego-server/internal/platform/ai"
	"ego-server/internal/platform/logging"
	"ego-server/internal/writing/domain"
)

// Embedder implements domain.EmbeddingGenerator by delegating to
// platform/ai.Client.CreateEmbedding. It maps the platform-level
// EmbeddingResult into the writing-domain EmbeddingEntry slice.
type Embedder struct {
	client *platformai.Client
}

func NewEmbedder(client *platformai.Client) *Embedder {
	return &Embedder{client: client}
}

func (e *Embedder) Generate(ctx context.Context, content string) ([]domain.EmbeddingEntry, error) {
	logger := logging.FromContext(ctx)
	logger.DebugContext(ctx, "generating embedding", "input_len", len([]rune(content)))

	result, err := e.client.CreateEmbedding(ctx, content)
	if err != nil {
		logger.ErrorContext(ctx, "embedding generation failed", "error", err)
		return nil, fmt.Errorf("ai embedder: %w", err)
	}

	logger.InfoContext(ctx, "embedding generated",
		"model", result.Model,
		"dim", len(result.Embedding),
		"tokens", result.Usage.TotalTokens,
	)
	return []domain.EmbeddingEntry{
		{Model: result.Model, Embedding: result.Embedding},
	}, nil
}
