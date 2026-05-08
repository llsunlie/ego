package sqlc

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func createTestTrace(t *testing.T, q *Queries) (trace Trace) {
	t.Helper()
	id := uuid.New()
	now := time.Now().UTC()
	params := CreateTraceParams{
		ID:        pgUUID(id.String()),
		UserID:    pgUUID(uuid.New().String()),
		Topic:     pgtype.Text{String: "test topic", Valid: true},
		CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
	}
	err := q.CreateTrace(context.Background(), params)
	if err != nil {
		t.Fatalf("CreateTrace: %v", err)
	}
	return Trace{
		ID:        params.ID,
		UserID:    params.UserID,
		Topic:     params.Topic,
		CreatedAt: params.CreatedAt,
		UpdatedAt: params.UpdatedAt,
	}
}

func TestCreateTrace(t *testing.T) {
	pool := testPool(t)
	q := New(pool)

	id := uuid.New()
	now := time.Now().UTC()

	err := q.CreateTrace(context.Background(), CreateTraceParams{
		ID:        pgUUID(id.String()),
		UserID:    pgUUID(uuid.New().String()),
		Topic:     pgtype.Text{String: "my topic", Valid: true},
		CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
	})
	if err != nil {
		t.Fatalf("CreateTrace: %v", err)
	}

	pool.Exec(context.Background(), "DELETE FROM traces WHERE id = $1", pgUUID(id.String()))
}

func TestCreateTrace_NullTopic(t *testing.T) {
	pool := testPool(t)
	q := New(pool)

	id := uuid.New()
	now := time.Now().UTC()

	err := q.CreateTrace(context.Background(), CreateTraceParams{
		ID:        pgUUID(id.String()),
		UserID:    pgUUID(uuid.New().String()),
		Topic:     pgtype.Text{Valid: false},
		CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
		UpdatedAt: pgtype.Timestamptz{Time: now, Valid: true},
	})
	if err != nil {
		t.Fatalf("CreateTrace with null topic: %v", err)
	}

	got, err := q.GetTraceByID(context.Background(), pgUUID(id.String()))
	if err != nil {
		t.Fatalf("GetTraceByID: %v", err)
	}
	if got.Topic.Valid {
		t.Fatal("expected topic to be NULL")
	}

	pool.Exec(context.Background(), "DELETE FROM traces WHERE id = $1", pgUUID(id.String()))
}

func TestGetTraceByID(t *testing.T) {
	pool := testPool(t)
	q := New(pool)

	tr := createTestTrace(t, q)

	got, err := q.GetTraceByID(context.Background(), tr.ID)
	if err != nil {
		t.Fatalf("GetTraceByID: %v", err)
	}
	if got.Topic.String != tr.Topic.String {
		t.Fatalf("expected Topic %q, got %q", tr.Topic.String, got.Topic.String)
	}

	pool.Exec(context.Background(), "DELETE FROM traces WHERE id = $1", tr.ID)
}

func TestGetTraceByID_NotFound(t *testing.T) {
	pool := testPool(t)
	q := New(pool)

	_, err := q.GetTraceByID(context.Background(), pgUUID(uuid.New().String()))
	if err == nil {
		t.Fatal("expected error for nonexistent trace")
	}
}

func TestUpdateTrace(t *testing.T) {
	pool := testPool(t)
	q := New(pool)

	tr := createTestTrace(t, q)

	newTopic := pgtype.Text{String: "updated topic", Valid: true}
	newUpdatedAt := pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true}
	err := q.UpdateTrace(context.Background(), UpdateTraceParams{
		ID:        tr.ID,
		Topic:     newTopic,
		UpdatedAt: newUpdatedAt,
	})
	if err != nil {
		t.Fatalf("UpdateTrace: %v", err)
	}

	got, err := q.GetTraceByID(context.Background(), tr.ID)
	if err != nil {
		t.Fatalf("GetTraceByID: %v", err)
	}
	if got.Topic.String != "updated topic" {
		t.Fatalf("expected Topic %q, got %q", "updated topic", got.Topic.String)
	}

	pool.Exec(context.Background(), "DELETE FROM traces WHERE id = $1", tr.ID)
}

func TestDeleteTrace(t *testing.T) {
	pool := testPool(t)
	q := New(pool)

	tr := createTestTrace(t, q)

	err := q.DeleteTrace(context.Background(), tr.ID)
	if err != nil {
		t.Fatalf("DeleteTrace: %v", err)
	}

	_, err = q.GetTraceByID(context.Background(), tr.ID)
	if err == nil {
		t.Fatal("expected error after deleting trace")
	}
}

func TestDeleteTrace_NotFound(t *testing.T) {
	pool := testPool(t)
	q := New(pool)

	err := q.DeleteTrace(context.Background(), pgUUID(uuid.New().String()))
	if err != nil {
		t.Fatalf("DeleteTrace on nonexistent should not error: %v", err)
	}
}
