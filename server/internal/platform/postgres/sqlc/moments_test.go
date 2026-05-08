package sqlc

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pgvector/pgvector-go"
)

func testEmbedding() pgvector.Vector {
	v := make([]float32, 1536)
	v[0] = 0.1
	v[1] = 0.2
	v[2] = 0.3
	return pgvector.NewVector(v)
}

func createTestMoment(t *testing.T, q *Queries, userID string) (moment Moment) {
	t.Helper()
	moment = Moment{
		ID:        pgUUID(uuid.New().String()),
		TraceID:   pgUUID(uuid.New().String()),
		UserID:    pgUUID(userID),
		Content:   "test content " + uuid.NewString(),
		Embedding: testEmbedding(),
		Connected: false,
		CreatedAt: pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
	}
	err := q.CreateMoment(context.Background(), CreateMomentParams{
		ID:        moment.ID,
		TraceID:   moment.TraceID,
		UserID:    moment.UserID,
		Content:   moment.Content,
		Embedding: moment.Embedding,
		Connected: moment.Connected,
		CreatedAt: moment.CreatedAt,
	})
	if err != nil {
		t.Fatalf("CreateMoment: %v", err)
	}
	return
}

func TestCreateMoment(t *testing.T) {
	pool := testPool(t)
	q := New(pool)

	userID := uuid.New()

	id := uuid.New()
	now := time.Now().UTC()

	err := q.CreateMoment(context.Background(), CreateMomentParams{
		ID:        pgUUID(id.String()),
		TraceID:   pgUUID(uuid.New().String()),
		UserID:    pgUUID(userID.String()),
		Content:   "hello world",
		Embedding: testEmbedding(),
		Connected: false,
		CreatedAt: pgtype.Timestamptz{Time: now, Valid: true},
	})
	if err != nil {
		t.Fatalf("CreateMoment: %v", err)
	}

	pool.Exec(context.Background(), "DELETE FROM moments WHERE id = $1", pgUUID(id.String()))
}

func TestGetMomentByID(t *testing.T) {
	pool := testPool(t)
	q := New(pool)

	userID := uuid.New()
	m := createTestMoment(t, q, userID.String())

	got, err := q.GetMomentByID(context.Background(), m.ID)
	if err != nil {
		t.Fatalf("GetMomentByID: %v", err)
	}
	if got.Content != m.Content {
		t.Fatalf("expected Content %q, got %q", m.Content, got.Content)
	}

	pool.Exec(context.Background(), "DELETE FROM moments WHERE id = $1", m.ID)
}

func TestGetMomentByID_NotFound(t *testing.T) {
	pool := testPool(t)
	q := New(pool)

	_, err := q.GetMomentByID(context.Background(), pgUUID(uuid.New().String()))
	if err == nil {
		t.Fatal("expected error for nonexistent moment")
	}
}

