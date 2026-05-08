package domain

import (
	"context"

	writingdomain "ego-server/internal/writing/domain"
)

// StarRepository persists stars.
type StarRepository interface {
	Create(ctx context.Context, star *Star) error
	FindByTraceID(ctx context.Context, traceID string) (*Star, error)
	FindByIDs(ctx context.Context, ids []string) ([]Star, error)
}

// ConstellationRepository persists constellations.
type ConstellationRepository interface {
	Create(ctx context.Context, c *Constellation) error
	Update(ctx context.Context, c *Constellation) error
	FindAllByUserID(ctx context.Context, userID string) ([]Constellation, error)
	FindByID(ctx context.Context, id string) (*Constellation, error)
	FindByStarID(ctx context.Context, starID string) (*Constellation, error)
}

// TraceReader reads traces from the writing module.
type TraceReader interface {
	GetTraceByID(ctx context.Context, id string) (*writingdomain.Trace, error)
	ListMomentsByTraceID(ctx context.Context, traceID string) ([]writingdomain.Moment, error)
}

// MomentReader reads moments from the writing module.
type MomentReader interface {
	ListByTraceID(ctx context.Context, traceID string) ([]writingdomain.Moment, error)
}

// TraceStasher marks a trace as stashed.
type TraceStasher interface {
	MarkStashed(ctx context.Context, traceID string) error
}

// TopicGenerator generates a Star.topic from moments.
type TopicGenerator interface {
	Generate(ctx context.Context, moments []writingdomain.Moment) (string, error)
}

// ConstellationMatcher finds a matching constellation for a topic.
// Returns the constellation ID, or empty string if no match.
type ConstellationMatcher interface {
	FindMatch(ctx context.Context, topic string, existing []Constellation) (string, error)
}

// ConstellationAssetGenerator generates constellation-level assets.
type ConstellationAssetGenerator interface {
	Generate(ctx context.Context, moments []writingdomain.Moment) (name string, insight string, prompts []string, err error)
}
