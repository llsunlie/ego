package app

import (
	"context"
	"sort"

	platformai "ego-server/internal/platform/ai"
	"ego-server/internal/platform/logging"
	"ego-server/internal/writing/domain"
)

const echoSimilarityThreshold = 0.55

// DefaultEchoMatcher finds historical Moments that resonate with the current
// Moment by comparing their embeddings with cosine similarity.
type DefaultEchoMatcher struct{}

func NewDefaultEchoMatcher() DefaultEchoMatcher {
	return DefaultEchoMatcher{}
}

func (DefaultEchoMatcher) Match(ctx context.Context, current *domain.Moment, history []domain.Moment) ([]domain.MatchedMoment, error) {
	logger := logging.FromContext(ctx)

	if len(current.Embeddings) == 0 {
		logger.WarnContext(ctx, "current moment has no embedding, skipping echo match", "moment_id", current.ID)
		return nil, nil
	}
	curEmb := current.Embeddings[0].Embedding

	var matches []domain.MatchedMoment
	skipped := 0
	for _, h := range history {
		if len(h.Embeddings) == 0 {
			skipped++
			continue
		}
		sim := platformai.CosineSimilarity(curEmb, h.Embeddings[0].Embedding)
		if sim >= echoSimilarityThreshold {
			matches = append(matches, domain.MatchedMoment{MomentID: h.ID, Similarity: sim})
		}
	}

	logger.DebugContext(ctx, "echo match done",
		"history_size", len(history),
		"skipped_no_embedding", skipped,
		"matched", len(matches),
		"threshold", echoSimilarityThreshold,
	)

	sort.Slice(matches, func(i, j int) bool { return matches[i].Similarity > matches[j].Similarity })
	return matches, nil
}
