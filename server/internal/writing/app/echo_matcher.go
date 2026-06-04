package app

import (
	"context"
	"sort"
	"time"

	platformai "ego-server/internal/platform/ai"
	"ego-server/internal/platform/logging"
	"ego-server/internal/writing/domain"
)

const (
	echoSimilarityThreshold = 0.65
	echoMaxMatches          = 3
	echoScoreLogLimit       = 10
)

// DefaultEchoMatcher finds historical Moments that resonate with the current
// Moment by comparing their embeddings with cosine similarity.
type DefaultEchoMatcher struct{}

func NewDefaultEchoMatcher() DefaultEchoMatcher {
	return DefaultEchoMatcher{}
}

func (DefaultEchoMatcher) Match(ctx context.Context, current *domain.Moment, history []domain.Moment) ([]domain.MatchedMoment, error) {
	logger := logging.FromContext(ctx)

	if len(current.Embeddings) == 0 {
		logger.WarnContext(ctx, "current moment has no embedding, skipping echo match", "moment_id", current.ID)
		return nil, nil
	}
	curEmb := current.Embeddings[0].Embedding

	rankedByTrace := make(map[string]rankedEchoMatch)
	var ungrouped []rankedEchoMatch
	skippedNoEmbedding := 0
	filteredSameTrace := 0
	scoreLogCapacity := len(history)
	if scoreLogCapacity > echoScoreLogLimit {
		scoreLogCapacity = echoScoreLogLimit
	}
	scoreLogs := make([]map[string]any, 0, scoreLogCapacity)
	for i, h := range history {
		scoreLog := echoCandidateScoreLog(i+1, current, h)
		if h.TraceID != "" && h.TraceID == current.TraceID {
			filteredSameTrace++
			scoreLog["skip_reason"] = "same_trace"
			scoreLogs = appendEchoCandidateScoreLog(scoreLogs, scoreLog)
			continue
		}
		if len(h.Embeddings) == 0 {
			skippedNoEmbedding++
			scoreLog["skip_reason"] = "no_embedding"
			scoreLogs = appendEchoCandidateScoreLog(scoreLogs, scoreLog)
			continue
		}
		sim := platformai.CosineSimilarity(curEmb, h.Embeddings[0].Embedding)
		timeAdjustment := echoTimeAdjustment(current.CreatedAt, h.CreatedAt)
		score := sim + timeAdjustment
		scoreLog["similarity"] = sim
		scoreLog["time_adjustment"] = timeAdjustment
		scoreLog["echo_score"] = score
		scoreLog["passed_threshold"] = score >= echoSimilarityThreshold
		if score < echoSimilarityThreshold {
			scoreLog["skip_reason"] = "below_threshold"
			scoreLogs = appendEchoCandidateScoreLog(scoreLogs, scoreLog)
			continue
		}
		scoreLogs = appendEchoCandidateScoreLog(scoreLogs, scoreLog)

		match := rankedEchoMatch{
			matched:       domain.MatchedMoment{MomentID: h.ID, Similarity: score},
			score:         score,
			rawSimilarity: sim,
		}
		if h.TraceID == "" {
			ungrouped = append(ungrouped, match)
			continue
		}
		existing, ok := rankedByTrace[h.TraceID]
		if !ok || match.score > existing.score {
			rankedByTrace[h.TraceID] = match
		}
	}

	ranked := make([]rankedEchoMatch, 0, len(rankedByTrace)+len(ungrouped))
	ranked = append(ranked, ungrouped...)
	for _, match := range rankedByTrace {
		ranked = append(ranked, match)
	}
	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].score == ranked[j].score {
			return ranked[i].rawSimilarity > ranked[j].rawSimilarity
		}
		return ranked[i].score > ranked[j].score
	})
	if len(ranked) > echoMaxMatches {
		ranked = ranked[:echoMaxMatches]
	}

	matches := make([]domain.MatchedMoment, len(ranked))
	for i, match := range ranked {
		matches[i] = match.matched
	}

	logger.DebugContext(ctx, "echo match candidate scores",
		"current_moment_id", current.ID,
		"current_trace_id", current.TraceID,
		"history_size", len(history),
		"logged_candidate_count", len(scoreLogs),
		"score_log_limit", echoScoreLogLimit,
		"threshold", echoSimilarityThreshold,
		"candidates", scoreLogs,
	)

	topScore := 0.0
	topRawSimilarity := 0.0
	if len(ranked) > 0 {
		topScore = ranked[0].score
		topRawSimilarity = ranked[0].rawSimilarity
	}
	logger.DebugContext(ctx, "echo match done",
		"history_size", len(history),
		"skipped_no_embedding", skippedNoEmbedding,
		"filtered_same_trace", filteredSameTrace,
		"matched", len(matches),
		"threshold", echoSimilarityThreshold,
		"max_matches", echoMaxMatches,
		"top_score", topScore,
		"top_raw_similarity", topRawSimilarity,
	)

	if len(matches) == 0 {
		return nil, nil
	}
	return matches, nil
}

func appendEchoCandidateScoreLog(scoreLogs []map[string]any, scoreLog map[string]any) []map[string]any {
	if len(scoreLogs) >= echoScoreLogLimit {
		return scoreLogs
	}
	return append(scoreLogs, scoreLog)
}

func echoCandidateScoreLog(rank int, current *domain.Moment, candidate domain.Moment) map[string]any {
	scoreLog := map[string]any{
		"rank":      rank,
		"moment_id": candidate.ID,
		"trace_id":  candidate.TraceID,
		"threshold": echoSimilarityThreshold,
	}
	if !candidate.CreatedAt.IsZero() {
		scoreLog["created_at"] = candidate.CreatedAt.Format(time.RFC3339)
	}
	if current != nil && !current.CreatedAt.IsZero() && !candidate.CreatedAt.IsZero() {
		scoreLog["age_hours"] = current.CreatedAt.Sub(candidate.CreatedAt).Hours()
	}
	return scoreLog
}

type rankedEchoMatch struct {
	matched       domain.MatchedMoment
	score         float64
	rawSimilarity float64
}

func echoTimeAdjustment(currentCreatedAt, candidateCreatedAt time.Time) float64 {
	if currentCreatedAt.IsZero() || candidateCreatedAt.IsZero() {
		return 0
	}
	age := currentCreatedAt.Sub(candidateCreatedAt)
	switch {
	case age < 24*time.Hour:
		return -0.01
	case age < 7*24*time.Hour:
		return 0.01
	case age < 90*24*time.Hour:
		return 0.005
	default:
		return 0.01
	}
}
