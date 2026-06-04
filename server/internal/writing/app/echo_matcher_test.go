package app

import (
	"context"
	"testing"
	"time"

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
	// vec=[1,0] vs [0.7,~0.714]=0.7, vs [0.8,0.6]=0.8
	cur := &domain.Moment{ID: "m1", Embeddings: []domain.EmbeddingEntry{
		{Model: "test", Embedding: []float32{1, 0}},
	}}
	history := []domain.Moment{
		{ID: "mid", Embeddings: []domain.EmbeddingEntry{
			{Model: "test", Embedding: []float32{0.7, 0.71414286}}, // sim = 0.7
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

func TestDefaultEchoMatcher_SimilarityJustAboveThreshold(t *testing.T) {
	matcher := NewDefaultEchoMatcher()
	// cos(theta) is just above the 0.65 threshold.
	cur := &domain.Moment{ID: "m1", Embeddings: []domain.EmbeddingEntry{
		{Model: "test", Embedding: []float32{1, 0}},
	}}
	history := []domain.Moment{
		{ID: "threshold", Embeddings: []domain.EmbeddingEntry{
			{Model: "test", Embedding: []float32{0.651, 0.7590771}}, // sim = 0.651
		}},
	}
	matches, err := matcher.Match(context.Background(), cur, history)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 1 {
		t.Fatalf("expected 1 match above threshold (sim >= 0.65), got %d", len(matches))
	}
}

func TestDefaultEchoMatcher_ExcludesSameTrace(t *testing.T) {
	matcher := NewDefaultEchoMatcher()
	now := time.Date(2026, 6, 4, 10, 0, 0, 0, time.UTC)
	cur := &domain.Moment{
		ID:        "m1",
		TraceID:   "trace-current",
		CreatedAt: now,
		Embeddings: []domain.EmbeddingEntry{
			{Model: "test", Embedding: []float32{1, 0}},
		},
	}
	history := []domain.Moment{
		{ID: "same-trace", TraceID: "trace-current", CreatedAt: now.Add(-24 * time.Hour), Embeddings: []domain.EmbeddingEntry{
			{Model: "test", Embedding: []float32{1, 0}},
		}},
		{ID: "other-trace", TraceID: "trace-old", CreatedAt: now.Add(-24 * time.Hour), Embeddings: []domain.EmbeddingEntry{
			{Model: "test", Embedding: []float32{1, 0}},
		}},
	}
	matches, err := matcher.Match(context.Background(), cur, history)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 1 || matches[0].MomentID != "other-trace" {
		t.Fatalf("expected only other trace match, got %+v", matches)
	}
}

func TestDefaultEchoMatcher_DoesNotFilterRecentCandidates(t *testing.T) {
	matcher := NewDefaultEchoMatcher()
	now := time.Date(2026, 6, 4, 10, 0, 0, 0, time.UTC)
	cur := &domain.Moment{
		ID:        "m1",
		TraceID:   "trace-current",
		CreatedAt: now,
		Embeddings: []domain.EmbeddingEntry{
			{Model: "test", Embedding: []float32{1, 0}},
		},
	}
	history := []domain.Moment{
		{ID: "recent", TraceID: "trace-recent", CreatedAt: now.Add(-5 * time.Minute), Embeddings: []domain.EmbeddingEntry{
			{Model: "test", Embedding: []float32{1, 0}},
		}},
		{ID: "old-enough", TraceID: "trace-old", CreatedAt: now.Add(-24 * time.Hour), Embeddings: []domain.EmbeddingEntry{
			{Model: "test", Embedding: []float32{1, 0}},
		}},
	}
	matches, err := matcher.Match(context.Background(), cur, history)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 2 {
		t.Fatalf("expected recent and old-enough matches, got %+v", matches)
	}
	seen := map[string]bool{}
	for _, match := range matches {
		seen[match.MomentID] = true
	}
	if !seen["recent"] || !seen["old-enough"] {
		t.Fatalf("expected recent and old-enough matches, got %+v", matches)
	}
}

func TestDefaultEchoMatcher_DedupesSameHistoricalTraceByEchoScore(t *testing.T) {
	matcher := NewDefaultEchoMatcher()
	now := time.Date(2026, 6, 4, 10, 0, 0, 0, time.UTC)
	cur := &domain.Moment{
		ID:        "m1",
		TraceID:   "trace-current",
		CreatedAt: now,
		Embeddings: []domain.EmbeddingEntry{
			{Model: "test", Embedding: []float32{1, 0}},
		},
	}
	history := []domain.Moment{
		{ID: "same-trace-lower", TraceID: "trace-shared", CreatedAt: now.Add(-24 * time.Hour), Embeddings: []domain.EmbeddingEntry{
			{Model: "test", Embedding: []float32{0.7, 0.71414286}},
		}},
		{ID: "same-trace-higher", TraceID: "trace-shared", CreatedAt: now.Add(-24 * time.Hour), Embeddings: []domain.EmbeddingEntry{
			{Model: "test", Embedding: []float32{0.8, 0.6}},
		}},
	}
	matches, err := matcher.Match(context.Background(), cur, history)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 1 || matches[0].MomentID != "same-trace-higher" {
		t.Fatalf("expected higher scoring candidate for shared trace, got %+v", matches)
	}
}

func TestDefaultEchoMatcher_LimitsMatchesToThree(t *testing.T) {
	matcher := NewDefaultEchoMatcher()
	now := time.Date(2026, 6, 4, 10, 0, 0, 0, time.UTC)
	cur := &domain.Moment{
		ID:        "m1",
		TraceID:   "trace-current",
		CreatedAt: now,
		Embeddings: []domain.EmbeddingEntry{
			{Model: "test", Embedding: []float32{1, 0}},
		},
	}
	history := []domain.Moment{
		{ID: "h1", TraceID: "t1", CreatedAt: now.Add(-24 * time.Hour), Embeddings: []domain.EmbeddingEntry{{Model: "test", Embedding: []float32{1, 0}}}},
		{ID: "h2", TraceID: "t2", CreatedAt: now.Add(-24 * time.Hour), Embeddings: []domain.EmbeddingEntry{{Model: "test", Embedding: []float32{0.95, 0.3122499}}}},
		{ID: "h3", TraceID: "t3", CreatedAt: now.Add(-24 * time.Hour), Embeddings: []domain.EmbeddingEntry{{Model: "test", Embedding: []float32{0.9, 0.4358899}}}},
		{ID: "h4", TraceID: "t4", CreatedAt: now.Add(-24 * time.Hour), Embeddings: []domain.EmbeddingEntry{{Model: "test", Embedding: []float32{0.85, 0.5267827}}}},
	}
	matches, err := matcher.Match(context.Background(), cur, history)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 3 {
		t.Fatalf("expected 3 matches, got %d", len(matches))
	}
}

func TestDefaultEchoMatcher_OrdersByEchoScoreButKeepsRawSimilarity(t *testing.T) {
	matcher := NewDefaultEchoMatcher()
	now := time.Date(2026, 6, 4, 10, 0, 0, 0, time.UTC)
	cur := &domain.Moment{
		ID:        "m1",
		TraceID:   "trace-current",
		CreatedAt: now,
		Embeddings: []domain.EmbeddingEntry{
			{Model: "test", Embedding: []float32{1, 0}},
		},
	}
	history := []domain.Moment{
		{ID: "yesterday", TraceID: "t1", CreatedAt: now.Add(-24 * time.Hour), Embeddings: []domain.EmbeddingEntry{
			{Model: "test", Embedding: []float32{0.66, 0.7510659}}, // score 0.67
		}},
		{ID: "old", TraceID: "t2", CreatedAt: now.Add(-30 * 24 * time.Hour), Embeddings: []domain.EmbeddingEntry{
			{Model: "test", Embedding: []float32{0.68, 0.7332121}}, // score 0.685
		}},
	}
	matches, err := matcher.Match(context.Background(), cur, history)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(matches) != 2 {
		t.Fatalf("expected 2 matches, got %d", len(matches))
	}
	if matches[0].MomentID != "old" {
		t.Fatalf("expected old candidate first by echo_score, got %+v", matches)
	}
	if matches[0].Similarity < 0.67 || matches[0].Similarity > 0.69 {
		t.Fatalf("expected raw cosine similarity around 0.68, got %f", matches[0].Similarity)
	}
}
