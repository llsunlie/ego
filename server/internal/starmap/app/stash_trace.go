package app

import (
	"context"
	"fmt"
	"hash/fnv"
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
	constellationSparseRecallTopK     = 10
	constellationHybridRRFK           = 60
	constellationStrongMatchThreshold = 0.68
	constellationMiddleMatchThreshold = 0.52
	constellationExplainableThreshold = 0.50
	maxSecondaryConstellationMatches  = 2
	borderlineCandidateLimit          = 3
	borderlineWeakThreshold           = 0.30
	borderlineProfileSimilarityFloor  = 0.38
	borderlineConfidenceThreshold     = 0.65
	borderlineSuggestedConfidence     = 0.55

	constellationProfileSimilarityWeight  = 0.42
	constellationCentroidSimilarityWeight = 0.22
	constellationKeywordOverlapWeight     = 0.14
	constellationSceneOverlapWeight       = 0.10
	constellationEmotionOverlapWeight     = 0.04
	constellationPatternTagsOverlapWeight = 0.08

	singleTraceProfileSimilarityWeight  = 0.56
	singleTraceKeywordOverlapWeight     = 0.16
	singleTraceSceneOverlapWeight       = 0.12
	singleTraceEmotionOverlapWeight     = 0.06
	singleTracePatternTagsOverlapWeight = 0.10
)

type StashTraceUseCase struct {
	traceReader       domain.TraceReader
	traceStasher      domain.TraceStasher
	stars             domain.StarRepository
	constellations    domain.ConstellationRepository
	assetGen          domain.ConstellationAssetGenerator
	profileGen        domain.TraceProfileGenerator
	borderlineJudge   domain.ConstellationBorderlineJudge
	profileRefiner    domain.ConstellationProfileRefiner
	profileRepo       domain.TraceProfileRepository
	constellationProf domain.ConstellationProfileRepository
	profileIndexer    domain.ConstellationProfileSearchIndexer
	sparseProfiles    domain.ConstellationProfileSparseCandidateReader
	sparseTopK        int
	hybridRRFK        int
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
	borderlineJudge domain.ConstellationBorderlineJudge,
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
		borderlineJudge:   borderlineJudge,
		profileRepo:       profileRepo,
		constellationProf: constellationProf,
		sparseTopK:        constellationSparseRecallTopK,
		hybridRRFK:        constellationHybridRRFK,
		ids:               ids,
	}
}

func (uc *StashTraceUseCase) UseConstellationSparseSearch(indexer domain.ConstellationProfileSearchIndexer, reader domain.ConstellationProfileSparseCandidateReader, sparseTopK int, hybridRRFK int) {
	uc.profileIndexer = indexer
	uc.sparseProfiles = reader
	if sparseTopK > 0 {
		uc.sparseTopK = sparseTopK
	}
	if hybridRRFK > 0 {
		uc.hybridRRFK = hybridRRFK
	}
}

