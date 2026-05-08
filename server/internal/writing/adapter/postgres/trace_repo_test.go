package postgres

import (
	"context"
	"testing"

	"ego-server/internal/writing/domain"

	"github.com/google/uuid"
)

func newTestTrace() domain.Trace {
	return domain.Trace{
		ID:     uuid.NewString(),
		UserID: uuid.NewString(),
		Topic:  "a test topic",
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
	if got.Topic != tr.Topic {
		t.Fatalf("expected Topic %q, got %q", tr.Topic, got.Topic)
	}
	if got.CreatedAt.IsZero() {
		t.Fatal("expected non-zero CreatedAt")
	}
	if got.UpdatedAt.IsZero() {
		t.Fatal("expected non-zero UpdatedAt")
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

	tr.Topic = "updated topic"
	err = repo.Update(context.Background(), &tr)
	if err != nil {
		t.Fatalf("Update: %v", err)
	}

	got, err := repo.GetByID(context.Background(), tr.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Topic != "updated topic" {
		t.Fatalf("expected Topic 'updated topic', got %q", got.Topic)
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

func TestTraceRepo_CreateWithEmptyTopic(t *testing.T) {
	q := testQueries(t)
	repo := NewTraceRepository(q)

	tr := domain.Trace{
		ID:     uuid.NewString(),
		UserID: uuid.NewString(),
		Topic:  "",
	}
	err := repo.Create(context.Background(), &tr)
	if err != nil {
		t.Fatalf("Create with empty topic: %v", err)
	}

	got, err := repo.GetByID(context.Background(), tr.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Topic != "" {
		t.Fatalf("expected empty Topic, got %q", got.Topic)
	}
}
