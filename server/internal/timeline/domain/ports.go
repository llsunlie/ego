package domain

import (
	"context"

	writingdomain "ego-server/internal/writing/domain"
)

// MomentReader is the read-only contract for Moment queries.
type MomentReader interface {
	GetByID(ctx context.Context, id string) (*writingdomain.Moment, error)
	ListByUserID(ctx context.Context, userID string, cursor string, pageSize int32) ([]writingdomain.Moment, string, bool, error)
	RandomByUserID(ctx context.Context, userID string, count int32) ([]writingdomain.Moment, error)
}

// TraceReader is the read-only contract for Trace queries.
type TraceReader interface {
	GetTraceByID(ctx context.Context, id string) (*writingdomain.Trace, error)
	ListMomentsByTraceID(ctx context.Context, traceID string) ([]writingdomain.Moment, error)
	ListTracesByUserID(ctx context.Context, userID string, cursor string, pageSize int32) ([]writingdomain.Trace, string, bool, error)
}

// EchoReader is the read-only contract for Echo queries.
type EchoReader interface {
	FindByMomentID(ctx context.Context, momentID string) (*writingdomain.Echo, error)
}

// InsightReader is the read-only contract for Insight queries.
type InsightReader interface {
	FindByMomentID(ctx context.Context, momentID string) (*writingdomain.Insight, error)
}
