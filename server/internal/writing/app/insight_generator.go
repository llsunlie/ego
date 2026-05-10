package app

import (
	"context"

	"ego-server/internal/writing/domain"
)

// DefaultInsightGenerator is the MVP insight policy used until real AI-backed
// insight generation is wired in.
type DefaultInsightGenerator struct{}

func NewDefaultInsightGenerator() DefaultInsightGenerator {
	return DefaultInsightGenerator{}
}

func (DefaultInsightGenerator) Generate(_ context.Context, momentID, _ string) (*domain.Insight, error) {
	return &domain.Insight{
		MomentID:         momentID,
		Text:             "你似乎在反复思考与自尊相关的话题。当你感到被否定时，童年时期形成的防御模式会被激活。",
		RelatedMomentIDs: []string{},
	}, nil
}
