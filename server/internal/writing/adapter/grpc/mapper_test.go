package grpc

import (
	"testing"
	"time"

	"ego-server/internal/writing/domain"
)

func TestMomentToProto(t *testing.T) {
	now := time.Now().UTC()
	m := domain.Moment{
		ID:        "mom-1",
		Content:   "hello world",
		CreatedAt: now,
		Connected: false,
		TraceID:   "tr-1",
	}

	result := momentToProto(m)

	if result.Id != "mom-1" {
		t.Fatalf("expected Id 'mom-1', got %q", result.Id)
	}
	if result.Content != "hello world" {
		t.Fatalf("expected Content 'hello world', got %q", result.Content)
	}
	if result.CreatedAt != now.UnixMilli() {
		t.Fatalf("expected CreatedAt %d, got %d", now.UnixMilli(), result.CreatedAt)
	}
	if result.Connected {
		t.Fatal("expected Connected to be false")
	}
	if result.TraceId != "tr-1" {
		t.Fatalf("expected TraceId 'tr-1', got %q", result.TraceId)
	}
}

func TestEchoToProto_Nil(t *testing.T) {
	result := echoToProto(nil)
	if result != nil {
		t.Fatal("expected nil for nil echo")
	}
}

func TestEchoToProto(t *testing.T) {
	now := time.Now().UTC()
	e := &domain.Echo{
		ID: "echo-1",
		TargetMoment: domain.Moment{
			ID:        "mom-target",
			Content:   "I remember this",
			CreatedAt: now,
			Connected: true,
			TraceID:   "tr-2",
		},
		Candidates: []domain.Moment{
			{ID: "cand-1", Content: "alt 1", CreatedAt: now},
			{ID: "cand-2", Content: "alt 2", CreatedAt: now},
		},
		Similarity: 0.85,
	}

	result := echoToProto(e)

	if result.Id != "echo-1" {
		t.Fatalf("expected Id 'echo-1', got %q", result.Id)
	}
	if result.TargetMoment.Id != "mom-target" {
		t.Fatalf("expected TargetMoment.Id 'mom-target', got %q", result.TargetMoment.Id)
	}
	if result.TargetMoment.Content != "I remember this" {
		t.Fatalf("expected TargetMoment.Content 'I remember this', got %q", result.TargetMoment.Content)
	}
	if len(result.Candidates) != 2 {
		t.Fatalf("expected 2 candidates, got %d", len(result.Candidates))
	}
	if result.Candidates[0].Id != "cand-1" {
		t.Fatalf("expected first candidate 'cand-1', got %q", result.Candidates[0].Id)
	}
	if result.Similarity != 0.85 {
		t.Fatalf("expected Similarity 0.85, got %v", result.Similarity)
	}
}

func TestInsightToProto_Nil(t *testing.T) {
	result := insightToProto(nil)
	if result != nil {
		t.Fatal("expected nil for nil insight")
	}
}

func TestInsightToProto(t *testing.T) {
	i := &domain.Insight{
		ID:               "ins-1",
		Text:             "You seem to be revisiting old themes.",
		RelatedMomentIDs: []string{"mom-1", "mom-2"},
	}

	result := insightToProto(i)

	if result.Id != "ins-1" {
		t.Fatalf("expected Id 'ins-1', got %q", result.Id)
	}
	if result.Text != "You seem to be revisiting old themes." {
		t.Fatalf("unexpected text: %q", result.Text)
	}
	if len(result.RelatedMomentIds) != 2 {
		t.Fatalf("expected 2 related moment IDs, got %d", len(result.RelatedMomentIds))
	}
}
