package app

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"ego-server/internal/platform/logging"
	"ego-server/internal/starmap/domain"
	writingdomain "ego-server/internal/writing/domain"
)

const (
	constellationCandidateLimit       = 10
	constellationStrongMatchThreshold = 0.72
	constellationMiddleMatchThreshold = 0.60
	constellationExplainableThreshold = 0.58
	maxSecondaryConstellationMatches  = 2
	profileClusteringMaxAttempts      = 3

	constellationProfileSimilarityWeight  = 0.45
	constellationCentroidSimilarityWeight = 0.20
	constellationKeywordOverlapWeight     = 0.12
	constellationSceneOverlapWeight       = 0.08
	constellationEmotionOverlapWeight     = 0.07
	constellationPatternTagsOverlapWeight = 0.08

	singleTraceProfileSimilarityWeight  = 0.60
	singleTraceKeywordOverlapWeight     = 0.14
	singleTraceSceneOverlapWeight       = 0.10
	singleTraceEmotionOverlapWeight     = 0.08
	singleTracePatternTagsOverlapWeight = 0.08
)

type StashTraceUseCase struct {
	traceReader       domain.TraceReader
	traceStasher      domain.TraceStasher
	stars             domain.StarRepository
	constellations    domain.ConstellationRepository
	assetGen          domain.ConstellationAssetGenerator
	profileGen        domain.TraceProfileGenerator
	profileRepo       domain.TraceProfileRepository
	constellationProf domain.ConstellationProfileRepository
	ids               IDGenerator
}

