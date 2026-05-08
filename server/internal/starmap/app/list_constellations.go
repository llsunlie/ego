package app

import (
	"context"
	"fmt"

	"ego-server/internal/starmap/domain"
)

type ListConstellationsUseCase struct {
	constellations domain.ConstellationRepository
}

func NewListConstellationsUseCase(constellations domain.ConstellationRepository) *ListConstellationsUseCase {
	return &ListConstellationsUseCase{constellations: constellations}
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

	var totalStars int32
	for _, c := range all {
		totalStars += int32(len(c.StarIDs))
	}

	return &ListConstellationsOutput{
		Constellations: all,
		TotalStarCount: totalStars,
	}, nil
}
