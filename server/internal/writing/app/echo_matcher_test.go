package app

import (
	"context"
	"testing"

	"ego-server/internal/writing/domain"
)

func TestDefaultEchoMatcher_EmptyHistory(t *testing.T) {
	matcher := NewDefaultEchoMatcher()
	cur := &domain.Moment{ID: "m1", Embeddings: []domain.EmbeddingEntry{
		{Model: "test", Embedding: []float32{0.1, 0.2}},
	}}
	matches, err := matcher.Match(context.Background(), cur, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if matches != nil {
		t.Fatalf("expected nil matches for empty history, got %d", len(matches))
	}
}

func TestDefaultEchoMatcher_NoEmbeddingOnCurrent(t *testing.T) {
	matcher := NewDefaultEchoMatcher()
	cur := &domain.Moment{ID: "m1"} // no embeddings
	history := []domain.Moment{
		{ID: "h1", Embeddings: []domain.EmbeddingEntry{
			{Model: "test", Embedding: []float32{0.1, 0.2}},
		}},
	}
	matches, err := matcher.Match(context.Background(), cur, history)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if matches != nil {
		t.Fatalf("expected nil when current has no embedding, got %d", len(matches))
	}
}

func TestDefaultEchoMatcher_AllHistorySkipped_NoEmbeddings(t *testing.T) {
	matcher := NewDefaultEchoMatcher()
	cur := &domain.Moment{ID: "m1", Embeddings: []domain.EmbeddingEntry{
		{Model: "test", Embedding: []float32{0.1, 0.2}},
	}}
	history := []domain.Moment{
		{ID: "h1"}, // no embeddings
		{ID: "h2"}, // no embeddings
	}
	matches, err := matcher.Match(context.Background(), cur, history)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if matches != nil {
		t.Fatalf("expected nil when all history skipped, got %d", len(matches))
	}
}

func TestDefaultEchoMatcher_BelowThreshold_NotMatched(t *testing.T) {
	matcher := NewDefaultEchoMatcher()
	// orthogonal vectors → cosine similarity = 0
	cur := &domain.Moment{ID: "m1", Embeddings: []domain.EmbeddingEntry{
		{Model: "test", Embedding: []float32{1, 0, 0}},
	}}
	history := []domain.Moment{
		{ID: "h1", Embeddings: []domain.EmbeddingEntry{
			{Model: "test", Embedding: []float32{0, 1, 0}},
		}},
	}
	matches, err := matcher.Match(context.Background(), cur, history)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if matches != nil {
		t.Fatalf("expected nil for below-threshold similarity, got %d matches", len(matches))
	}
}

func TestDefaultEchoMatcher_AboveThreshold_Matched(t *testing.T) {
	matcher := NewDefaultEchoMatcher()
	// identical vectors → cosine similarity = 1.0
	cur := &domain.Moment{ID: "m1", Embeddings: []domain.EmbeddingEntry{
		{Model: "test", Embedding: []float32{1, 0, 0}},
	}}
	history := []domain.Moment{
		{ID: "h1", Embeddings: []domain.EmbeddingEntry{
			{Model: "test", Embedding: []float32{1, 0, 0}},
		}},
	}
	matches, err := matcher.Match(context.Background(), cur, history)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
	if matches[0].MomentID != "h1" {
		t.Fatalf("expected match to h1, got %s", matches[0].MomentID)
	}
	if matches[0].Similarity != 1.0 {
		t.Fatalf("expected similarity 1.0, got %f", matches[0].Similarity)
	}
}

func TestDefaultEchoMatcher_SortedBySimilarityDescending(t *testing.T) {
	matcher := NewDefaultEchoMatcher()
	// vec=[1,0] vs [1,0]=1.0, vs [0.6,0.8]=0.6, vs [0.8,0.6]=0.8
	cur := &domain.Moment{ID: "m1", Embeddings: []domain.EmbeddingEntry{
		{Model: "test", Embedding: []float32{1, 0}},
	}}
	history := []domain.Moment{
		{ID: "mid", Embeddings: []domain.EmbeddingEntry{
			{Model: "test", Embedding: []float32{0.6, 0.8}}, // sim = 0.6
		}},
		{ID: "high", Embeddings: []domain.EmbeddingEntry{
			{Model: "test", Embedding: []float32{0.8, 0.6}}, // sim = 0.8
		}},
	}
	matches, err := matcher.Match(context.Background(), cur, history)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matches))
	}
	if matches[0].MomentID != "high" || matches[0].Similarity < matches[1].Similarity {
		t.Fatalf("expected sorted descending: got [0]=%s(%.2f), [1]=%s(%.2f)",
			matches[0].MomentID, matches[0].Similarity,
			matches[1].MomentID, matches[1].Similarity)
	}
}

func TestDefaultEchoMatcher_MixedAboveAndBelowThreshold(t *testing.T) {
	matcher := NewDefaultEchoMatcher()
	cur := &domain.Moment{ID: "m1", Embeddings: []domain.EmbeddingEntry{
		{Model: "test", Embedding: []float32{1, 0}},
	}}
	history := []domain.Moment{
		{ID: "sim-high", Embeddings: []domain.EmbeddingEntry{
			{Model: "test", Embedding: []float32{1, 0}}, // sim = 1.0 → match
		}},
		{ID: "sim-low", Embeddings: []domain.EmbeddingEntry{
			{Model: "test", Embedding: []float32{0, 1}}, // sim = 0 → no match
		}},
	}
	matches, err := matcher.Match(context.Background(), cur, history)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected 1 match, got %d", len(matches))
	}
	if matches[0].MomentID != "sim-high" {
		t.Fatalf("expected sim-high, got %s", matches[0].MomentID)
	}
}

func TestDefaultEchoMatcher_MixedSkippedAndMatched(t *testing.T) {
	matcher := NewDefaultEchoMatcher()
	cur := &domain.Moment{ID: "m1", Embeddings: []domain.EmbeddingEntry{
		{Model: "test", Embedding: []float32{1, 0}},
	}}
	history := []domain.Moment{
		{ID: "no-emb"}, // skipped
		{ID: "matched", Embeddings: []domain.EmbeddingEntry{
			{Model: "test", Embedding: []float32{1, 0}}, // sim = 1.0
		}},
	}
	matches, err := matcher.Match(context.Background(), cur, history)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected 1 match (skipped one without embedding), got %d", len(matches))
	}
	if matches[0].MomentID != "matched" {
		t.Fatalf("expected matched, got %s", matches[0].MomentID)
	}
}

func TestDefaultEchoMatcher_SimilarityAtThreshold(t *testing.T) {
	matcher := NewDefaultEchoMatcher()
	// cos(θ) where dot product / norms = 0.55:
	// [1,0] dot [0.55, sqrt(1-0.55^2)] = 0.55
	// norm of [0.55, ~0.835] → sqrt(0.3025 + 0.6975) = 1.0
	// So sim = 0.55 / 1.0 = 0.55 → should match (>= threshold)
	cur := &domain.Moment{ID: "m1", Embeddings: []domain.EmbeddingEntry{
		{Model: "test", Embedding: []float32{1, 0}},
	}}
	history := []domain.Moment{
		{ID: "threshold", Embeddings: []domain.EmbeddingEntry{
			{Model: "test", Embedding: []float32{0.55, 0.835164}}, // sim ≈ 0.55
		}},
	}
	matches, err := matcher.Match(context.Background(), cur, history)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected 1 match at threshold (sim >= 0.55), got %d", len(matches))
	}
}
