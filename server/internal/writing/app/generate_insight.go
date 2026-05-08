package app

import (
	"context"
	"fmt"
	"time"

	"ego-server/internal/writing/domain"
)

// GenerateInsightUseCase orchestrates the generation of an Insight
// based on a Moment and its Echo.
type GenerateInsightUseCase struct {
	insights       domain.InsightRepository
	insightGen     domain.InsightGenerator
	ids            IDGenerator
}

func NewGenerateInsightUseCase(
	insights domain.InsightRepository,
	insightGen domain.InsightGenerator,
	ids IDGenerator,
) *GenerateInsightUseCase {
	return &GenerateInsightUseCase{
		insights:   insights,
		insightGen: insightGen,
		ids:        ids,
	}
}

type GenerateInsightInput struct {
	MomentID string
	EchoID   string
	UserID   string
}

func (uc *GenerateInsightUseCase) Execute(ctx context.Context, input GenerateInsightInput) (*domain.Insight, error) {
	if input.MomentID == "" {
		return nil, domain.ErrEmptyContent
	}

	insight, err := uc.insightGen.Generate(ctx, input.MomentID, input.EchoID)
	if err != nil {
		return nil, fmt.Errorf("generate insight: %w", err)
	}
	if insight == nil {
		return nil, nil
	}

	insight.ID = uc.ids.New()
	insight.MomentID = input.MomentID
	insight.EchoID = input.EchoID
	insight.UserID = input.UserID
	insight.CreatedAt = time.Now()

	if err := uc.insights.Create(ctx, insight); err != nil {
		return nil, fmt.Errorf("persist insight: %w", err)
	}

	return insight, nil
}
