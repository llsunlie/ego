package postgres

import (
	"context"
	"testing"

	"ego-server/internal/writing/domain"

	"github.com/google/uuid"
)

func newTestTrace() domain.Trace {
	return domain.Trace{
		ID:         uuid.NewString(),
		UserID:     uuid.NewString(),
		Motivation: "direct",
		Stashed:    false,
	}
}

func TestTraceRepo_CreateAndGetByID(t *testing.T) {
	q := testQueries(t)
	repo := NewTraceRepository(q)

	tr := newTestTrace()
	err := repo.Create(context.Background(), &tr)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	got, err := repo.GetByID(context.Background(), tr.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.ID != tr.ID {
		t.Fatalf("expected ID %s, got %s", tr.ID, got.ID)
	}
	if got.UserID != tr.UserID {
		t.Fatalf("expected UserID %s, got %s", tr.UserID, got.UserID)
	}
	if got.Motivation != tr.Motivation {
		t.Fatalf("expected Motivation %q, got %q", tr.Motivation, got.Motivation)
	}
	if got.Stashed {
		t.Fatal("expected Stashed to be false")
	}
	if got.CreatedAt.IsZero() {
		t.Fatal("expected non-zero CreatedAt")
	}
}

func TestTraceRepo_GetByID_NotFound(t *testing.T) {
	q := testQueries(t)
	repo := NewTraceRepository(q)

	_, err := repo.GetByID(context.Background(), uuid.NewString())
	if err == nil {
		t.Fatal("expected error for nonexistent trace")
	}
}

func TestTraceRepo_Update(t *testing.T) {
	q := testQueries(t)
	repo := NewTraceRepository(q)

	tr := newTestTrace()
	err := repo.Create(context.Background(), &tr)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	tr.Stashed = true
	err = repo.Update(context.Background(), &tr)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := repo.GetByID(context.Background(), tr.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if !got.Stashed {
		t.Fatal("expected Stashed to be true after update")
	}
}

func TestTraceRepo_Delete(t *testing.T) {
	q := testQueries(t)
	repo := NewTraceRepository(q)

	tr := newTestTrace()
	err := repo.Create(context.Background(), &tr)
	if err != nil {
		t.Fatalf("Create: %v", err)
	}

	err = repo.Delete(context.Background(), tr.ID)
	if err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err = repo.GetByID(context.Background(), tr.ID)
	if err == nil {
		t.Fatal("expected error after deleting trace")
	}
}
