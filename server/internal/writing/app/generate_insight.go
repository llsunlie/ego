package app

import (
	"context"
	"fmt"

	"ego-server/internal/writing/domain"
)

// GenerateInsightUseCase orchestrates the generation of a current-session Insight
// based on the user's current content and a matched echo Moment.
type GenerateInsightUseCase struct {
	insight domain.InsightGenerator
}

func NewGenerateInsightUseCase(insight domain.InsightGenerator) *GenerateInsightUseCase {
	return &GenerateInsightUseCase{insight: insight}
}

type GenerateInsightInput struct {
	CurrentContent string
	EchoMomentID   string
}

func (uc *GenerateInsightUseCase) Execute(ctx context.Context, input GenerateInsightInput) (*domain.Insight, error) {
	if input.CurrentContent == "" {
		return nil, domain.ErrEmptyContent
	}

	insight, err := uc.insight.Generate(ctx, input.CurrentContent, input.EchoMomentID)
	if err != nil {
		return nil, fmt.Errorf("generate insight: %w", err)
	}

	return insight, nil
}
