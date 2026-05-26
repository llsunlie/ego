package app

import (
	"context"
	"fmt"

	"ego-server/internal/starmap/domain"
	writingdomain "ego-server/internal/writing/domain"
)

type GetConstellationUseCase struct {
	constellations domain.ConstellationRepository
	stars          domain.StarRepository
	traceReader    domain.TraceReader
}

func NewGetConstellationUseCase(
	constellations domain.ConstellationRepository,
	stars domain.StarRepository,
	traceReader domain.TraceReader,
) *GetConstellationUseCase {
	return &GetConstellationUseCase{
		constellations: constellations,
		stars:          stars,
		traceReader:    traceReader,
	}
}

type GetConstellationInput struct {
	ConstellationID string
}

type GetConstellationOutput struct {
	Constellation domain.Constellation
	Moments       []writingdomain.Moment
	Stars         []domain.Star
}

func (uc *GetConstellationUseCase) Execute(ctx context.Context, input GetConstellationInput) (*GetConstellationOutput, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user_id not found in context")
	}

	c, err := uc.constellations.FindByID(ctx, input.ConstellationID)
	if err != nil {
		// Constellation not found — check if the ID matches a pending star (fake constellation)
		stars, starErr := uc.stars.FindByIDs(ctx, []string{input.ConstellationID})
		if starErr != nil || len(stars) == 0 || stars[0].UserID != userID {
			return nil, fmt.Errorf("find constellation: %w", err)
		}
		s := stars[0]
		now := s.CreatedAt
		return &GetConstellationOutput{
			Constellation: domain.Constellation{
				ID:                   s.ID,
				UserID:               s.UserID,
				Name:                 s.Topic,
				ConstellationInsight: "正在分析这些想法，稍后就会汇聚成星座…",
				StarIDs:              []string{s.ID},
				CreatedAt:            now,
				UpdatedAt:            now,
			},
			Stars:   []domain.Star{s},
			Moments: nil,
		}, nil
	}
	if c.UserID != userID {
		return nil, domain.ErrConstellationNotFound
	}

	stars, err := uc.stars.FindByIDs(ctx, c.StarIDs)
	if err != nil {
		return nil, fmt.Errorf("find stars: %w", err)
	}

	var moments []writingdomain.Moment
	for _, star := range stars {
		ms, err := uc.traceReader.ListMomentsByTraceID(ctx, star.TraceID)
		if err != nil {
			continue
		}
		moments = append(moments, ms...)
	}

	return &GetConstellationOutput{
		Constellation: *c,
		Moments:       moments,
		Stars:         stars,
	}, nil
}
