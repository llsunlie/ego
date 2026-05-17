package app

import (
	"context"

	"ego-server/internal/timeline/domain"
	writingdomain "ego-server/internal/writing/domain"
)

// GetTraceDetailUseCase returns a full trace detail with moments, echos, and insights.
type GetTraceDetailUseCase struct {
	traces   domain.TraceReader
	echos    domain.EchoReader
	insights domain.InsightReader
}

func NewGetTraceDetailUseCase(
	traces domain.TraceReader,
	echos domain.EchoReader,
	insights domain.InsightReader,
) *GetTraceDetailUseCase {
	return &GetTraceDetailUseCase{
		traces:   traces,
		echos:    echos,
		insights: insights,
	}
}

type GetTraceDetailInput struct {
	TraceID string
}

type GetTraceDetailOutput struct {
	Trace writingdomain.Trace
	Items []writingdomain.TraceItem
}

func (uc *GetTraceDetailUseCase) Execute(ctx context.Context, input GetTraceDetailInput) (*GetTraceDetailOutput, error) {
	trace, err := uc.traces.GetTraceByID(ctx, input.TraceID)
	if err != nil {
		return nil, err
	}

	moments, err := uc.traces.ListMomentsByTraceID(ctx, input.TraceID)
	if err != nil {
		return nil, err
	}

	items := make([]writingdomain.TraceItem, len(moments))
	for i, m := range moments {
		echo, _ := uc.echos.FindByMomentID(ctx, m.ID)
		var echos []writingdomain.Echo
		if echo != nil {
			echos = []writingdomain.Echo{*echo}
		}

		var insight *writingdomain.Insight
		insight, _ = uc.insights.FindByMomentID(ctx, m.ID)

		items[i] = writingdomain.TraceItem{
			Moment:  m,
			Echos:   echos,
			Insight: insight,
		}
	}

	return &GetTraceDetailOutput{
		Trace: *trace,
		Items: items,
	}, nil
}
