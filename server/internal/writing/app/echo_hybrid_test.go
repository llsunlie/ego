package app

import (
	"reflect"
	"testing"

	"ego-server/internal/writing/domain"
)

func TestMergeEchoCandidatesRRF_PromotesCandidatesSeenByBothRetrievers(t *testing.T) {
	dense := []domain.Moment{{ID: "dense-only"}, {ID: "both"}}
	sparse := []domain.Moment{{ID: "both"}, {ID: "sparse-only"}}

	got := mergeEchoCandidatesRRF(dense, sparse, 60, 3)
	ids := momentIDs(got)
	want := []string{"both", "dense-only", "sparse-only"}
	if !reflect.DeepEqual(ids, want) {
		t.Fatalf("expected %v, got %v", want, ids)
	}
}

func TestMergeEchoCandidatesRRF_LimitsResults(t *testing.T) {
	dense := []domain.Moment{{ID: "d1"}, {ID: "d2"}, {ID: "d3"}}
	sparse := []domain.Moment{{ID: "s1"}, {ID: "s2"}}

	got := mergeEchoCandidatesRRF(dense, sparse, 60, 2)
	if len(got) != 2 {
		t.Fatalf("expected 2 candidates, got %d", len(got))
	}
}

func TestOrderMomentsByIDs(t *testing.T) {
	moments := []domain.Moment{{ID: "b"}, {ID: "a"}, {ID: "c"}}
	got := orderMomentsByIDs([]string{"a", "c"}, moments)
	ids := momentIDs(got)
	want := []string{"a", "c"}
	if !reflect.DeepEqual(ids, want) {
		t.Fatalf("expected %v, got %v", want, ids)
	}
}

func momentIDs(moments []domain.Moment) []string {
	ids := make([]string, len(moments))
	for i, moment := range moments {
		ids[i] = moment.ID
	}
	return ids
}
