package app

import (
	"context"
	"fmt"
	"time"

	"ego-server/internal/platform/logging"
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
	profileGen       domain.TraceProfileGenerator
	profileRepo      domain.TraceProfileRepository
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
	return NewStashTraceUseCaseWithTraceProfile(
		traceReader,
		traceStasher,
		stars,
		constellations,
		topicGen,
		constellationMat,
		assetGen,
		nil,
		nil,
		ids,
	)
}

func NewStashTraceUseCaseWithTraceProfile(
	traceReader domain.TraceReader,
	traceStasher domain.TraceStasher,
	stars domain.StarRepository,
	constellations domain.ConstellationRepository,
	topicGen domain.TopicGenerator,
	constellationMat domain.ConstellationMatcher,
	assetGen domain.ConstellationAssetGenerator,
	profileGen domain.TraceProfileGenerator,
	profileRepo domain.TraceProfileRepository,
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
		profileGen:       profileGen,
		profileRepo:      profileRepo,
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

	star := &domain.Star{
		ID:        uc.ids.New(),
		UserID:    userID,
		TraceID:   input.TraceID,
		Topic:     "聚合中",
		CreatedAt: time.Now(),
	}

	if err := uc.stars.Create(ctx, star); err != nil {
		return nil, fmt.Errorf("create star: %w", err)
	}

	if err := uc.traceStasher.MarkStashed(ctx, input.TraceID); err != nil {
		return nil, fmt.Errorf("mark stashed: %w", err)
	}

	go uc.clusterAsync(userID, star.ID, moments)
	go uc.generateTraceProfileAsync(*trace, moments)

	return star, nil
}

func (uc *StashTraceUseCase) generateTraceProfileAsync(trace writingdomain.Trace, moments []writingdomain.Moment) {
	if uc.profileGen == nil || uc.profileRepo == nil {
		return
	}

	ctx := context.Background()
	logger := logging.FromContext(ctx)
	profile, vector, err := uc.profileGen.Generate(ctx, trace, moments)
	if err != nil {
		logger.ErrorContext(ctx, "starmap: async trace profile generation failed",
			"trace_id", trace.ID,
			"error", err,
		)
		return
	}
	if err := uc.profileRepo.Upsert(ctx, profile, vector); err != nil {
		logger.ErrorContext(ctx, "starmap: async trace profile upsert failed",
			"trace_id", trace.ID,
			"error", err,
		)
		return
	}
	logger.InfoContext(ctx, "starmap: async trace profile completed",
		"trace_id", trace.ID,
		"status", profile.Status,
		"has_vector", vector != nil,
	)
}

func (uc *StashTraceUseCase) clusterAsync(userID string, starID string, moments []writingdomain.Moment) {
	ctx := context.Background()
	logger := logging.FromContext(ctx)

	topic, err := uc.topicGen.Generate(ctx, moments)
	if err != nil {
		logger.ErrorContext(ctx, "starmap: async topic generation failed",
			"star_id", starID,
			"error", err,
		)
		return
	}

	if err := uc.stars.UpdateTopic(ctx, starID, topic); err != nil {
		logger.ErrorContext(ctx, "starmap: async update star topic failed",
			"star_id", starID,
			"error", err,
		)
		return
	}

	existing, err := uc.constellations.FindAllByUserID(ctx, userID)
	if err != nil {
		logger.ErrorContext(ctx, "starmap: async list constellations failed",
			"star_id", starID,
			"error", err,
		)
		return
	}

	matchID, err := uc.constellationMat.FindMatch(ctx, topic, existing)
	if err != nil {
		logger.ErrorContext(ctx, "starmap: async constellation match failed",
			"star_id", starID,
			"error", err,
		)
		return
	}

	if matchID != "" {
		c, err := uc.constellations.FindByID(ctx, matchID)
		if err != nil {
			logger.ErrorContext(ctx, "starmap: async find constellation failed",
				"star_id", starID,
				"constellation_id", matchID,
				"error", err,
			)
			return
		}

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

		topic, topicEmb, name, insight, prompts, err := uc.assetGen.Generate(ctx, allMoments)
		if err != nil {
			logger.ErrorContext(ctx, "starmap: async asset regeneration failed",
				"star_id", starID,
				"constellation_id", matchID,
				"error", err,
			)
			return
		}

		c.Topic = topic
		c.TopicEmbedding = topicEmb
		c.Name = name
		c.ConstellationInsight = insight
		c.TopicPrompts = prompts
		c.StarIDs = append(c.StarIDs, starID)
		c.UpdatedAt = time.Now()

		if err := uc.constellations.Update(ctx, c); err != nil {
			logger.ErrorContext(ctx, "starmap: async update constellation failed",
				"star_id", starID,
				"constellation_id", matchID,
				"error", err,
			)
			return
		}
	} else {
		topic, topicEmb, name, insight, prompts, err := uc.assetGen.Generate(ctx, moments)
		if err != nil {
			logger.ErrorContext(ctx, "starmap: async asset generation failed",
				"star_id", starID,
				"error", err,
			)
			return
		}

		now := time.Now()
		c := &domain.Constellation{
			ID:                   uc.ids.New(),
			UserID:               userID,
			Topic:                topic,
			TopicEmbedding:       topicEmb,
			Name:                 name,
			ConstellationInsight: insight,
			StarIDs:              []string{starID},
			TopicPrompts:         prompts,
			CreatedAt:            now,
			UpdatedAt:            now,
		}

		if err := uc.constellations.Create(ctx, c); err != nil {
			logger.ErrorContext(ctx, "starmap: async create constellation failed",
				"star_id", starID,
				"error", err,
			)
			return
		}
	}

	logger.InfoContext(ctx, "starmap: async clustering completed", "star_id", starID)
}
