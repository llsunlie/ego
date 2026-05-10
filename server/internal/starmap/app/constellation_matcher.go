package app

import (
	"context"

	"ego-server/internal/starmap/domain"
)

// DefaultConstellationMatcher is the MVP constellation-matching policy.
// It always returns no match, producing a lone-star constellation each time.
type DefaultConstellationMatcher struct{}

func NewDefaultConstellationMatcher() DefaultConstellationMatcher {
	return DefaultConstellationMatcher{}
}

func (DefaultConstellationMatcher) FindMatch(_ context.Context, _ string, _ []domain.Constellation) (string, error) {
	return "", nil
}

// Compile-time check.
var _ domain.ConstellationMatcher = DefaultConstellationMatcher{}
