package app

import (
	"context"

	"ego-server/internal/writing/domain"
)

// DefaultEchoMatcher is the MVP echo matching policy. It keeps the workflow
// usable until semantic matching is backed by embeddings/vector search.
type DefaultEchoMatcher struct{}

func NewDefaultEchoMatcher() DefaultEchoMatcher {
	return DefaultEchoMatcher{}
}

func (DefaultEchoMatcher) Match(_ context.Context, _ *domain.Moment, history []domain.Moment) ([]domain.MatchedMoment, error) {
	if len(history) == 0 {
		return nil, nil
	}

	matches := make([]domain.MatchedMoment, 0, len(history))
	for _, h := range history {
		matches = append(matches, domain.MatchedMoment{
			MomentID:   h.ID,
			Similarity: 0.85,
		})
	}
	return matches, nil
}
