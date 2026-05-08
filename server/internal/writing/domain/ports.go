package domain

import "context"

// TraceRepository is the persistence contract for Trace aggregates.
type TraceRepository interface {
	Create(ctx context.Context, trace *Trace) error
	GetByID(ctx context.Context, id string) (*Trace, error)
	Update(ctx context.Context, trace *Trace) error
	Delete(ctx context.Context, id string) error
}

// MomentRepository is the persistence contract for Moment entities.
type MomentRepository interface {
	Create(ctx context.Context, moment *Moment) error
	GetByID(ctx context.Context, id string) (*Moment, error)
	ListByTraceID(ctx context.Context, traceID string) ([]Moment, error)
	ListByUserID(ctx context.Context, userID string) ([]Moment, error)
}

// MomentReader is the cross-module read-only contract for Moments.
// Used by Timeline, Starmap, and Conversation modules.
type MomentReader interface {
	GetByID(ctx context.Context, id string) (*Moment, error)
	ListByUserID(ctx context.Context, userID string, cursor string, pageSize int32) (moments []Moment, nextCursor string, hasMore bool, err error)
	RandomByUserID(ctx context.Context, userID string, count int32) ([]Moment, error)
}

// TraceReader is the cross-module read-only contract for Traces.
type TraceReader interface {
	GetTraceByID(ctx context.Context, id string) (*Trace, error)
	ListMomentsByTraceID(ctx context.Context, traceID string) ([]Moment, error)
	ListTracesByUserID(ctx context.Context, userID string, cursor string, pageSize int32) (traces []Trace, nextCursor string, hasMore bool, err error)
}

// EchoRepository is the persistence contract for Echo entities.
type EchoRepository interface {
	Create(ctx context.Context, echo *Echo) error
	FindByMomentID(ctx context.Context, momentID string) (*Echo, error)
}

// InsightRepository is the persistence contract for Insight entities.
type InsightRepository interface {
	Create(ctx context.Context, insight *Insight) error
	FindByMomentID(ctx context.Context, momentID string) (*Insight, error)
}

// EmbeddingGenerator converts text content into embedding vectors.
type EmbeddingGenerator interface {
	Generate(ctx context.Context, content string) ([]EmbeddingEntry, error)
}

// EchoMatcher finds historical Moments that resonate with the current Moment.
type EchoMatcher interface {
	Match(ctx context.Context, current *Moment, history []Moment) ([]MatchedMoment, error)
}

// InsightGenerator produces an AI observation from a Moment and its Echo.
type InsightGenerator interface {
	Generate(ctx context.Context, momentID, echoID string) (*Insight, error)
}