func NewStashTraceUseCase(
	traceReader domain.TraceReader,
	traceStasher domain.TraceStasher,
	stars domain.StarRepository,
	constellations domain.ConstellationRepository,
	assetGen domain.ConstellationAssetGenerator,
	ids IDGenerator,
) *StashTraceUseCase {
	return NewStashTraceUseCaseWithTraceProfile(
		traceReader,
		traceStasher,
		stars,
		constellations,
		assetGen,
		nil,
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
	assetGen domain.ConstellationAssetGenerator,
	profileGen domain.TraceProfileGenerator,
	profileRepo domain.TraceProfileRepository,
	constellationProf domain.ConstellationProfileRepository,
	ids IDGenerator,
) *StashTraceUseCase {
	return &StashTraceUseCase{
		traceReader:       traceReader,
		traceStasher:      traceStasher,
		stars:             stars,
		constellations:    constellations,
		assetGen:          assetGen,
		profileGen:        profileGen,
		profileRepo:       profileRepo,
		constellationProf: constellationProf,
		ids:               ids,
	}
}

type StashTraceInput struct {
	TraceID string
}

func (uc *StashTraceUseCase) Execute(ctx context.Context, input StashTraceInput) (*domain.Star, error) {
	logger := logging.FromContext(ctx)
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

	bgCtx := logging.WithLogger(context.Background(), logger)
	go uc.clusterWithProfileAsync(bgCtx, *trace, *star, moments)

	return star, nil
}

type scoredConstellationCandidate struct {
	candidate          domain.ConstellationProfileCandidate
	score              float64
	profileSimilarity  float64
	centroidSimilarity float64
	keywordOverlap     float64
	sceneOverlap       float64
	emotionOverlap     float64
	patternTagsOverlap float64
	matchedKeywords    []string
	matchedScenes      []string
	matchedEmotions    []string
	matchedPatternTags []string
	explainableMiddle  bool
	dimensions         []string
	reason             string
}

func (uc *StashTraceUseCase) clusterWithProfileAsync(ctx context.Context, trace writingdomain.Trace, star domain.Star, moments []writingdomain.Moment) {
	logger := logging.FromContext(ctx)
	if uc.profileGen == nil || uc.profileRepo == nil || uc.constellationProf == nil {
		logger.ErrorContext(ctx, "starmap profile clustering dependency missing",
			"trace_id", trace.ID,
			"star_id", star.ID,
			"has_profile_generator", uc.profileGen != nil,
			"has_profile_repository", uc.profileRepo != nil,
			"has_constellation_profile_repository", uc.constellationProf != nil,
			"recovery", "pending_message_queue",
		)
		return
	}

	profile, vector, err := uc.generateTraceProfileWithRetry(ctx, trace, star.ID, moments)
	if err != nil {
		logger.ErrorContext(ctx, "starmap profile clustering trace profile generation exhausted",
			"trace_id", trace.ID,
			"star_id", star.ID,
			"error", err,
			"recovery", "pending_message_queue",
		)
		return
	}
	if err := uc.profileRepo.Upsert(ctx, profile, vector); err != nil {
		logger.ErrorContext(ctx, "starmap profile clustering trace profile upsert failed",
			"trace_id", trace.ID,
			"star_id", star.ID,
			"error", err,
			"recovery", "pending_message_queue",
		)
		return
	}
	if profile.Topic != "" {
		if err := uc.stars.UpdateTopic(ctx, star.ID, profile.Topic); err != nil {
			logger.ErrorContext(ctx, "starmap profile clustering update star topic failed",
				"trace_id", trace.ID,
				"star_id", star.ID,
				"topic", profile.Topic,
				"error", err,
				"recovery", "pending_message_queue",
			)
			return
		}
		star.Topic = profile.Topic
	}
	logger.DebugContext(ctx, "starmap profile clustering trace profile ready",
		"trace_id", trace.ID,
		"star_id", star.ID,
		"topic", profile.Topic,
		"status", profile.Status,
		"keyword_count", len(profile.Keywords),
		"keywords", profile.Keywords,
		"emotion_count", len(profile.Emotions),
		"emotions", profile.Emotions,
		"scene_count", len(profile.Scenes),
		"scenes", profile.Scenes,
		"has_central_pattern", strings.TrimSpace(profile.CentralPattern) != "",
		"pattern_tag_count", len(profile.PatternTags),
		"pattern_tags", profile.PatternTags,
		"has_vector", vector != nil,
		"vector_dim", traceVectorDim(vector),
	)

	var ranked []scoredConstellationCandidate
	if vector != nil {
		candidates, err := uc.constellationProf.FindCandidates(ctx, trace.UserID, vector.Embedding, constellationCandidateLimit)
		if err != nil {
			logger.ErrorContext(ctx, "starmap profile clustering candidate recall failed",
				"trace_id", trace.ID,
				"star_id", star.ID,
				"error", err,
				"recovery", "pending_message_queue",
			)
			return
		}
		logger.DebugContext(ctx, "starmap profile clustering candidates recalled",
			"trace_id", trace.ID,
			"star_id", star.ID,
			"candidate_limit", constellationCandidateLimit,
			"candidate_count", len(candidates),
			"candidate_summaries", constellationCandidateSummaries(candidates, constellationCandidateLimit),
		)
		ranked = rankConstellationCandidates(profile, vector, candidates)
		logger.DebugContext(ctx, "starmap profile clustering candidates ranked",
			"trace_id", trace.ID,
			"star_id", star.ID,
			"strong_threshold", constellationStrongMatchThreshold,
			"middle_threshold", constellationMiddleMatchThreshold,
			"score_weights", constellationScoreWeights(),
			"candidate_count", len(ranked),
			"ranked_candidates", scoredConstellationCandidateSummaries(ranked, constellationCandidateLimit),
		)
	} else {
		logger.DebugContext(ctx, "starmap profile clustering candidate recall skipped",
			"trace_id", trace.ID,
			"star_id", star.ID,
			"reason", "missing_trace_profile_vector",
		)
	}

	var primaryID string
	var primaryDimensions []string
	var primaryScore float64
	primaryDecision := "create_new"
	if len(ranked) > 0 && isPrimaryAttachCandidate(ranked[0]) {
		primaryID = ranked[0].candidate.Profile.ConstellationID
		primaryDimensions = ranked[0].dimensions
		primaryScore = ranked[0].score
		primaryDecision = "attach_existing"
		if ranked[0].score < constellationStrongMatchThreshold {
			primaryDecision = "attach_existing_middle"
		}
		logger.DebugContext(ctx, "starmap profile clustering primary decision",
			"trace_id", trace.ID,
			"star_id", star.ID,
			"decision", primaryDecision,
			"constellation_id", primaryID,
			"score", primaryScore,
			"strong_threshold", constellationStrongMatchThreshold,
			"strong_threshold_gap", primaryScore-constellationStrongMatchThreshold,
			"middle_threshold", constellationMiddleMatchThreshold,
			"middle_threshold_gap", primaryScore-constellationMiddleMatchThreshold,
			"explainable_threshold", constellationExplainableThreshold,
			"explainable_middle", ranked[0].explainableMiddle,
			"dimensions", primaryDimensions,
			"reason", ranked[0].reason,
			"top_candidate", scoredConstellationCandidateSummary(ranked[0], 1),
		)
		if err := uc.attachStarToConstellation(ctx, primaryID, star, trace, moments, profile, vector, ranked[0], domain.ConstellationMatchTypePrimary, 1.0); err != nil {
			logger.ErrorContext(ctx, "starmap profile clustering attach primary failed",
				"trace_id", trace.ID,
				"star_id", star.ID,
				"constellation_id", primaryID,
				"error", err,
				"recovery", "pending_message_queue",
			)
			return
		}
	} else {
		var err error
		topScore := 0.0
		if len(ranked) > 0 {
			topScore = ranked[0].score
		}
		var topCandidate map[string]any
		if len(ranked) > 0 {
			topCandidate = scoredConstellationCandidateSummary(ranked[0], 1)
		}
		logger.DebugContext(ctx, "starmap profile clustering primary decision",
			"trace_id", trace.ID,
			"star_id", star.ID,
			"decision", primaryDecision,
			"top_score", topScore,
			"strong_threshold", constellationStrongMatchThreshold,
			"strong_threshold_gap", topScore-constellationStrongMatchThreshold,
			"middle_threshold", constellationMiddleMatchThreshold,
			"middle_threshold_gap", topScore-constellationMiddleMatchThreshold,
			"explainable_threshold", constellationExplainableThreshold,
			"candidate_count", len(ranked),
			"top_candidate", topCandidate,
		)
		primaryID, err = uc.createConstellationFromTraceProfile(ctx, star, trace, moments, profile, vector)
		if err != nil {
			logger.ErrorContext(ctx, "starmap profile clustering create primary constellation failed",
				"trace_id", trace.ID,
				"star_id", star.ID,
				"error", err,
				"recovery", "pending_message_queue",
			)
			return
		}
		primaryDimensions = []string{"new_theme"}
		primaryScore = 1.0
	}

	secondaryCount := 0
	for _, candidate := range ranked {
		if secondaryCount >= maxSecondaryConstellationMatches {
			logger.DebugContext(ctx, "starmap profile clustering secondary candidate skipped",
				"trace_id", trace.ID,
				"star_id", star.ID,
				"constellation_id", candidate.candidate.Profile.ConstellationID,
				"score", candidate.score,
				"reason", "secondary_limit_reached",
				"secondary_limit", maxSecondaryConstellationMatches,
			)
			break
		}
		if candidate.candidate.Profile.ConstellationID == primaryID {
			logger.DebugContext(ctx, "starmap profile clustering secondary candidate skipped",
				"trace_id", trace.ID,
				"star_id", star.ID,
				"constellation_id", candidate.candidate.Profile.ConstellationID,
				"score", candidate.score,
				"reason", "same_as_primary",
			)
			continue
		}
		if candidate.score < constellationMiddleMatchThreshold && !candidate.explainableMiddle {
			logger.DebugContext(ctx, "starmap profile clustering secondary candidate skipped",
				"trace_id", trace.ID,
				"star_id", star.ID,
				"constellation_id", candidate.candidate.Profile.ConstellationID,
				"score", candidate.score,
				"threshold", constellationMiddleMatchThreshold,
				"explainable_threshold", constellationExplainableThreshold,
				"explainable_middle", candidate.explainableMiddle,
				"reason", "below_middle_threshold",
			)
			continue
		}
		if !hasDistinctDimensions(primaryDimensions, candidate.dimensions) {
			logger.DebugContext(ctx, "starmap profile clustering secondary candidate skipped",
				"trace_id", trace.ID,
				"star_id", star.ID,
				"constellation_id", candidate.candidate.Profile.ConstellationID,
				"score", candidate.score,
				"primary_dimensions", primaryDimensions,
				"candidate_dimensions", candidate.dimensions,
				"reason", "same_match_dimensions",
			)
			continue
		}
		logger.DebugContext(ctx, "starmap profile clustering secondary decision",
			"trace_id", trace.ID,
			"star_id", star.ID,
			"decision", "attach_secondary",
			"constellation_id", candidate.candidate.Profile.ConstellationID,
			"score", candidate.score,
			"threshold", constellationMiddleMatchThreshold,
			"explainable_middle", candidate.explainableMiddle,
			"dimensions", candidate.dimensions,
			"reason", candidate.reason,
		)
		if err := uc.attachStarToConstellation(ctx, candidate.candidate.Profile.ConstellationID, star, trace, moments, profile, vector, candidate, domain.ConstellationMatchTypeSecondary, 0.5); err != nil {
			logger.WarnContext(ctx, "starmap profile clustering attach secondary failed",
				"trace_id", trace.ID,
				"star_id", star.ID,
				"constellation_id", candidate.candidate.Profile.ConstellationID,
				"score", candidate.score,
				"error", err,
				"recovery", "pending_message_queue",
			)
			continue
		}
		secondaryCount++
	}

	logger.InfoContext(ctx, "starmap profile clustering completed",
		"trace_id", trace.ID,
		"star_id", star.ID,
		"decision", primaryDecision,
		"primary_constellation_id", primaryID,
		"primary_score", primaryScore,
		"secondary_count", secondaryCount,
		"candidate_count", len(ranked),
	)
}

func (uc *StashTraceUseCase) generateTraceProfileWithRetry(ctx context.Context, trace writingdomain.Trace, starID string, moments []writingdomain.Moment) (*domain.TraceProfile, *domain.TraceProfileVector, error) {
	logger := logging.FromContext(ctx)
	var lastErr error
	for attempt := 1; attempt <= profileClusteringMaxAttempts; attempt++ {
		profile, vector, err := uc.profileGen.Generate(ctx, trace, moments)
		if err == nil {
			if attempt > 1 {
				logger.InfoContext(ctx, "starmap trace profile generation retry succeeded",
					"trace_id", trace.ID,
					"star_id", starID,
					"attempt", attempt,
				)
			}
			return profile, vector, nil
		}
		lastErr = err
		logger.WarnContext(ctx, "starmap trace profile generation attempt failed",
			"trace_id", trace.ID,
			"star_id", starID,
			"attempt", attempt,
			"max_attempts", profileClusteringMaxAttempts,
			"error", err,
		)
	}
	return nil, nil, lastErr
}

func (uc *StashTraceUseCase) createConstellationFromTraceProfile(ctx context.Context, star domain.Star, trace writingdomain.Trace, moments []writingdomain.Moment, profile *domain.TraceProfile, vector *domain.TraceProfileVector) (string, error) {
	topic, _, name, insight, prompts, err := uc.assetGen.Generate(ctx, moments)
	if err != nil {
		return "", fmt.Errorf("generate constellation assets: %w", err)
	}
	if profile.Topic != "" {
		topic = profile.Topic
	}
	if strings.TrimSpace(name) == "" {
		name = topic
	}
	if strings.TrimSpace(insight) == "" {
		insight = profile.Summary
	}

	now := time.Now()
	constellationID := uc.ids.New()
	c := &domain.Constellation{
		ID:                   constellationID,
		UserID:               star.UserID,
		Topic:                topic,
		Name:                 name,
		ConstellationInsight: insight,
		StarIDs:              []string{star.ID},
		TopicPrompts:         prompts,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	if err := uc.constellations.Create(ctx, c); err != nil {
		return "", fmt.Errorf("create constellation: %w", err)
	}

	cProfile := constellationProfileFromTraceProfile(constellationID, trace.UserID, profile, 1.0, float64(len(moments)), now)
	var cVector *domain.ConstellationProfileVector
	if vector != nil {
		cVector = &domain.ConstellationProfileVector{
			ConstellationID:   constellationID,
			UserID:            trace.UserID,
			Model:             vector.Model,
			Dim:               vector.Dim,
			ProfileEmbedding:  vector.Embedding,
			CentroidEmbedding: vector.Embedding,
			CreatedAt:         now,
			UpdatedAt:         now,
		}
	}
	if err := uc.constellationProf.Upsert(ctx, cProfile, cVector); err != nil {
		return "", fmt.Errorf("upsert constellation profile: %w", err)
	}
	if err := uc.constellationProf.AddMembership(ctx, domain.ConstellationMembership{
		ConstellationID: constellationID,
		StarID:          star.ID,
		TraceID:         trace.ID,
		UserID:          trace.UserID,
		MatchScore:      1.0,
		MatchType:       domain.ConstellationMatchTypePrimary,
		MatchDimensions: []string{"new_theme"},
		MatchReason:     "没有达到强匹配阈值的已有星座，创建新星座作为主归属",
		Weight:          1.0,
		CreatedAt:       now,
	}); err != nil {
		return "", fmt.Errorf("add primary membership: %w", err)
	}
	return constellationID, nil
}

func (uc *StashTraceUseCase) attachStarToConstellation(ctx context.Context, constellationID string, star domain.Star, trace writingdomain.Trace, moments []writingdomain.Moment, traceProfile *domain.TraceProfile, traceVector *domain.TraceProfileVector, scored scoredConstellationCandidate, matchType string, weight float64) error {
	c, err := uc.constellations.FindByID(ctx, constellationID)
	if err != nil {
		return fmt.Errorf("find constellation: %w", err)
	}
	if !containsString(c.StarIDs, star.ID) {
		c.StarIDs = append(c.StarIDs, star.ID)
	}
	c.Topic = scored.candidate.Profile.Topic
	if strings.TrimSpace(c.Name) == "" {
		c.Name = scored.candidate.Profile.Topic
	}
	if strings.TrimSpace(c.ConstellationInsight) == "" {
		c.ConstellationInsight = scored.candidate.Profile.Summary
	}
	c.UpdatedAt = time.Now()
	if err := uc.constellations.Update(ctx, c); err != nil {
		return fmt.Errorf("update constellation: %w", err)
	}

	now := time.Now()
	updatedProfile := mergeConstellationProfile(scored.candidate.Profile, traceProfile, weight, float64(len(moments)), now)
	var updatedVector *domain.ConstellationProfileVector
	if traceVector != nil {
		updatedVector = &domain.ConstellationProfileVector{
			ConstellationID:   constellationID,
			UserID:            trace.UserID,
			Model:             scored.candidate.Vector.Model,
			Dim:               scored.candidate.Vector.Dim,
			ProfileEmbedding:  scored.candidate.Vector.ProfileEmbedding,
			CentroidEmbedding: weightedCentroid(scored.candidate.Vector.CentroidEmbedding, scored.candidate.Profile.TraceCount, traceVector.Embedding, weight),
			CreatedAt:         scored.candidate.Vector.CreatedAt,
			UpdatedAt:         now,
		}
	}
	if err := uc.constellationProf.Upsert(ctx, updatedProfile, updatedVector); err != nil {
		return fmt.Errorf("upsert constellation profile: %w", err)
	}
	if err := uc.constellationProf.AddMembership(ctx, domain.ConstellationMembership{
		ConstellationID: constellationID,
		StarID:          star.ID,
		TraceID:         trace.ID,
		UserID:          trace.UserID,
		MatchScore:      scored.score,
		MatchType:       matchType,
		MatchDimensions: scored.dimensions,
		MatchReason:     scored.reason,
		Weight:          weight,
		CreatedAt:       now,
	}); err != nil {
		return fmt.Errorf("add membership: %w", err)
	}
	return nil
}

func rankConstellationCandidates(traceProfile *domain.TraceProfile, traceVector *domain.TraceProfileVector, candidates []domain.ConstellationProfileCandidate) []scoredConstellationCandidate {
	if traceProfile == nil || traceVector == nil {
		return nil
	}
	scored := make([]scoredConstellationCandidate, 0, len(candidates))
	for _, candidate := range candidates {
		profileSimilarity := cosineSimilarity(traceVector.Embedding, candidate.Vector.ProfileEmbedding)
		centroidSimilarity := cosineSimilarity(traceVector.Embedding, candidate.Vector.CentroidEmbedding)
		keywordOverlap := listOverlapRatio(traceProfile.Keywords, candidate.Profile.Keywords)
		sceneOverlap := listOverlapRatio(traceProfile.Scenes, candidate.Profile.Scenes)
		emotionOverlap := listOverlapRatio(traceProfile.Emotions, candidate.Profile.Emotions)
		patternTagsOverlap := listOverlapRatio(traceProfile.PatternTags, candidate.Profile.PatternTags)
		score := scoreConstellationCandidate(candidate.Profile.TraceCount, profileSimilarity, centroidSimilarity, keywordOverlap, sceneOverlap, emotionOverlap, patternTagsOverlap)
		explainableMiddle := isExplainableMiddleCandidate(keywordOverlap, sceneOverlap, emotionOverlap, patternTagsOverlap, score)

		dimensions := matchDimensions(profileSimilarity, centroidSimilarity, keywordOverlap, sceneOverlap, emotionOverlap, patternTagsOverlap)
		scored = append(scored, scoredConstellationCandidate{
			candidate:          candidate,
			score:              score,
			profileSimilarity:  profileSimilarity,
			centroidSimilarity: centroidSimilarity,
			keywordOverlap:     keywordOverlap,
			sceneOverlap:       sceneOverlap,
			emotionOverlap:     emotionOverlap,
			patternTagsOverlap: patternTagsOverlap,
			matchedKeywords:    listIntersection(traceProfile.Keywords, candidate.Profile.Keywords),
			matchedScenes:      listIntersection(traceProfile.Scenes, candidate.Profile.Scenes),
			matchedEmotions:    listIntersection(traceProfile.Emotions, candidate.Profile.Emotions),
			matchedPatternTags: listIntersection(traceProfile.PatternTags, candidate.Profile.PatternTags),
			explainableMiddle:  explainableMiddle,
			dimensions:         dimensions,
			reason:             matchReason(dimensions, score),
		})
	}
	sort.SliceStable(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})
	return scored
}

func traceVectorDim(vector *domain.TraceProfileVector) int {
	if vector == nil {
		return 0
	}
	if vector.Dim > 0 {
		return vector.Dim
	}
	return len(vector.Embedding)
}

func constellationCandidateSummaries(candidates []domain.ConstellationProfileCandidate, limit int) []map[string]any {
	if limit > 0 && len(candidates) > limit {
		candidates = candidates[:limit]
	}
	result := make([]map[string]any, 0, len(candidates))
	for _, candidate := range candidates {
		result = append(result, map[string]any{
			"constellation_id": candidate.Profile.ConstellationID,
			"topic":            candidate.Profile.Topic,
			"trace_count":      candidate.Profile.TraceCount,
			"moment_count":     candidate.Profile.MomentCount,
			"keywords":         candidate.Profile.Keywords,
			"emotions":         candidate.Profile.Emotions,
			"scenes":           candidate.Profile.Scenes,
			"pattern_tags":     candidate.Profile.PatternTags,
			"vector_dim":       len(candidate.Vector.ProfileEmbedding),
			"centroid_dim":     len(candidate.Vector.CentroidEmbedding),
		})
	}
	return result
}

func scoredConstellationCandidateSummaries(candidates []scoredConstellationCandidate, limit int) []map[string]any {
	if limit > 0 && len(candidates) > limit {
		candidates = candidates[:limit]
	}
	result := make([]map[string]any, 0, len(candidates))
	for rank, candidate := range candidates {
		result = append(result, scoredConstellationCandidateSummary(candidate, rank+1))
	}
	return result
}

func scoredConstellationCandidateSummary(candidate scoredConstellationCandidate, rank int) map[string]any {
	return map[string]any{
		"rank":                 rank,
		"constellation_id":     candidate.candidate.Profile.ConstellationID,
		"topic":                candidate.candidate.Profile.Topic,
		"score":                candidate.score,
		"strong_threshold":     constellationStrongMatchThreshold,
		"strong_threshold_gap": candidate.score - constellationStrongMatchThreshold,
		"middle_threshold":     constellationMiddleMatchThreshold,
		"middle_threshold_gap": candidate.score - constellationMiddleMatchThreshold,
		"profile_similarity":   candidate.profileSimilarity,
		"centroid_similarity":  candidate.centroidSimilarity,
		"keyword_overlap":      candidate.keywordOverlap,
		"scene_overlap":        candidate.sceneOverlap,
		"emotion_overlap":      candidate.emotionOverlap,
		"pattern_tags_overlap": candidate.patternTagsOverlap,
		"matched_keywords":     candidate.matchedKeywords,
		"matched_scenes":       candidate.matchedScenes,
		"matched_emotions":     candidate.matchedEmotions,
		"matched_pattern_tags": candidate.matchedPatternTags,
		"explainable_middle":   candidate.explainableMiddle,
		"score_components":     constellationScoreComponents(candidate),
		"dimensions":           candidate.dimensions,
		"reason":               candidate.reason,
	}
}

func constellationScoreWeights() map[string]any {
	return map[string]any{
		"default": map[string]float64{
			"profile_similarity":   constellationProfileSimilarityWeight,
			"centroid_similarity":  constellationCentroidSimilarityWeight,
			"keyword_overlap":      constellationKeywordOverlapWeight,
			"scene_overlap":        constellationSceneOverlapWeight,
			"emotion_overlap":      constellationEmotionOverlapWeight,
			"pattern_tags_overlap": constellationPatternTagsOverlapWeight,
		},
		"single_trace": map[string]float64{
			"profile_similarity":   singleTraceProfileSimilarityWeight,
			"centroid_similarity":  0,
			"keyword_overlap":      singleTraceKeywordOverlapWeight,
			"scene_overlap":        singleTraceSceneOverlapWeight,
			"emotion_overlap":      singleTraceEmotionOverlapWeight,
			"pattern_tags_overlap": singleTracePatternTagsOverlapWeight,
		},
	}
}

func constellationProfileFromTraceProfile(constellationID string, userID string, profile *domain.TraceProfile, traceCount float64, momentCount float64, now time.Time) *domain.ConstellationProfile {
	result := &domain.ConstellationProfile{
		ConstellationID: constellationID,
		UserID:          userID,
		Topic:           profile.Topic,
		Summary:         profile.Summary,
		Keywords:        append([]string(nil), profile.Keywords...),
		Emotions:        append([]string(nil), profile.Emotions...),
		Scenes:          append([]string(nil), profile.Scenes...),
		CentralPattern:  profile.CentralPattern,
		PatternTags:     append([]string(nil), profile.PatternTags...),
		TraceCount:      traceCount,
		MomentCount:     momentCount,
		Status:          domain.ConstellationProfileStatusReady,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	result.ProfileText = buildConstellationProfileText(result)
	return result
}

func mergeConstellationProfile(existing domain.ConstellationProfile, incoming *domain.TraceProfile, weight float64, momentCount float64, now time.Time) *domain.ConstellationProfile {
	result := existing
	result.Keywords = mergeStringLists(result.Keywords, incoming.Keywords, 12)
	result.Emotions = mergeStringLists(result.Emotions, incoming.Emotions, 8)
	result.Scenes = mergeStringLists(result.Scenes, incoming.Scenes, 8)
	result.PatternTags = mergeStringLists(result.PatternTags, incoming.PatternTags, 8)
	if strings.TrimSpace(result.CentralPattern) == "" {
		result.CentralPattern = incoming.CentralPattern
	}
	if strings.TrimSpace(result.Summary) == "" {
		result.Summary = incoming.Summary
	}
	if strings.TrimSpace(result.Topic) == "" {
		result.Topic = incoming.Topic
	}
	result.TraceCount += weight
	result.MomentCount += momentCount * weight
	result.Status = domain.ConstellationProfileStatusReady
	result.LastError = ""
	result.UpdatedAt = now
	result.ProfileText = buildConstellationProfileText(&result)
	return &result
}

func buildConstellationProfileText(profile *domain.ConstellationProfile) string {
	var b strings.Builder
	if profile.Topic != "" {
		fmt.Fprintf(&b, "主题：%s\n", profile.Topic)
	}
	if profile.Summary != "" {
		fmt.Fprintf(&b, "摘要：%s\n", profile.Summary)
	}
	if len(profile.Keywords) > 0 {
		fmt.Fprintf(&b, "关键词：%s\n", strings.Join(profile.Keywords, "，"))
	}
	if len(profile.Emotions) > 0 {
		fmt.Fprintf(&b, "情绪：%s\n", strings.Join(profile.Emotions, "，"))
	}
	if len(profile.Scenes) > 0 {
		fmt.Fprintf(&b, "场景：%s\n", strings.Join(profile.Scenes, "，"))
	}
	if profile.CentralPattern != "" {
		fmt.Fprintf(&b, "核心模式：%s\n", profile.CentralPattern)
	}
	if len(profile.PatternTags) > 0 {
		fmt.Fprintf(&b, "模式标签：%s\n", strings.Join(profile.PatternTags, "，"))
	}
	return strings.TrimSpace(b.String())
}

func weightedCentroid(existing []float32, existingWeight float64, incoming []float32, incomingWeight float64) []float32 {
	if len(existing) == 0 {
		return append([]float32(nil), incoming...)
	}
	if len(incoming) == 0 || len(existing) != len(incoming) {
		return append([]float32(nil), existing...)
	}
	totalWeight := existingWeight + incomingWeight
	if totalWeight <= 0 {
		return append([]float32(nil), existing...)
	}
	result := make([]float32, len(existing))
	for i := range existing {
		result[i] = float32((float64(existing[i])*existingWeight + float64(incoming[i])*incomingWeight) / totalWeight)
	}
	return result
}

func cosineSimilarity(a []float32, b []float32) float64 {
	if len(a) == 0 || len(a) != len(b) {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		av := float64(a[i])
		bv := float64(b[i])
		dot += av * bv
		normA += av * av
		normB += bv * bv
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

func listOverlapRatio(a []string, b []string) float64 {
	aSet := normalizedSet(a)
	bSet := normalizedSet(b)
	if len(aSet) == 0 || len(bSet) == 0 {
		return 0
	}
	intersection := 0
	for value := range aSet {
		if _, ok := bSet[value]; ok {
			intersection++
		}
	}
	denominator := math.Min(float64(len(aSet)), float64(len(bSet)))
	if denominator == 0 {
		return 0
	}
	return float64(intersection) / denominator
}

func scoreConstellationCandidate(traceCount float64, profileSimilarity float64, centroidSimilarity float64, keywordOverlap float64, sceneOverlap float64, emotionOverlap float64, patternTagsOverlap float64) float64 {
	if traceCount <= 1 {
		return singleTraceProfileSimilarityWeight*profileSimilarity +
			singleTraceKeywordOverlapWeight*keywordOverlap +
			singleTraceSceneOverlapWeight*sceneOverlap +
			singleTraceEmotionOverlapWeight*emotionOverlap +
			singleTracePatternTagsOverlapWeight*patternTagsOverlap
	}
	return constellationProfileSimilarityWeight*profileSimilarity +
		constellationCentroidSimilarityWeight*centroidSimilarity +
		constellationKeywordOverlapWeight*keywordOverlap +
		constellationSceneOverlapWeight*sceneOverlap +
		constellationEmotionOverlapWeight*emotionOverlap +
		constellationPatternTagsOverlapWeight*patternTagsOverlap
}

func constellationScoreComponents(candidate scoredConstellationCandidate) map[string]float64 {
	if candidate.candidate.Profile.TraceCount <= 1 {
		return map[string]float64{
			"profile_similarity":   singleTraceProfileSimilarityWeight * candidate.profileSimilarity,
			"centroid_similarity":  0,
			"keyword_overlap":      singleTraceKeywordOverlapWeight * candidate.keywordOverlap,
			"scene_overlap":        singleTraceSceneOverlapWeight * candidate.sceneOverlap,
			"emotion_overlap":      singleTraceEmotionOverlapWeight * candidate.emotionOverlap,
			"pattern_tags_overlap": singleTracePatternTagsOverlapWeight * candidate.patternTagsOverlap,
		}
	}
	return map[string]float64{
		"profile_similarity":   constellationProfileSimilarityWeight * candidate.profileSimilarity,
		"centroid_similarity":  constellationCentroidSimilarityWeight * candidate.centroidSimilarity,
		"keyword_overlap":      constellationKeywordOverlapWeight * candidate.keywordOverlap,
		"scene_overlap":        constellationSceneOverlapWeight * candidate.sceneOverlap,
		"emotion_overlap":      constellationEmotionOverlapWeight * candidate.emotionOverlap,
		"pattern_tags_overlap": constellationPatternTagsOverlapWeight * candidate.patternTagsOverlap,
	}
}

func isPrimaryAttachCandidate(candidate scoredConstellationCandidate) bool {
	return candidate.score >= constellationMiddleMatchThreshold || candidate.explainableMiddle
}

func isExplainableMiddleCandidate(keywordOverlap float64, sceneOverlap float64, emotionOverlap float64, patternTagsOverlap float64, score float64) bool {
	if score < constellationExplainableThreshold || score >= constellationMiddleMatchThreshold {
		return false
	}
	positive := 0
	for _, value := range []float64{keywordOverlap, sceneOverlap, emotionOverlap, patternTagsOverlap} {
		if value > 0 {
			positive++
		}
	}
	return positive >= 3
}

func matchDimensions(profileSimilarity float64, centroidSimilarity float64, keywordOverlap float64, sceneOverlap float64, emotionOverlap float64, patternTagsOverlap float64) []string {
	var dimensions []string
	if profileSimilarity >= 0.72 || centroidSimilarity >= 0.72 {
		dimensions = append(dimensions, "profile")
	}
	if keywordOverlap > 0 {
		dimensions = append(dimensions, "keyword")
	}
	if sceneOverlap > 0 {
		dimensions = append(dimensions, "scene")
	}
	if emotionOverlap > 0 {
		dimensions = append(dimensions, "emotion")
	}
	if patternTagsOverlap > 0 {
		dimensions = append(dimensions, "pattern_tags")
	}
	if len(dimensions) == 0 {
		dimensions = append(dimensions, "profile")
	}
	return dimensions
}

func matchReason(dimensions []string, score float64) string {
	if len(dimensions) == 0 {
		return fmt.Sprintf("画像相似度 %.3f", score)
	}
	return fmt.Sprintf("基于 %s 的画像匹配，综合分 %.3f", strings.Join(dimensions, ","), score)
}

func hasDistinctDimensions(primary []string, secondary []string) bool {
	primarySet := normalizedSet(primary)
	for _, dimension := range secondary {
		if _, ok := primarySet[strings.TrimSpace(dimension)]; !ok {
			return true
		}
	}
	return false
}

func mergeStringLists(a []string, b []string, limit int) []string {
	result := make([]string, 0, len(a)+len(b))
	seen := map[string]struct{}{}
	for _, value := range append(append([]string{}, a...), b...) {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
		if limit > 0 && len(result) >= limit {
			break
		}
	}
	return result
}

func listIntersection(a []string, b []string) []string {
	bSet := normalizedSet(b)
	if len(bSet) == 0 {
		return nil
	}
	result := make([]string, 0)
	seen := map[string]struct{}{}
	for _, value := range a {
		normalized := strings.TrimSpace(value)
		if normalized == "" {
			continue
		}
		if _, ok := bSet[normalized]; !ok {
			continue
		}
		if _, ok := seen[normalized]; ok {
			continue
		}
		seen[normalized] = struct{}{}
		result = append(result, normalized)
	}
	return result
}

func normalizedSet(values []string) map[string]struct{} {
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		result[value] = struct{}{}
	}
	return result
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
