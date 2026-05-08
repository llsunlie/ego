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
// Used by Starmap to read all Moments under a Trace during StashTrace.
type TraceReader interface {
	GetTraceByID(ctx context.Context, id string) (*Trace, error)
	ListMomentsByTraceID(ctx context.Context, traceID string) ([]Moment, error)
}

// EmbeddingGenerator converts text content into a vector embedding.
type EmbeddingGenerator interface {
	Generate(ctx context.Context, content string) ([]float32, error)
}

// EchoMatcher finds historical Moments that resonate with the current Moment.
type EchoMatcher interface {
	Match(ctx context.Context, current *Moment, history []Moment) (*Echo, error)
}

// InsightGenerator produces a current-session observation from the current content
// and the echoed historical Moment.
type InsightGenerator interface {
	Generate(ctx context.Context, currentContent string, echoMomentID string) (*Insight, error)
}
