package app

import (
	"sort"

	"ego-server/internal/writing/domain"
)

func mergeEchoCandidatesRRF(dense []domain.Moment, sparse []domain.Moment, rrfK int, limit int) []domain.Moment {
	if rrfK <= 0 {
		rrfK = 60
	}
	if limit <= 0 {
		limit = maxInt(len(dense), len(sparse))
	}

	byID := make(map[string]rankedCandidate)
	add := func(moment domain.Moment, rank int) {
		if moment.ID == "" {
			return
		}
		candidate := byID[moment.ID]
		if candidate.moment.ID == "" {
			candidate.moment = moment
		}
		candidate.score += 1 / float64(rrfK+rank)
		byID[moment.ID] = candidate
	}
	for i, moment := range dense {
		add(moment, i+1)
	}
	for i, moment := range sparse {
		add(moment, i+1)
	}

	ranked := make([]rankedCandidate, 0, len(byID))
	for _, candidate := range byID {
		ranked = append(ranked, candidate)
	}
	sort.Slice(ranked, func(i, j int) bool {
		return ranked[i].score > ranked[j].score
	})
	if len(ranked) > limit {
		ranked = ranked[:limit]
	}

	result := make([]domain.Moment, len(ranked))
	for i, candidate := range ranked {
		result[i] = candidate.moment
	}
	return result
}

func orderMomentsByIDs(ids []string, moments []domain.Moment) []domain.Moment {
	byID := make(map[string]domain.Moment, len(moments))
	for _, moment := range moments {
		byID[moment.ID] = moment
	}
	result := make([]domain.Moment, 0, len(ids))
	for _, id := range ids {
		if moment, ok := byID[id]; ok {
			result = append(result, moment)
		}
	}
	return result
}

type rankedCandidate struct {
	moment domain.Moment
	score  float64
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