func (uc *StashTraceUseCase) UseConstellationProfileRefiner(refiner domain.ConstellationProfileRefiner) {
	uc.profileRefiner = refiner
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

	profile, vector, err := uc.profileGen.Generate(ctx, trace, moments)
	if err != nil {
		logger.ErrorContext(ctx, "starmap profile clustering trace profile generation failed",
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
		"keywords", profile.Keywords,
		"emotions", profile.Emotions,
		"scenes", profile.Scenes,
		"entral_pattern", profile.CentralPattern,
		"pattern_tags", profile.PatternTags,
		"has_vector", vector != nil,
		"vector_dim", traceVectorDim(vector),
	)

	var ranked []scoredConstellationCandidate
	if vector != nil {
		candidates, err := uc.recallConstellationCandidates(ctx, trace, star, profile, vector)
		if err != nil {
			logger.ErrorContext(ctx, "starmap profile clustering candidate recall failed",
				"trace_id", trace.ID,
				"star_id", star.ID,
				"error", err,
				"recovery", "pending_message_queue",
			)
			return
		}
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
	var borderlineJudgement *domain.ConstellationBorderlineJudgement
	var llmSecondary []scoredConstellationCandidate
	if len(ranked) > 0 && isPrimaryAttachCandidate(ranked[0]) {
		primaryID = ranked[0].candidate.Profile.ConstellationID
		primaryDimensions = ranked[0].dimensions
		primaryScore = ranked[0].score
		primaryDecision = "attach_existing"
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
		borderlineJudgement = uc.runBorderlineJudgement(ctx, trace, star, profile, moments, ranked)
		if accepted, selected := uc.acceptBorderlineJudgement(ctx, trace, star, borderlineJudgement, ranked); accepted {
			primaryID = selected.candidate.Profile.ConstellationID
			primaryDimensions = selected.dimensions
			primaryScore = selected.score
			primaryDecision = "attach_existing_borderline"
			llmSecondary = uc.acceptBorderlineSecondaryJudgements(ctx, trace, star, borderlineJudgement, ranked, primaryID)
			logger.DebugContext(ctx, "starmap profile clustering borderline primary decision",
				"trace_id", trace.ID,
				"star_id", star.ID,
				"decision", primaryDecision,
				"constellation_id", primaryID,
				"score", primaryScore,
				"llm_confidence", selectionConfidence(borderlineJudgement.Primary, borderlineJudgement.Confidence),
				"llm_shared_theme", selectionSharedTheme(borderlineJudgement.Primary, borderlineJudgement.SharedSituation),
				"llm_match_dimensions", selected.dimensions,
				"llm_secondary_count", len(llmSecondary),
				"reason", selected.reason,
				"top_candidate", scoredConstellationCandidateSummary(selected, 1),
			)
			if err := uc.attachStarToConstellation(ctx, primaryID, star, trace, moments, profile, vector, selected, domain.ConstellationMatchTypePrimary, 1.0); err != nil {
				logger.ErrorContext(ctx, "starmap profile clustering attach borderline primary failed",
					"trace_id", trace.ID,
					"star_id", star.ID,
					"constellation_id", primaryID,
					"error", err,
					"recovery", "pending_message_queue",
				)
				return
			}
		} else {
			primaryID, err = uc.createConstellationFromTraceProfile(ctx, star, trace, moments, profile, vector, borderlineJudgement)
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
			llmSecondary = uc.acceptBorderlineSecondaryJudgements(ctx, trace, star, borderlineJudgement, ranked, primaryID)
		}
	}

	secondaryCount := 0
	attachedSecondary := map[string]struct{}{}
	for _, candidate := range llmSecondary {
		if secondaryCount >= maxSecondaryConstellationMatches {
			break
		}
		if candidate.candidate.Profile.ConstellationID == primaryID {
			continue
		}
		logger.DebugContext(ctx, "starmap profile clustering secondary decision",
			"trace_id", trace.ID,
			"star_id", star.ID,
			"decision", "attach_secondary_llm",
			"constellation_id", candidate.candidate.Profile.ConstellationID,
			"score", candidate.score,
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
		attachedSecondary[candidate.candidate.Profile.ConstellationID] = struct{}{}
		secondaryCount++
	}
	for _, candidate := range ranked {
		if primaryDecision == "create_new" {
			logger.DebugContext(ctx, "starmap profile clustering secondary candidate skipped",
				"trace_id", trace.ID,
				"star_id", star.ID,
				"constellation_id", candidate.candidate.Profile.ConstellationID,
				"score", candidate.score,
				"reason", "primary_created_new",
			)
			continue
		}
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
		if _, ok := attachedSecondary[candidate.candidate.Profile.ConstellationID]; ok {
			logger.DebugContext(ctx, "starmap profile clustering secondary candidate skipped",
				"trace_id", trace.ID,
				"star_id", star.ID,
				"constellation_id", candidate.candidate.Profile.ConstellationID,
				"score", candidate.score,
				"reason", "already_attached_by_llm",
			)
			continue
		}
		if candidate.score < borderlineWeakThreshold && !candidate.explainableMiddle {
			logger.DebugContext(ctx, "starmap profile clustering secondary candidate skipped",
				"trace_id", trace.ID,
				"star_id", star.ID,
				"constellation_id", candidate.candidate.Profile.ConstellationID,
				"score", candidate.score,
				"threshold", borderlineWeakThreshold,
				"explainable_threshold", constellationExplainableThreshold,
				"explainable_middle", candidate.explainableMiddle,
				"reason", "below_secondary_threshold",
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

type constellationDenseRecallResult struct {
	candidates []domain.ConstellationProfileCandidate
	err        error
}

type constellationSparseRecallResult struct {
	candidates []domain.ConstellationProfileCandidate
	sparse     []domain.ConstellationProfileSparseCandidate
	err        error
}

func (uc *StashTraceUseCase) recallConstellationCandidates(ctx context.Context, trace writingdomain.Trace, star domain.Star, profile *domain.TraceProfile, vector *domain.TraceProfileVector) ([]domain.ConstellationProfileCandidate, error) {
	logger := logging.FromContext(ctx)
	denseCh := make(chan constellationDenseRecallResult, 1)
	sparseCh := make(chan constellationSparseRecallResult, 1)

	go func() {
		candidates, err := uc.constellationProf.FindCandidates(ctx, trace.UserID, vector.Embedding, constellationCandidateLimit)
		denseCh <- constellationDenseRecallResult{candidates: candidates, err: err}
	}()
	go func() {
		if uc.sparseProfiles == nil || uc.sparseTopK <= 0 {
			sparseCh <- constellationSparseRecallResult{}
			return
		}
		sparse, err := uc.sparseProfiles.SearchCandidates(ctx, *profile, uc.sparseTopK)
		if err != nil {
			sparseCh <- constellationSparseRecallResult{err: err}
			return
		}
		ids := constellationSparseCandidateIDs(sparse)
		candidates, err := uc.constellationProf.FindCandidatesByIDs(ctx, trace.UserID, ids)
		sparseCh <- constellationSparseRecallResult{candidates: candidates, sparse: sparse, err: err}
	}()

	denseResult := <-denseCh
	if denseResult.err != nil {
		return nil, denseResult.err
	}
	sparseResult := <-sparseCh
	if sparseResult.err != nil {
		logger.WarnContext(ctx, "starmap constellation sparse recall failed",
			"trace_id", trace.ID,
			"star_id", star.ID,
			"error", sparseResult.err,
		)
		sparseResult = constellationSparseRecallResult{}
	}

	fused := mergeConstellationCandidatesRRF(denseResult.candidates, sparseResult.candidates, uc.hybridRRFK, constellationCandidateLimit)
	logger.DebugContext(ctx, "starmap constellation recall candidates",
		"trace_id", trace.ID,
		"star_id", star.ID,
		"current_topic", profile.Topic,
		"dense_count", len(denseResult.candidates),
		"sparse_count", len(sparseResult.candidates),
		"fused_count", len(fused),
		"dense_top_k", constellationCandidateLimit,
		"sparse_top_k", uc.sparseTopK,
		"rrf_k", uc.hybridRRFK,
		"dense_candidates", constellationCandidateSummaries(denseResult.candidates, constellationCandidateLimit),
		"sparse_candidates", sparseConstellationCandidateSummaries(sparseResult.sparse, sparseResult.candidates, uc.sparseTopK),
		"fused_candidates", constellationCandidateSummaries(fused, constellationCandidateLimit),
	)
	return fused, nil
}

func constellationSparseCandidateIDs(candidates []domain.ConstellationProfileSparseCandidate) []string {
	ids := make([]string, 0, len(candidates))
	seen := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		id := strings.TrimSpace(candidate.ConstellationID)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		ids = append(ids, id)
	}
	return ids
}

type rankedConstellationCandidate struct {
	candidate domain.ConstellationProfileCandidate
	score     float64
}

func mergeConstellationCandidatesRRF(dense []domain.ConstellationProfileCandidate, sparse []domain.ConstellationProfileCandidate, rrfK int, limit int) []domain.ConstellationProfileCandidate {
	if rrfK <= 0 {
		rrfK = constellationHybridRRFK
	}
	if limit <= 0 {
		limit = maxInt(len(dense), len(sparse))
	}
	byID := make(map[string]rankedConstellationCandidate)
	add := func(candidate domain.ConstellationProfileCandidate, rank int) {
		id := candidate.Profile.ConstellationID
		if id == "" {
			return
		}
		current := byID[id]
		if current.candidate.Profile.ConstellationID == "" {
			current.candidate = candidate
		}
		current.score += 1 / float64(rrfK+rank)
		byID[id] = current
	}
	for i, candidate := range dense {
		add(candidate, i+1)
	}
	for i, candidate := range sparse {
		add(candidate, i+1)
	}
	ranked := make([]rankedConstellationCandidate, 0, len(byID))
	for _, candidate := range byID {
		ranked = append(ranked, candidate)
	}
	sort.SliceStable(ranked, func(i, j int) bool {
		if ranked[i].score == ranked[j].score {
			return ranked[i].candidate.Profile.ConstellationID < ranked[j].candidate.Profile.ConstellationID
		}
		return ranked[i].score > ranked[j].score
	})
	if len(ranked) > limit {
		ranked = ranked[:limit]
	}
	result := make([]domain.ConstellationProfileCandidate, 0, len(ranked))
	for _, candidate := range ranked {
		result = append(result, candidate.candidate)
	}
	return result
}

func (uc *StashTraceUseCase) runBorderlineJudgement(ctx context.Context, trace writingdomain.Trace, star domain.Star, profile *domain.TraceProfile, moments []writingdomain.Moment, ranked []scoredConstellationCandidate) *domain.ConstellationBorderlineJudgement {
	logger := logging.FromContext(ctx)
	if uc.borderlineJudge == nil {
		return nil
	}
	candidates := borderlineCandidates(ranked)
	if len(candidates) == 0 {
		return nil
	}
	logger.DebugContext(ctx, "starmap borderline judgement started",
		"trace_id", trace.ID,
		"star_id", star.ID,
		"candidate_count", len(candidates),
		"top_score", ranked[0].score,
		"candidates", borderlineCandidateLogSummaries(candidates),
	)
	judgement, err := uc.borderlineJudge.Judge(ctx, domain.ConstellationBorderlineJudgeInput{
		TraceProfile:         *profile,
		RepresentativeMoment: representativeMomentContent(profile.RepresentativeMomentID, moments),
		Candidates:           candidates,
	})
	if err != nil {
		logger.WarnContext(ctx, "starmap borderline judgement failed",
			"trace_id", trace.ID,
			"star_id", star.ID,
			"error", err,
			"fallback_reason", "judge_error",
		)
		return nil
	}
	logger.DebugContext(ctx, "starmap borderline judgement completed",
		"trace_id", trace.ID,
		"star_id", star.ID,
		"decision", judgement.Decision,
		"constellation_id", judgement.ConstellationID,
		"theme_code", judgement.ThemeCode,
		"confidence", judgement.Confidence,
		"shared_situation", judgement.SharedSituation,
		"match_dimensions", judgement.MatchDimensions,
		"reason", judgement.Reason,
		"primary", borderlineSelectionLogSummary(judgement.Primary),
		"secondary", borderlineSelectionLogSummaries(judgement.Secondary),
		"suggested_theme_code", judgement.SuggestedThemeCode,
		"suggested_theme_label", judgement.SuggestedThemeLabel,
	)
	return judgement
}

func (uc *StashTraceUseCase) acceptBorderlineJudgement(ctx context.Context, trace writingdomain.Trace, star domain.Star, judgement *domain.ConstellationBorderlineJudgement, ranked []scoredConstellationCandidate) (bool, scoredConstellationCandidate) {
	logger := logging.FromContext(ctx)
	if judgement == nil {
		return false, scoredConstellationCandidate{}
	}
	if judgement.Decision != domain.ConstellationBorderlineDecisionUseExisting {
		logger.DebugContext(ctx, "starmap borderline judgement rejected",
			"trace_id", trace.ID,
			"star_id", star.ID,
			"reason", "decision_not_use_existing",
			"decision", judgement.Decision,
			"confidence", judgement.Confidence,
		)
		return false, scoredConstellationCandidate{}
	}
	selection := judgement.Primary
	if selection == nil && judgement.ConstellationID != "" {
		selection = &domain.ConstellationBorderlineSelection{
			ConstellationID: strings.TrimSpace(judgement.ConstellationID),
			ThemeCode:       strings.TrimSpace(judgement.ThemeCode),
			Confidence:      judgement.Confidence,
			SharedTheme:     strings.TrimSpace(judgement.SharedSituation),
			MatchDimensions: append([]string(nil), judgement.MatchDimensions...),
			Reason:          strings.TrimSpace(judgement.Reason),
		}
	}
	if selection == nil {
		logger.DebugContext(ctx, "starmap borderline judgement rejected",
			"trace_id", trace.ID,
			"star_id", star.ID,
			"reason", "missing_primary_selection",
		)
		return false, scoredConstellationCandidate{}
	}
	return uc.acceptBorderlineSelection(ctx, trace, star, *selection, ranked, borderlineConfidenceThreshold, true)
}

func (uc *StashTraceUseCase) acceptBorderlineSecondaryJudgements(ctx context.Context, trace writingdomain.Trace, star domain.Star, judgement *domain.ConstellationBorderlineJudgement, ranked []scoredConstellationCandidate, primaryID string) []scoredConstellationCandidate {
	if judgement == nil || len(judgement.Secondary) == 0 {
		return nil
	}
	result := make([]scoredConstellationCandidate, 0, len(judgement.Secondary))
	seen := map[string]struct{}{primaryID: {}}
	for _, selection := range judgement.Secondary {
		id := strings.TrimSpace(selection.ConstellationID)
		if id == "" {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		accepted, scored := uc.acceptBorderlineSelection(ctx, trace, star, selection, ranked, 0.60, false)
		if !accepted {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, scored)
		if len(result) >= maxSecondaryConstellationMatches {
			break
		}
	}
	return result
}

func (uc *StashTraceUseCase) acceptBorderlineSelection(ctx context.Context, trace writingdomain.Trace, star domain.Star, selection domain.ConstellationBorderlineSelection, ranked []scoredConstellationCandidate, confidenceThreshold float64, primary bool) (bool, scoredConstellationCandidate) {
	logger := logging.FromContext(ctx)
	selection.ConstellationID = strings.TrimSpace(selection.ConstellationID)
	selection.ThemeCode = strings.TrimSpace(selection.ThemeCode)
	selection.SharedTheme = strings.TrimSpace(selection.SharedTheme)
	selection.Reason = strings.TrimSpace(selection.Reason)
	if selection.Confidence < confidenceThreshold {
		logger.DebugContext(ctx, "starmap borderline judgement rejected",
			"trace_id", trace.ID,
			"star_id", star.ID,
			"reason", "low_confidence",
			"constellation_id", selection.ConstellationID,
			"confidence", selection.Confidence,
			"threshold", confidenceThreshold,
			"primary", primary,
		)
		return false, scoredConstellationCandidate{}
	}
	if selection.SharedTheme == "" {
		logger.DebugContext(ctx, "starmap borderline judgement rejected",
			"trace_id", trace.ID,
			"star_id", star.ID,
			"reason", "missing_shared_theme",
			"constellation_id", selection.ConstellationID,
			"primary", primary,
		)
		return false, scoredConstellationCandidate{}
	}
	if !hasAllowedBorderlineDimension(selection.MatchDimensions) {
		logger.DebugContext(ctx, "starmap borderline judgement rejected",
			"trace_id", trace.ID,
			"star_id", star.ID,
			"reason", "missing_allowed_dimension",
			"constellation_id", selection.ConstellationID,
			"match_dimensions", selection.MatchDimensions,
			"primary", primary,
		)
		return false, scoredConstellationCandidate{}
	}
	for _, candidate := range ranked {
		if candidate.candidate.Profile.ConstellationID != selection.ConstellationID {
			continue
		}
		if !isBorderlineCandidate(candidate) && !(selection.Confidence >= 0.75 && !primary) {
			logger.DebugContext(ctx, "starmap borderline judgement rejected",
				"trace_id", trace.ID,
				"star_id", star.ID,
				"reason", "candidate_not_in_borderline_range",
				"constellation_id", selection.ConstellationID,
				"score", candidate.score,
				"confidence", selection.Confidence,
				"primary", primary,
			)
			return false, scoredConstellationCandidate{}
		}
		if strings.TrimSpace(candidate.candidate.Profile.ThemeCode) != "" && candidate.candidate.Profile.ThemeCode != selection.ThemeCode {
			logger.DebugContext(ctx, "starmap borderline judgement rejected",
				"trace_id", trace.ID,
				"star_id", star.ID,
				"reason", "theme_code_mismatch",
				"constellation_id", selection.ConstellationID,
				"candidate_theme_code", candidate.candidate.Profile.ThemeCode,
				"judgement_theme_code", selection.ThemeCode,
				"primary", primary,
			)
			return false, scoredConstellationCandidate{}
		}
		selected := candidate
		if len(selection.MatchDimensions) > 0 {
			selected.dimensions = append([]string(nil), selection.MatchDimensions...)
		}
		if selection.Reason != "" {
			selected.reason = selection.Reason
		}
		return true, selected
	}
	logger.DebugContext(ctx, "starmap borderline judgement rejected",
		"trace_id", trace.ID,
		"star_id", star.ID,
		"reason", "constellation_id_not_in_candidates",
		"constellation_id", selection.ConstellationID,
		"primary", primary,
	)
	return false, scoredConstellationCandidate{}
}

func (uc *StashTraceUseCase) createConstellationFromTraceProfile(ctx context.Context, star domain.Star, trace writingdomain.Trace, moments []writingdomain.Moment, profile *domain.TraceProfile, vector *domain.TraceProfileVector, suggested *domain.ConstellationBorderlineJudgement) (string, error) {
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

	cProfile := constellationProfileFromTraceProfile(constellationID, trace.UserID, profile, moments, 1.0, float64(len(moments)), now)
	applySuggestedThemeCodebook(cProfile, suggested)
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
	if err := uc.upsertConstellationProfile(ctx, cProfile, cVector); err != nil {
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
	updatedProfile, updatedVector = uc.refineConstellationProfileIfNeeded(ctx, scored.candidate.Profile, updatedProfile, traceProfile, moments, updatedVector)
	if err := uc.upsertConstellationProfile(ctx, updatedProfile, updatedVector); err != nil {
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

func (uc *StashTraceUseCase) refineConstellationProfileIfNeeded(ctx context.Context, existing domain.ConstellationProfile, merged *domain.ConstellationProfile, incoming *domain.TraceProfile, moments []writingdomain.Moment, vector *domain.ConstellationProfileVector) (*domain.ConstellationProfile, *domain.ConstellationProfileVector) {
	logger := logging.FromContext(ctx)
	if merged == nil || incoming == nil {
		logger.DebugContext(ctx, "starmap constellation profile refinement skipped",
			"reason", "missing_profile",
		)
		return merged, vector
	}
	trigger, ok := constellationProfileRefinementTrigger(existing.TraceCount, merged.TraceCount)
	if !ok {
		logger.DebugContext(ctx, "starmap constellation profile refinement skipped",
			"constellation_id", merged.ConstellationID,
			"old_trace_count", existing.TraceCount,
			"new_trace_count", merged.TraceCount,
			"reason", "trigger_not_reached",
		)
		return merged, vector
	}
	if uc.profileRefiner == nil {
		logger.DebugContext(ctx, "starmap constellation profile refinement skipped",
			"constellation_id", merged.ConstellationID,
			"old_trace_count", existing.TraceCount,
			"new_trace_count", merged.TraceCount,
			"trigger", trigger,
			"reason", "refiner_not_configured",
		)
		return merged, vector
	}
	logger.DebugContext(ctx, "starmap constellation profile refinement started",
		"constellation_id", merged.ConstellationID,
		"old_trace_count", existing.TraceCount,
		"new_trace_count", merged.TraceCount,
		"trigger", trigger,
		"topic", merged.Topic,
	)
	refinement, err := uc.profileRefiner.Refine(ctx, domain.ConstellationProfileRefineInput{
		Existing:             existing,
		RuleMerged:           *merged,
		IncomingTraceProfile: *incoming,
		RepresentativeMoment: representativeMomentContent(incoming.RepresentativeMomentID, moments),
		Trigger:              trigger,
	})
	if err != nil {
		logger.WarnContext(ctx, "starmap constellation profile refinement failed",
			"constellation_id", merged.ConstellationID,
			"old_trace_count", existing.TraceCount,
			"new_trace_count", merged.TraceCount,
			"trigger", trigger,
			"fallback", "rule_merged_profile",
			"error", err,
		)
		return merged, vector
	}
	refined := refinement.Profile
	refined.ConstellationID = merged.ConstellationID
	refined.UserID = merged.UserID
	refined.TraceCount = merged.TraceCount
	refined.MomentCount = merged.MomentCount
	refined.Status = domain.ConstellationProfileStatusReady
	refined.LastError = ""
	refined.CreatedAt = merged.CreatedAt
	refined.UpdatedAt = merged.UpdatedAt
	if strings.TrimSpace(refined.ThemeCode) == "" {
		refined.ThemeCode = merged.ThemeCode
	}
	if strings.TrimSpace(refined.ProfileText) == "" {
		refined.ProfileText = buildConstellationProfileText(&refined)
	}
	if vector != nil && len(refinement.ProfileEmbedding) > 0 {
		vector.ProfileEmbedding = refinement.ProfileEmbedding
		vector.Model = refinement.Model
		vector.Dim = refinement.Dim
	}
	logger.DebugContext(ctx, "starmap constellation profile refinement completed",
		"constellation_id", refined.ConstellationID,
		"trigger", trigger,
		"topic", refined.Topic,
		"keywords", refined.Keywords,
		"scenes", refined.Scenes,
		"pattern_tags", refined.PatternTags,
		"theme_label", refined.ThemeLabel,
		"profile_embedding_dim", len(refinement.ProfileEmbedding),
	)
	return &refined, vector
}

func constellationProfileRefinementTrigger(oldTraceCount float64, newTraceCount float64) (int, bool) {
	oldFloor := int(math.Floor(oldTraceCount))
	newFloor := int(math.Floor(newTraceCount))
	if newFloor <= oldFloor {
		return 0, false
	}
	for _, trigger := range []int{3, 5, 8, 13} {
		if oldFloor < trigger && newFloor >= trigger {
			return trigger, true
		}
	}
	if newFloor <= 13 {
		return 0, false
	}
	oldBucket := maxInt(0, (oldFloor-13)/8)
	newBucket := maxInt(0, (newFloor-13)/8)
	if newBucket > oldBucket {
		return 13 + newBucket*8, true
	}
	return 0, false
}

func (uc *StashTraceUseCase) upsertConstellationProfile(ctx context.Context, profile *domain.ConstellationProfile, vector *domain.ConstellationProfileVector) error {
	if err := uc.constellationProf.Upsert(ctx, profile, vector); err != nil {
		return err
	}
	if uc.profileIndexer != nil && profile != nil {
		if err := uc.profileIndexer.IndexProfile(ctx, *profile); err != nil {
			logging.FromContext(ctx).WarnContext(ctx, "starmap constellation profile sparse index failed",
				"constellation_id", profile.ConstellationID,
				"user_id", profile.UserID,
				"topic", profile.Topic,
				"error", err,
			)
		}
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
			"theme_code":       candidate.Profile.ThemeCode,
			"theme_label":      candidate.Profile.ThemeLabel,
			"vector_dim":       len(candidate.Vector.ProfileEmbedding),
			"centroid_dim":     len(candidate.Vector.CentroidEmbedding),
		})
	}
	return result
}

func sparseConstellationCandidateSummaries(sparse []domain.ConstellationProfileSparseCandidate, candidates []domain.ConstellationProfileCandidate, limit int) []map[string]any {
	byID := make(map[string]domain.ConstellationProfileCandidate, len(candidates))
	for _, candidate := range candidates {
		byID[candidate.Profile.ConstellationID] = candidate
	}
	if limit > 0 && len(sparse) > limit {
		sparse = sparse[:limit]
	}
	result := make([]map[string]any, 0, len(sparse))
	for rank, candidate := range sparse {
		summary := map[string]any{
			"rank":             rank + 1,
			"constellation_id": candidate.ConstellationID,
			"score":            candidate.Score,
			"matched_fields":   candidate.MatchedFields,
			"preview":          candidate.Preview,
		}
		if loaded, ok := byID[candidate.ConstellationID]; ok {
			summary["topic"] = loaded.Profile.Topic
			summary["theme_code"] = loaded.Profile.ThemeCode
			summary["theme_label"] = loaded.Profile.ThemeLabel
		}
		result = append(result, summary)
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
		"theme_code":           candidate.candidate.Profile.ThemeCode,
		"theme_label":          candidate.candidate.Profile.ThemeLabel,
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

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func borderlineSelectionLogSummary(selection *domain.ConstellationBorderlineSelection) map[string]any {
	if selection == nil {
		return nil
	}
	return map[string]any{
		"constellation_id": selection.ConstellationID,
		"theme_code":       selection.ThemeCode,
		"confidence":       selection.Confidence,
		"shared_theme":     selection.SharedTheme,
		"match_dimensions": selection.MatchDimensions,
		"reason":           selection.Reason,
	}
}

func borderlineSelectionLogSummaries(selections []domain.ConstellationBorderlineSelection) []map[string]any {
	result := make([]map[string]any, 0, len(selections))
	for i := range selections {
		result = append(result, borderlineSelectionLogSummary(&selections[i]))
	}
	return result
}

func selectionConfidence(selection *domain.ConstellationBorderlineSelection, fallback float64) float64 {
	if selection == nil {
		return fallback
	}
	return selection.Confidence
}

func selectionSharedTheme(selection *domain.ConstellationBorderlineSelection, fallback string) string {
	if selection == nil {
		return fallback
	}
	return selection.SharedTheme
}

func borderlineCandidates(ranked []scoredConstellationCandidate) []domain.ConstellationBorderlineCandidate {
	if len(ranked) == 0 {
		return nil
	}
	limit := borderlineCandidateLimit
	if len(ranked) < limit {
		limit = len(ranked)
	}
	result := make([]domain.ConstellationBorderlineCandidate, 0, limit)
	for i := 0; i < limit; i++ {
		candidate := ranked[i]
		if !isBorderlineCandidate(candidate) {
			continue
		}
		profile := candidate.candidate.Profile
		result = append(result, domain.ConstellationBorderlineCandidate{
			ConstellationID:    profile.ConstellationID,
			Topic:              profile.Topic,
			Summary:            profile.Summary,
			Keywords:           append([]string(nil), profile.Keywords...),
			Emotions:           append([]string(nil), profile.Emotions...),
			Scenes:             append([]string(nil), profile.Scenes...),
			CentralPattern:     profile.CentralPattern,
			PatternTags:        append([]string(nil), profile.PatternTags...),
			ThemeCode:          profile.ThemeCode,
			ThemeLabel:         profile.ThemeLabel,
			ThemeDescription:   profile.ThemeDescription,
			ThemeExamples:      append([]string(nil), profile.ThemeExamples...),
			Score:              candidate.score,
			ProfileSimilarity:  candidate.profileSimilarity,
			CentroidSimilarity: candidate.centroidSimilarity,
			KeywordOverlap:     candidate.keywordOverlap,
			SceneOverlap:       candidate.sceneOverlap,
			EmotionOverlap:     candidate.emotionOverlap,
			PatternTagsOverlap: candidate.patternTagsOverlap,
			MatchedKeywords:    append([]string(nil), candidate.matchedKeywords...),
			MatchedScenes:      append([]string(nil), candidate.matchedScenes...),
			MatchedEmotions:    append([]string(nil), candidate.matchedEmotions...),
			MatchedPatternTags: append([]string(nil), candidate.matchedPatternTags...),
			Dimensions:         append([]string(nil), candidate.dimensions...),
			Reason:             candidate.reason,
		})
	}
	return result
}

func borderlineCandidateLogSummaries(candidates []domain.ConstellationBorderlineCandidate) []map[string]any {
	result := make([]map[string]any, 0, len(candidates))
	for _, candidate := range candidates {
		result = append(result, map[string]any{
			"constellation_id":     candidate.ConstellationID,
			"topic":                candidate.Topic,
			"theme_code":           candidate.ThemeCode,
			"theme_label":          candidate.ThemeLabel,
			"score":                candidate.Score,
			"profile_similarity":   candidate.ProfileSimilarity,
			"keyword_overlap":      candidate.KeywordOverlap,
			"scene_overlap":        candidate.SceneOverlap,
			"emotion_overlap":      candidate.EmotionOverlap,
			"pattern_tags_overlap": candidate.PatternTagsOverlap,
			"matched_keywords":     candidate.MatchedKeywords,
			"matched_scenes":       candidate.MatchedScenes,
			"matched_emotions":     candidate.MatchedEmotions,
			"matched_pattern_tags": candidate.MatchedPatternTags,
		})
	}
	return result
}

func isBorderlineCandidate(candidate scoredConstellationCandidate) bool {
	if candidate.score >= constellationStrongMatchThreshold {
		return false
	}
	if candidate.score < borderlineWeakThreshold {
		return false
	}
	return candidate.profileSimilarity >= borderlineProfileSimilarityFloor ||
		candidate.keywordOverlap > 0 ||
		candidate.sceneOverlap > 0 ||
		candidate.emotionOverlap > 0 ||
		candidate.patternTagsOverlap > 0
}

func hasAllowedBorderlineDimension(values []string) bool {
	allowed := map[string]struct{}{
		"situation":     {},
		"self_pattern":  {},
		"relationship":  {},
		"identity":      {},
		"need_conflict": {},
	}
	for _, value := range values {
		if _, ok := allowed[strings.TrimSpace(value)]; ok {
			return true
		}
	}
	return false
}

func representativeMomentContent(id string, moments []writingdomain.Moment) string {
	id = strings.TrimSpace(id)
	for _, moment := range moments {
		if id != "" && moment.ID != id {
			continue
		}
		return truncateText(moment.Content, 160)
	}
	return ""
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

func constellationProfileFromTraceProfile(constellationID string, userID string, profile *domain.TraceProfile, moments []writingdomain.Moment, traceCount float64, momentCount float64, now time.Time) *domain.ConstellationProfile {
	themeCode, themeLabel, themeDescription, themeExamples := fallbackThemeCodebook(profile, moments)
	result := &domain.ConstellationProfile{
		ConstellationID:  constellationID,
		UserID:           userID,
		Topic:            profile.Topic,
		Summary:          profile.Summary,
		Keywords:         append([]string(nil), profile.Keywords...),
		Emotions:         append([]string(nil), profile.Emotions...),
		Scenes:           append([]string(nil), profile.Scenes...),
		CentralPattern:   profile.CentralPattern,
		PatternTags:      append([]string(nil), profile.PatternTags...),
		ThemeCode:        themeCode,
		ThemeLabel:       themeLabel,
		ThemeDescription: themeDescription,
		ThemeExamples:    themeExamples,
		TraceCount:       traceCount,
		MomentCount:      momentCount,
		Status:           domain.ConstellationProfileStatusReady,
		CreatedAt:        now,
		UpdatedAt:        now,
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
	if strings.TrimSpace(result.ThemeCode) == "" {
		themeCode, themeLabel, themeDescription, _ := fallbackThemeCodebook(incoming, nil)
		result.ThemeCode = themeCode
		result.ThemeLabel = themeLabel
		result.ThemeDescription = themeDescription
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
	if profile.ThemeCode != "" {
		fmt.Fprintf(&b, "主题码：%s\n", profile.ThemeCode)
	}
	if profile.ThemeLabel != "" {
		fmt.Fprintf(&b, "主题标签：%s\n", profile.ThemeLabel)
	}
	if profile.ThemeDescription != "" {
		fmt.Fprintf(&b, "主题边界：%s\n", profile.ThemeDescription)
	}
	if len(profile.ThemeExamples) > 0 {
		fmt.Fprintf(&b, "代表例子：%s\n", strings.Join(profile.ThemeExamples, "；"))
	}
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

func fallbackThemeCodebook(profile *domain.TraceProfile, moments []writingdomain.Moment) (string, string, string, []string) {
	label := firstNonEmpty(profile.Topic, "未命名主题")
	description := firstNonEmpty(profile.Summary, profile.CentralPattern, profile.Topic, "暂时没有足够信息定义主题边界。")
	code := fallbackThemeCode(firstNonEmpty(profile.CentralPattern, profile.Topic, profile.Summary))
	examples := representativeThemeExamples(profile, moments)
	return code, truncateText(label, 32), truncateText(description, 120), examples
}

func applySuggestedThemeCodebook(profile *domain.ConstellationProfile, judgement *domain.ConstellationBorderlineJudgement) {
	if profile == nil || judgement == nil {
		return
	}
	if judgement.Decision != domain.ConstellationBorderlineDecisionSuggestNew || judgement.Confidence < borderlineSuggestedConfidence {
		return
	}
	code := fallbackThemeCode(judgement.SuggestedThemeCode)
	label := strings.TrimSpace(judgement.SuggestedThemeLabel)
	description := strings.TrimSpace(judgement.SuggestedThemeDescription)
	if code == "" || label == "" || description == "" {
		return
	}
	profile.ThemeCode = code
	profile.ThemeLabel = truncateText(label, 32)
	profile.ThemeDescription = truncateText(description, 120)
	profile.ProfileText = buildConstellationProfileText(profile)
}

func fallbackThemeCode(seed string) string {
	seed = strings.TrimSpace(strings.ToLower(seed))
	var b strings.Builder
	lastUnderscore := false
	for _, r := range seed {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
			lastUnderscore = false
		case r >= '0' && r <= '9':
			b.WriteRune(r)
			lastUnderscore = false
		case r == '_' || r == '-' || r == ' ' || r == '\t' || r == '\n':
			if b.Len() > 0 && !lastUnderscore {
				b.WriteByte('_')
				lastUnderscore = true
			}
		}
		if b.Len() >= 40 {
			break
		}
	}
	code := strings.Trim(b.String(), "_")
	if code != "" {
		return code
	}
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(seed))
	return fmt.Sprintf("theme_%08x", hash.Sum32())
}

func representativeThemeExamples(profile *domain.TraceProfile, moments []writingdomain.Moment) []string {
	if len(moments) == 0 {
		return nil
	}
	result := make([]string, 0, 3)
	seen := map[string]struct{}{}
	add := func(content string) {
		content = truncateText(content, 80)
		if content == "" {
			return
		}
		if _, ok := seen[content]; ok {
			return
		}
		seen[content] = struct{}{}
		result = append(result, content)
	}
	if profile != nil && strings.TrimSpace(profile.RepresentativeMomentID) != "" {
		add(representativeMomentContent(profile.RepresentativeMomentID, moments))
	}
	for _, moment := range moments {
		if len(result) >= 3 {
			break
		}
		add(moment.Content)
	}
	return result
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func truncateText(value string, limit int) string {
	value = strings.TrimSpace(value)
	if limit <= 0 {
		return value
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
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
	return candidate.score >= constellationStrongMatchThreshold
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