func TestListMomentsByTraceID(t *testing.T) {
	pool := testPool(t)
	q := New(pool)

	userID := uuid.New()
	traceID := uuid.New()

	// Create 2 moments sharing the same trace_id
	for range 2 {
		err := q.CreateMoment(context.Background(), CreateMomentParams{
			ID:        pgUUID(uuid.New().String()),
			TraceID:   pgUUID(traceID.String()),
			UserID:    pgUUID(userID.String()),
			Content:   "moment in trace",
			Embedding: testEmbedding(),
			Connected: false,
			CreatedAt: pgtype.Timestamptz{Time: time.Now().UTC(), Valid: true},
		})
		if err != nil {
			t.Fatalf("CreateMoment: %v", err)
		}
	}

	// Create a third moment with a different trace_id
	createTestMoment(t, q, userID.String())

	items, err := q.ListMomentsByTraceID(context.Background(), pgUUID(traceID.String()))
	if err != nil {
		t.Fatalf("ListMomentsByTraceID: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	pool.Exec(context.Background(), "DELETE FROM moments WHERE user_id = $1", pgUUID(userID.String()))
}

func TestListMomentsByTraceID_Empty(t *testing.T) {
	pool := testPool(t)
	q := New(pool)

	items, err := q.ListMomentsByTraceID(context.Background(), pgUUID(uuid.New().String()))
	if err != nil {
		t.Fatalf("ListMomentsByTraceID: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected 0 items, got %d", len(items))
	}
}

func TestListMomentsByUserID(t *testing.T) {
	pool := testPool(t)
	q := New(pool)

	userID := uuid.New()
	m1 := createTestMoment(t, q, userID.String())
	m2 := createTestMoment(t, q, userID.String())

	items, err := q.ListMomentsByUserID(context.Background(), pgUUID(userID.String()))
	if err != nil {
		t.Fatalf("ListMomentsByUserID: %v", err)
	}
	if len(items) < 2 {
		t.Fatalf("expected at least 2 items, got %d", len(items))
	}

	pool.Exec(context.Background(), "DELETE FROM moments WHERE id = $1", m1.ID)
	pool.Exec(context.Background(), "DELETE FROM moments WHERE id = $1", m2.ID)
}

func TestListMomentsByUserIDCursor(t *testing.T) {
	pool := testPool(t)
	q := New(pool)

	userID := uuid.New()

	// Create 3 moments with different timestamps
	t1 := time.Now().UTC().Add(-2 * time.Hour)
	t2 := time.Now().UTC().Add(-1 * time.Hour)
	t3 := time.Now().UTC()

	m1 := Moment{
		ID:        pgUUID(uuid.New().String()),
		TraceID:   pgUUID(uuid.New().String()),
		UserID:    pgUUID(userID.String()),
		Content:   "first",
		Embedding: testEmbedding(),
		Connected: false,
		CreatedAt: pgtype.Timestamptz{Time: t1, Valid: true},
	}
	m2 := Moment{
		ID:        pgUUID(uuid.New().String()),
		TraceID:   pgUUID(uuid.New().String()),
		UserID:    pgUUID(userID.String()),
		Content:   "second",
		Embedding: testEmbedding(),
		Connected: false,
		CreatedAt: pgtype.Timestamptz{Time: t2, Valid: true},
	}
	m3 := Moment{
		ID:        pgUUID(uuid.New().String()),
		TraceID:   pgUUID(uuid.New().String()),
		UserID:    pgUUID(userID.String()),
		Content:   "third",
		Embedding: testEmbedding(),
		Connected: false,
		CreatedAt: pgtype.Timestamptz{Time: t3, Valid: true},
	}

	for _, m := range []Moment{m1, m2, m3} {
		err := q.CreateMoment(context.Background(), CreateMomentParams{
			ID:        m.ID,
			TraceID:   m.TraceID,
			UserID:    m.UserID,
			Content:   m.Content,
			Embedding: m.Embedding,
			Connected: m.Connected,
			CreatedAt: m.CreatedAt,
		})
		if err != nil {
			t.Fatalf("CreateMoment: %v", err)
		}
	}

	// Page from t3 (newest), should get earlier items (t2, t1 for example)
	items, err := q.ListMomentsByUserIDCursor(context.Background(), ListMomentsByUserIDCursorParams{
		UserID:     pgUUID(userID.String()),
		Limit:      2,
		CursorTime: pgtype.Timestamptz{Time: t3, Valid: true},
	})
	if err != nil {
		t.Fatalf("ListMomentsByUserIDCursor: %v", err)
	}
	if len(items) == 0 {
		t.Fatal("expected at least 1 item")
	}

	for _, m := range []Moment{m1, m2, m3} {
		pool.Exec(context.Background(), "DELETE FROM moments WHERE id = $1", m.ID)
	}
}

func TestRandomMomentsByUserID(t *testing.T) {
	pool := testPool(t)
	q := New(pool)

	userID := uuid.New()
	for range 5 {
		createTestMoment(t, q, userID.String())
	}

	items, err := q.RandomMomentsByUserID(context.Background(), RandomMomentsByUserIDParams{
		UserID: pgUUID(userID.String()),
		Limit:  2,
	})
	if err != nil {
		t.Fatalf("RandomMomentsByUserID: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	pool.Exec(context.Background(), "DELETE FROM moments WHERE user_id = $1", pgUUID(userID.String()))
}

func TestCountMomentsByUserID(t *testing.T) {
	pool := testPool(t)
	q := New(pool)

	userID := uuid.New()
	for range 3 {
		createTestMoment(t, q, userID.String())
	}

	count, err := q.CountMomentsByUserID(context.Background(), pgUUID(userID.String()))
	if err != nil {
		t.Fatalf("CountMomentsByUserID: %v", err)
	}
	if count != 3 {
		t.Fatalf("expected count 3, got %d", count)
	}

	pool.Exec(context.Background(), "DELETE FROM moments WHERE user_id = $1", pgUUID(userID.String()))
}

func TestCountMomentsByUserID_Zero(t *testing.T) {
	pool := testPool(t)
	q := New(pool)

	count, err := q.CountMomentsByUserID(context.Background(), pgUUID(uuid.New().String()))
	if err != nil {
		t.Fatalf("CountMomentsByUserID: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected count 0, got %d", count)
	}
}
