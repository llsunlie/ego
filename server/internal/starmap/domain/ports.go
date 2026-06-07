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
	FindAllByUserID(ctx context.Context, userID string) ([]Star, error)
	UpdateTopic(ctx context.Context, starID string, topic string) error
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

// ConstellationAssetGenerator generates constellation-level display assets.
type ConstellationAssetGenerator interface {
	Generate(ctx context.Context, moments []writingdomain.Moment) (topic string, topicEmbedding []float32, name string, insight string, prompts []string, err error)
}

// TraceProfileGenerator builds a persistent algorithm profile for a stashed Trace.
type TraceProfileGenerator interface {
	Generate(ctx context.Context, trace writingdomain.Trace, moments []writingdomain.Moment) (*TraceProfile, *TraceProfileVector, error)
}

// ConstellationBorderlineJudge judges ambiguous constellation matches.
type ConstellationBorderlineJudge interface {
	Judge(ctx context.Context, input ConstellationBorderlineJudgeInput) (*ConstellationBorderlineJudgement, error)
}

// TraceProfileRepository persists TraceProfiles and their optional vectors.
type TraceProfileRepository interface {
	Upsert(ctx context.Context, profile *TraceProfile, vector *TraceProfileVector) error
}

// ConstellationProfileRepository persists long-term constellation algorithm profiles.
type ConstellationProfileRepository interface {
	FindCandidates(ctx context.Context, userID string, embedding []float32, limit int) ([]ConstellationProfileCandidate, error)
	Upsert(ctx context.Context, profile *ConstellationProfile, vector *ConstellationProfileVector) error
	AddMembership(ctx context.Context, membership ConstellationMembership) error
}
