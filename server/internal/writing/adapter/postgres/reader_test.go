package postgres

import (
	"context"
	"testing"
	"time"

	"ego-server/internal/writing/domain"

	"github.com/google/uuid"
)

func testEmbeddingEntries() []domain.EmbeddingEntry {
	return []domain.EmbeddingEntry{
		{Model: "test", Embedding: []float32{0.1, 0.2, 0.3}},
	}
}

func TestReader_GetByID_MomentReader(t *testing.T) {
	q := testQueries(t)
	repo := NewMomentRepository(q)
	reader := NewReader(q)

	m := newTestMoment()
	if err := repo.Create(context.Background(), &m); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := reader.GetByID(context.Background(), m.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.ID != m.ID {
		t.Fatalf("expected ID %s, got %s", m.ID, got.ID)
	}
	if got.Content != m.Content {
		t.Fatalf("expected Content %q, got %q", m.Content, got.Content)
	}
}

func TestReader_GetByID_NotFound(t *testing.T) {
	q := testQueries(t)
	reader := NewReader(q)

	_, err := reader.GetByID(context.Background(), uuid.NewString())
	if err == nil {
		t.Fatal("expected error for nonexistent moment")
	}
}

func TestReader_ListByUserID_Cursor(t *testing.T) {
	q := testQueries(t)
	repo := NewMomentRepository(q)
	reader := NewReader(q)

	userID := uuid.NewString()
	traceID := uuid.NewString()

	// Create 3 moments with different timestamps
	for range 3 {
		m := domain.Moment{
			ID:         uuid.NewString(),
			TraceID:    traceID,
			UserID:     userID,
			Content:    "moment",
			Embeddings: testEmbeddingEntries(),
		}
		err := repo.Create(context.Background(), &m)
		if err != nil {
			t.Fatalf("Create: %v", err)
		}
		time.Sleep(time.Millisecond * 2)
	}

	items, nextCursor, hasMore, err := reader.ListByUserID(context.Background(), userID, "", 2)
	if err != nil {
		t.Fatalf("ListByUserID: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
	if !hasMore {
		t.Fatal("expected hasMore=true")
	}
	if nextCursor == "" {
		t.Fatal("expected non-empty nextCursor")
	}

	// Get next page
	items2, _, hasMore2, err := reader.ListByUserID(context.Background(), userID, nextCursor, 2)
	if err != nil {
		t.Fatalf("ListByUserID page 2: %v", err)
	}
	if len(items2) < 1 {
		t.Fatalf("expected at least 1 item on page 2, got %d", len(items2))
	}
	if hasMore2 {
		t.Fatal("expected hasMore=false on page 2")
	}
}

func TestReader_RandomByUserID(t *testing.T) {
	q := testQueries(t)
	repo := NewMomentRepository(q)
	reader := NewReader(q)

	userID := uuid.NewString()
	for range 5 {
		m := domain.Moment{
			ID:         uuid.NewString(),
			TraceID:    uuid.NewString(),
			UserID:     userID,
			Content:    "moment",
			Embeddings: testEmbeddingEntries(),
		}
		if err := repo.Create(context.Background(), &m); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	items, err := reader.RandomByUserID(context.Background(), userID, 3)
	if err != nil {
		t.Fatalf("RandomByUserID: %v", err)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
}

func TestReader_RandomByUserID_NoMoments(t *testing.T) {
	q := testQueries(t)
	reader := NewReader(q)

	items, err := reader.RandomByUserID(context.Background(), uuid.NewString(), 5)
	if err != nil {
		t.Fatalf("RandomByUserID: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items for user with no moments, got %d", len(items))
	}
}

func TestReader_GetTraceByID(t *testing.T) {
	q := testQueries(t)
	traceRepo := NewTraceRepository(q)
	reader := NewReader(q)

	tr := newTestTrace()
	if err := traceRepo.Create(context.Background(), &tr); err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := reader.GetTraceByID(context.Background(), tr.ID)
	if err != nil {
		t.Fatalf("GetTraceByID: %v", err)
	}
	if got.ID != tr.ID {
		t.Fatalf("expected ID %s, got %s", tr.ID, got.ID)
	}
}

func TestReader_GetTraceByID_NotFound(t *testing.T) {
	q := testQueries(t)
	reader := NewReader(q)

	_, err := reader.GetTraceByID(context.Background(), uuid.NewString())
	if err == nil {
		t.Fatal("expected error for nonexistent trace")
	}
}

func TestReader_ListMomentsByTraceID(t *testing.T) {
	q := testQueries(t)
	repo := NewMomentRepository(q)
	reader := NewReader(q)

	traceID := uuid.NewString()
	userID := uuid.NewString()

	for range 2 {
		m := domain.Moment{
			ID:         uuid.NewString(),
			TraceID:    traceID,
			UserID:     userID,
			Content:    "moment in trace",
			Embeddings: testEmbeddingEntries(),
		}
		if err := repo.Create(context.Background(), &m); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	items, err := reader.ListMomentsByTraceID(context.Background(), traceID)
	if err != nil {
		t.Fatalf("ListMomentsByTraceID: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
}

func TestReader_ListTracesByUserID(t *testing.T) {
	q := testQueries(t)
	traceRepo := NewTraceRepository(q)
	reader := NewReader(q)

	userID := uuid.NewString()

	for range 3 {
		tr := domain.Trace{
			ID:         uuid.NewString(),
			UserID:     userID,
			Motivation: "direct",
			Stashed:    false,
		}
		if err := traceRepo.Create(context.Background(), &tr); err != nil {
			t.Fatalf("Create: %v", err)
		}
		time.Sleep(time.Millisecond * 2)
	}

	traces, nextCursor, hasMore, err := reader.ListTracesByUserID(context.Background(), userID, "", 2)
	if err != nil {
		t.Fatalf("ListTracesByUserID: %v", err)
	}
	if len(traces) != 2 {
		t.Fatalf("expected 2 traces on first page, got %d", len(traces))
	}
	if !hasMore {
		t.Fatal("expected hasMore=true on first page")
	}
	if nextCursor == "" {
		t.Fatal("expected non-empty nextCursor")
	}

	// Page 2
	traces2, _, hasMore2, err := reader.ListTracesByUserID(context.Background(), userID, nextCursor, 2)
	if err != nil {
		t.Fatalf("ListTracesByUserID page 2: %v", err)
	}
	if len(traces2) < 1 {
		t.Fatalf("expected at least 1 trace on page 2, got %d", len(traces2))
	}
	if hasMore2 {
		t.Fatal("expected hasMore=false on page 2")
	}
}
