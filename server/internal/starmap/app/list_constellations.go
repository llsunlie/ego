package app

import (
	"context"
	"fmt"

	"ego-server/internal/starmap/domain"
)

type ListConstellationsUseCase struct {
	constellations domain.ConstellationRepository
	stars          domain.StarRepository
}

func NewListConstellationsUseCase(
	constellations domain.ConstellationRepository,
	stars domain.StarRepository,
) *ListConstellationsUseCase {
	return &ListConstellationsUseCase{constellations: constellations, stars: stars}
}

type ListConstellationsOutput struct {
	Constellations []domain.Constellation
	TotalStarCount int32
}

func (uc *ListConstellationsUseCase) Execute(ctx context.Context) (*ListConstellationsOutput, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user_id not found in context")
	}

	all, err := uc.constellations.FindAllByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list constellations: %w", err)
	}

	clusteredStarIDs := make(map[string]bool)
	for _, c := range all {
		for _, sid := range c.StarIDs {
			clusteredStarIDs[sid] = true
		}
	}

	stars, err := uc.stars.FindAllByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list stars: %w", err)
	}

	for _, s := range stars {
		if clusteredStarIDs[s.ID] {
			continue
		}
		all = append(all, domain.Constellation{
			ID:                   s.ID,
			UserID:               s.UserID,
			Name:                 s.Topic,
			ConstellationInsight: "正在分析这些想法，稍后就会汇聚成星座…",
			StarIDs:              []string{s.ID},
			TopicPrompts:         nil,
			Topic:                s.Topic,
			CreatedAt:            s.CreatedAt,
			UpdatedAt:            s.CreatedAt,
		})
		clusteredStarIDs[s.ID] = true
	}

	return &ListConstellationsOutput{
		Constellations: all,
		TotalStarCount: int32(len(clusteredStarIDs)),
	}, nil
}
