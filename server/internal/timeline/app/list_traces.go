package app

import (
	"context"

	"ego-server/internal/timeline/domain"
	writingdomain "ego-server/internal/writing/domain"
)

// ListTracesUseCase returns a cursor-paginated list of Traces for the user.
type ListTracesUseCase struct {
	traces domain.TraceReader
}

func NewListTracesUseCase(traces domain.TraceReader) *ListTracesUseCase {
	return &ListTracesUseCase{traces: traces}
}

type ListTracesInput struct {
	UserID   string
	Cursor   string
	PageSize int32
}

type ListTracesOutput struct {
	Traces     []writingdomain.Trace
	NextCursor string
	HasMore    bool
}

func (uc *ListTracesUseCase) Execute(ctx context.Context, input ListTracesInput) (*ListTracesOutput, error) {
	pageSize := input.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	traces, nextCursor, hasMore, err := uc.traces.ListTracesByUserID(ctx, input.UserID, input.Cursor, pageSize)
	if err != nil {
		return nil, err
	}

	return &ListTracesOutput{
		Traces:     traces,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}
