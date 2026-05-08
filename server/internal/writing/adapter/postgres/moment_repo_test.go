package postgres

import (
	"context"
	"testing"

	"ego-server/internal/writing/domain"

	"github.com/google/uuid"
)

func newTestMoment() domain.Moment {
	return domain.Moment{
		ID:        uuid.NewString(),
		TraceID:   uuid.NewString(),
		UserID:    uuid.NewString(),
		Content:   "a moment of reflection",
		Embedding: testEmbeddingSlice(),
		Connected: false,
	}
}

func testEmbeddingSlice() []float32 {
	v := make([]float32, 1536)
	v[0], v[1], v[2] = 0.1, 0.2, 0.3
	return v
}

func TestMomentRepo_CreateAndGetByID(t *testing.T) {
	q := testQueries(t)
	repo := NewMomentRepository(q)

	m := newTestMoment()
	err := repo.Create(context.Background(), &m)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByID(context.Background(), m.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.ID != m.ID {
		t.Fatalf("expected ID %s, got %s", m.ID, got.ID)
	}
	if got.Content != m.Content {
		t.Fatalf("expected Content %q, got %q", m.Content, got.Content)
	}
	if got.TraceID != m.TraceID {
		t.Fatalf("expected TraceID %s, got %s", m.TraceID, got.TraceID)
	}
	if got.UserID != m.UserID {
		t.Fatalf("expected UserID %s, got %s", m.UserID, got.UserID)
	}
	if got.Connected {
		t.Fatal("expected Connected to be false")
	}
}

func TestMomentRepo_GetByID_NotFound(t *testing.T) {
	q := testQueries(t)
	repo := NewMomentRepository(q)

	_, err := repo.GetByID(context.Background(), uuid.NewString())
	if err == nil {
		t.Fatal("expected error for nonexistent moment")
	}
}

func TestMomentRepo_ListByTraceID(t *testing.T) {
	q := testQueries(t)
	repo := NewMomentRepository(q)

	traceID := uuid.NewString()
	userID := uuid.NewString()

	// Create 2 moments with the same trace
	for range 2 {
		m := domain.Moment{
			ID:        uuid.NewString(),
			TraceID:   traceID,
			UserID:    userID,
			Content:   "moment in trace",
			Embedding: testEmbeddingSlice(),
			Connected: false,
		}
		if err := repo.Create(context.Background(), &m); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	// Create a moment with a different trace
	other := domain.Moment{
		ID:        uuid.NewString(),
		TraceID:   uuid.NewString(),
		UserID:    userID,
		Content:   "other",
		Embedding: testEmbeddingSlice(),
		Connected: false,
	}
	if err := repo.Create(context.Background(), &other); err != nil {
		t.Fatalf("Create: %v", err)
	}

	items, err := repo.ListByTraceID(context.Background(), traceID)
	if err != nil {
		t.Fatalf("ListByTraceID: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}
}

func TestMomentRepo_ListByUserID(t *testing.T) {
	q := testQueries(t)
	repo := NewMomentRepository(q)

	userID := uuid.NewString()

	for range 3 {
		m := domain.Moment{
			ID:        uuid.NewString(),
			TraceID:   uuid.NewString(),
			UserID:    userID,
			Content:   "moment for user",
			Embedding: testEmbeddingSlice(),
			Connected: false,
		}
		if err := repo.Create(context.Background(), &m); err != nil {
			t.Fatalf("Create: %v", err)
		}
	}

	items, err := repo.ListByUserID(context.Background(), userID)
	if err != nil {
		t.Fatalf("ListByUserID: %v", err)
	}
	if len(items) < 3 {
		t.Fatalf("expected at least 3 items, got %d", len(items))
	}
}
