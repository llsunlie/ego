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
		TraceID:   "tr-1",
		Embeddings: []domain.EmbeddingEntry{
			{Model: "test", Embedding: []float32{0.1}},
		},
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
	e := &domain.Echo{
		ID:               "echo-1",
		MomentID:         "mom-1",
		MatchedMomentIDs: []string{"mom-old-1", "mom-old-2"},
		Similarities:     []float64{0.85, 0.42},
	}

	result := echoToProto(e)

	if result.Id != "echo-1" {
		t.Fatalf("expected Id 'echo-1', got %q", result.Id)
	}
	if result.MomentId != "mom-1" {
		t.Fatalf("expected MomentId 'mom-1', got %q", result.MomentId)
	}
	if len(result.MatchedMomentIds) != 2 {
		t.Fatalf("expected 2 matched moment IDs, got %d", len(result.MatchedMomentIds))
	}
	if result.MatchedMomentIds[0] != "mom-old-1" {
		t.Fatalf("expected first matched 'mom-old-1', got %q", result.MatchedMomentIds[0])
	}
	if len(result.Similarities) != 2 {
		t.Fatalf("expected 2 similarities, got %d", len(result.Similarities))
	}
	if result.Similarities[0] != 0.85 {
		t.Fatalf("expected Similarities[0] 0.85, got %v", result.Similarities[0])
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
		MomentID:         "mom-1",
		EchoID:           "echo-1",
		Text:             "You seem to be revisiting old themes.",
		RelatedMomentIDs: []string{"mom-1", "mom-2"},
	}

	result := insightToProto(i)

	if result.Id != "ins-1" {
		t.Fatalf("expected Id 'ins-1', got %q", result.Id)
	}
	if result.MomentId != "mom-1" {
		t.Fatalf("expected MomentId 'mom-1', got %q", result.MomentId)
	}
	if result.EchoId != "echo-1" {
		t.Fatalf("expected EchoId 'echo-1', got %q", result.EchoId)
	}
	if result.Text != "You seem to be revisiting old themes." {
		t.Fatalf("unexpected text: %q", result.Text)
	}
	if len(result.RelatedMomentIds) != 2 {
		t.Fatalf("expected 2 related moment IDs, got %d", len(result.RelatedMomentIds))
	}
}
