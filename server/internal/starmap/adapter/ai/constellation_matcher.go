package ai

import (
	"context"
	"sync"

	platformai "ego-server/internal/platform/ai"
	"ego-server/internal/platform/logging"
	"ego-server/internal/starmap/domain"
)

const matchThreshold = 0.65

// ConstellationMatcher implements domain.ConstellationMatcher with
// embedding-based semantic matching. Topic and constellation embeddings are
// compared with cosine similarity. Constellation embeddings are compared in
// parallel with goroutines so the latency is dominated by the slowest single
// call rather than the sum.
//
// Cached topic embeddings (Constellation.TopicEmbedding) are preferred to avoid
// redundant API calls; if the cache is empty, the topic is embedded on the fly.
type ConstellationMatcher struct {
	client *platformai.Client
}

func NewConstellationMatcher(client *platformai.Client) *ConstellationMatcher {
	return &ConstellationMatcher{client: client}
}

func (m *ConstellationMatcher) FindMatch(ctx context.Context, topic string, existing []domain.Constellation) (string, error) {
	logger := logging.FromContext(ctx)
	logger.DebugContext(ctx, "ConstellationMatcher: start", "topic", topic, "constellation_count", len(existing))

	if len(existing) == 0 {
		return "", nil
	}

	topicEmb, err := m.client.CreateEmbedding(ctx, topic)
	if err != nil {
		logger.ErrorContext(ctx, "ConstellationMatcher: topic embedding failed", "error", err)
		return "", nil
	}

	type matchResult struct {
		id    string
		score float64
	}

	results := make([]matchResult, len(existing))
	var wg sync.WaitGroup

	for i, c := range existing {
		wg.Add(1)
		go func(idx int, constellation domain.Constellation) {
			defer wg.Done()

			emb, err := m.constellationEmbedding(ctx, constellation)
			if err != nil {
				logger.WarnContext(ctx, "ConstellationMatcher: constellation embedding failed",
					"constellation_id", constellation.ID,
					"error", err,
				)
				return
			}

			score := platformai.CosineSimilarity(topicEmb.Embedding, emb)
			results[idx] = matchResult{id: constellation.ID, score: score}
		}(i, c)
	}

	wg.Wait()

	var bestID string
	var bestScore float64
	for _, r := range results {
		if r.score > bestScore {
			bestScore = r.score
			bestID = r.id
		}
	}

	if bestScore < matchThreshold {
		logger.DebugContext(ctx, "ConstellationMatcher: no match above threshold",
			"best_score", bestScore,
			"threshold", matchThreshold,
		)
		return "", nil
	}

	logger.InfoContext(ctx, "ConstellationMatcher: match found",
		"constellation_id", bestID,
		"score", bestScore,
	)
	return bestID, nil
}

// constellationEmbedding returns the cached embedding if available,
// otherwise computes one on the fly.
func (m *ConstellationMatcher) constellationEmbedding(ctx context.Context, c domain.Constellation) ([]float32, error) {
	if len(c.TopicEmbedding) > 0 {
		return c.TopicEmbedding, nil
	}

	result, err := m.client.CreateEmbedding(ctx, c.Topic)
	if err != nil {
		return nil, err
	}
	return result.Embedding, nil
}
