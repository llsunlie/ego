package app

import (
	"context"
	"fmt"
	"time"

	"ego-server/internal/starmap/domain"
	writingdomain "ego-server/internal/writing/domain"
)

type StashTraceUseCase struct {
	traceReader      domain.TraceReader
	traceStasher     domain.TraceStasher
	stars            domain.StarRepository
	constellations   domain.ConstellationRepository
	topicGen         domain.TopicGenerator
	constellationMat domain.ConstellationMatcher
	assetGen         domain.ConstellationAssetGenerator
	ids              IDGenerator
}

func NewStashTraceUseCase(
	traceReader domain.TraceReader,
	traceStasher domain.TraceStasher,
	stars domain.StarRepository,
	constellations domain.ConstellationRepository,
	topicGen domain.TopicGenerator,
	constellationMat domain.ConstellationMatcher,
	assetGen domain.ConstellationAssetGenerator,
	ids IDGenerator,
) *StashTraceUseCase {
	return &StashTraceUseCase{
		traceReader:      traceReader,
		traceStasher:     traceStasher,
		stars:            stars,
		constellations:   constellations,
		topicGen:         topicGen,
		constellationMat: constellationMat,
		assetGen:         assetGen,
		ids:              ids,
	}
}

type StashTraceInput struct {
	TraceID string
}

func (uc *StashTraceUseCase) Execute(ctx context.Context, input StashTraceInput) (*domain.Star, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user_id not found in context")
	}

	trace, err := uc.traceReader.GetTraceByID(ctx, input.TraceID)
	if err != nil {
		return nil, fmt.Errorf("get trace: %w", err)
	}
	if trace.UserID != userID {
		return nil, domain.ErrTraceNotFound
	}
	if trace.Stashed {
		return nil, domain.ErrTraceAlreadyStashed
	}

	moments, err := uc.traceReader.ListMomentsByTraceID(ctx, input.TraceID)
	if err != nil {
		return nil, fmt.Errorf("list moments: %w", err)
	}

	topic, err := uc.topicGen.Generate(ctx, moments)
	if err != nil {
		return nil, fmt.Errorf("generate topic: %w", err)
	}

	star := &domain.Star{
		ID:        uc.ids.New(),
		UserID:    userID,
		TraceID:   input.TraceID,
		Topic:     topic,
		CreatedAt: time.Now(),
	}

	if err := uc.stars.Create(ctx, star); err != nil {
		return nil, fmt.Errorf("create star: %w", err)
	}

	if err := uc.traceStasher.MarkStashed(ctx, input.TraceID); err != nil {
		return nil, fmt.Errorf("mark stashed: %w", err)
	}

	// Constellation matching: find existing or create new lone-star constellation
	existing, err := uc.constellations.FindAllByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list constellations: %w", err)
	}

	matchID, err := uc.constellationMat.FindMatch(ctx, topic, existing)
	if err != nil {
		return nil, fmt.Errorf("match constellation: %w", err)
	}

	if matchID != "" {
		c, err := uc.constellations.FindByID(ctx, matchID)
		if err != nil {
			return nil, fmt.Errorf("find constellation: %w", err)
		}

		// Collect moments from all stars in the constellation for asset regeneration
		allMoments := make([]writingdomain.Moment, 0, len(moments)*(len(c.StarIDs)+1))
		allMoments = append(allMoments, moments...)
		for _, sid := range c.StarIDs {
			stars, err := uc.stars.FindByIDs(ctx, []string{sid})
			if err != nil || len(stars) == 0 {
				continue
			}
			sm, err := uc.traceReader.ListMomentsByTraceID(ctx, stars[0].TraceID)
			if err != nil {
				continue
			}
			allMoments = append(allMoments, sm...)
		}

		name, insight, prompts, err := uc.assetGen.Generate(ctx, allMoments)
		if err != nil {
			return nil, fmt.Errorf("regenerate assets: %w", err)
		}

		c.Name = name
		c.ConstellationInsight = insight
		c.TopicPrompts = prompts
		c.StarIDs = append(c.StarIDs, star.ID)
		c.UpdatedAt = time.Now()

		if err := uc.constellations.Update(ctx, c); err != nil {
			return nil, fmt.Errorf("update constellation: %w", err)
		}
	} else {
		name, insight, prompts, err := uc.assetGen.Generate(ctx, moments)
		if err != nil {
			return nil, fmt.Errorf("generate assets: %w", err)
		}

		now := time.Now()
		c := &domain.Constellation{
			ID:                   uc.ids.New(),
			UserID:               userID,
			Name:                 name,
			ConstellationInsight: insight,
			StarIDs:              []string{star.ID},
			TopicPrompts:         prompts,
			CreatedAt:            now,
			UpdatedAt:            now,
		}

		if err := uc.constellations.Create(ctx, c); err != nil {
			return nil, fmt.Errorf("create constellation: %w", err)
		}
	}

	return star, nil
}
