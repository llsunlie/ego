package app

import (
	"context"
	"math/rand/v2"

	"ego-server/internal/starmap/domain"
)

// DefaultConstellationMatcher is the MVP constellation-matching policy.
// It has a 65% chance to cluster with an existing constellation (simulates AI matching).
type DefaultConstellationMatcher struct{}

func NewDefaultConstellationMatcher() DefaultConstellationMatcher {
	return DefaultConstellationMatcher{}
}

func (DefaultConstellationMatcher) FindMatch(_ context.Context, _ string, existing []domain.Constellation) (string, error) {
	if len(existing) == 0 {
		return "", nil
	}
	if rand.IntN(100) < 65 {
		return existing[rand.IntN(len(existing))].ID, nil
	}
	return "", nil
}

// Compile-time check.
var _ domain.ConstellationMatcher = DefaultConstellationMatcher{}
