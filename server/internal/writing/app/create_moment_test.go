package app

import (
	"context"
	"errors"
	"testing"

	"ego-server/internal/writing/domain"
)

// --- Mock implementations ---

type mockTraceRepo struct {
	createFn  func(ctx context.Context, trace *domain.Trace) error
	getByIDFn func(ctx context.Context, id string) (*domain.Trace, error)
	deleteFn  func(ctx context.Context, id string) error
}

func (m *mockTraceRepo) Create(ctx context.Context, trace *domain.Trace) error { return m.createFn(ctx, trace) }
func (m *mockTraceRepo) GetByID(ctx context.Context, id string) (*domain.Trace, error) { return m.getByIDFn(ctx, id) }
func (m *mockTraceRepo) Update(ctx context.Context, trace *domain.Trace) error { return nil }
func (m *mockTraceRepo) Delete(ctx context.Context, id string) error           { return m.deleteFn(ctx, id) }

type mockMomentRepo struct {
	createFn       func(ctx context.Context, moment *domain.Moment) error
	getByIDFn      func(ctx context.Context, id string) (*domain.Moment, error)
	listByTraceFn  func(ctx context.Context, traceID string) ([]domain.Moment, error)
	listByUserFn   func(ctx context.Context, userID string) ([]domain.Moment, error)
}

func (m *mockMomentRepo) Create(ctx context.Context, moment *domain.Moment) error { return m.createFn(ctx, moment) }
func (m *mockMomentRepo) GetByID(ctx context.Context, id string) (*domain.Moment, error) { return m.getByIDFn(ctx, id) }
func (m *mockMomentRepo) ListByTraceID(ctx context.Context, traceID string) ([]domain.Moment, error) { return m.listByTraceFn(ctx, traceID) }
func (m *mockMomentRepo) ListByUserID(ctx context.Context, userID string) ([]domain.Moment, error) { return m.listByUserFn(ctx, userID) }

type mockEchoRepo struct {
	createFn         func(ctx context.Context, echo *domain.Echo) error
	findByMomentIDFn func(ctx context.Context, momentID string) (*domain.Echo, error)
}

func (m *mockEchoRepo) Create(ctx context.Context, echo *domain.Echo) error { return m.createFn(ctx, echo) }
func (m *mockEchoRepo) FindByMomentID(ctx context.Context, momentID string) (*domain.Echo, error) {
	return m.findByMomentIDFn(ctx, momentID)
}

type mockEmbeddingGen struct {
	generateFn func(ctx context.Context, content string) ([]domain.EmbeddingEntry, error)
}

func (m *mockEmbeddingGen) Generate(ctx context.Context, content string) ([]domain.EmbeddingEntry, error) {
	return m.generateFn(ctx, content)
}

type mockEchoMatcher struct {
	matchFn func(ctx context.Context, current *domain.Moment, history []domain.Moment) ([]domain.MatchedMoment, error)
}

func (m *mockEchoMatcher) Match(ctx context.Context, current *domain.Moment, history []domain.Moment) ([]domain.MatchedMoment, error) {
	return m.matchFn(ctx, current, history)
}

type mockIDGen struct {
	id string
}

func (m *mockIDGen) New() string { return m.id }

func withUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, "user_id", userID)
}

// --- Tests ---

func TestCreateMoment_EmptyContent(t *testing.T) {
	uc := NewCreateMomentUseCase(nil, nil, nil, nil, nil, nil)
	_, err := uc.Execute(withUserID(context.Background(), "user-1"), CreateMomentInput{Content: ""})
	if !errors.Is(err, domain.ErrEmptyContent) {
		t.Fatalf("expected ErrEmptyContent, got %v", err)
	}
}

func TestCreateMoment_MissingUserID(t *testing.T) {
	uc := NewCreateMomentUseCase(nil, nil, nil, nil, nil, nil)
	_, err := uc.Execute(context.Background(), CreateMomentInput{Content: "hello"})
	if err == nil {
		t.Fatal("expected error for missing user_id")
	}
}

func TestCreateMoment_NewTrace(t *testing.T) {
	traceID := "trace-1"
	momentID := "moment-1"
	userID := "user-1"

	traces := &mockTraceRepo{
		createFn: func(ctx context.Context, trace *domain.Trace) error {
			if trace.UserID != userID {
				t.Fatalf("expected UserID %s, got %s", userID, trace.UserID)
			}
			trace.ID = traceID
			return nil
		},
		deleteFn: func(ctx context.Context, id string) error { return nil },
	}

	moments := &mockMomentRepo{
		createFn: func(ctx context.Context, moment *domain.Moment) error {
			moment.ID = momentID
			return nil
		},
		listByUserFn: func(ctx context.Context, id string) ([]domain.Moment, error) {
			return []domain.Moment{{ID: "old-1"}}, nil
		},
	}

	echos := &mockEchoRepo{
		createFn: func(ctx context.Context, echo *domain.Echo) error { return nil },
	}

	embedding := &mockEmbeddingGen{
		generateFn: func(ctx context.Context, content string) ([]domain.EmbeddingEntry, error) {
			return []domain.EmbeddingEntry{{Model: "test", Embedding: []float32{0.1, 0.2}}}, nil
		},
	}

	echo := &mockEchoMatcher{
		matchFn: func(ctx context.Context, current *domain.Moment, history []domain.Moment) ([]domain.MatchedMoment, error) {
			return []domain.MatchedMoment{{MomentID: "old-1", Similarity: 0.8}}, nil
		},
	}

	ids := &mockIDGen{id: "id-seq"}

	uc := NewCreateMomentUseCase(traces, moments, echos, embedding, echo, ids)
	output, err := uc.Execute(withUserID(context.Background(), userID), CreateMomentInput{
		Content:    "first moment",
		Motivation: "direct",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output.Moment.ID != momentID {
		t.Fatalf("expected Moment.ID %s, got %s", momentID, output.Moment.ID)
	}
	if output.Echo == nil {
		t.Fatal("expected non-nil Echo")
	}
	if output.Echo.MatchedMomentIDs[0] != "old-1" {
		t.Fatalf("expected MatchedMomentIDs[0] 'old-1', got %q", output.Echo.MatchedMomentIDs[0])
	}
}

func TestCreateMoment_ExistingTrace(t *testing.T) {
	userID := "user-1"

	traces := &mockTraceRepo{
		getByIDFn: func(ctx context.Context, id string) (*domain.Trace, error) {
			return &domain.Trace{ID: id, UserID: userID}, nil
		},
	}

	moments := &mockMomentRepo{
		createFn: func(ctx context.Context, moment *domain.Moment) error { return nil },
		listByUserFn: func(ctx context.Context, id string) ([]domain.Moment, error) {
			return []domain.Moment{{ID: "old-1"}, {ID: "old-2"}}, nil
		},
	}

	echos := &mockEchoRepo{
		createFn: func(ctx context.Context, echo *domain.Echo) error { return nil },
	}

	embedding := &mockEmbeddingGen{
		generateFn: func(ctx context.Context, content string) ([]domain.EmbeddingEntry, error) {
			return []domain.EmbeddingEntry{{Model: "test", Embedding: []float32{0.1}}}, nil
		},
	}

	echo := &mockEchoMatcher{
		matchFn: func(ctx context.Context, current *domain.Moment, history []domain.Moment) ([]domain.MatchedMoment, error) {
			if len(history) != 2 {
				t.Fatalf("expected 2 history moments, got %d", len(history))
			}
			return []domain.MatchedMoment{{MomentID: "old-1", Similarity: 0.7}}, nil
		},
	}

	ids := &mockIDGen{id: "seq-1"}

	uc := NewCreateMomentUseCase(traces, moments, echos, embedding, echo, ids)
	output, err := uc.Execute(withUserID(context.Background(), userID), CreateMomentInput{
		Content: "continuing",
		TraceID: "existing-trace",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output.Moment.TraceID != "existing-trace" {
		t.Fatalf("expected TraceID existing-trace, got %s", output.Moment.TraceID)
	}
}

func TestCreateMoment_TraceNotFound(t *testing.T) {
	userID := "user-1"

	traces := &mockTraceRepo{
		getByIDFn: func(ctx context.Context, id string) (*domain.Trace, error) {
			return nil, domain.ErrTraceNotFound
		},
	}

	uc := NewCreateMomentUseCase(traces, nil, nil, nil, nil, &mockIDGen{id: "x"})
	_, err := uc.Execute(withUserID(context.Background(), userID), CreateMomentInput{
		Content: "hello",
		TraceID: "nonexistent",
	})
	if !errors.Is(err, domain.ErrTraceNotFound) {
		t.Fatalf("expected ErrTraceNotFound, got %v", err)
	}
}

func TestCreateMoment_TraceNotOwnedByUser(t *testing.T) {
	userID := "user-1"

	traces := &mockTraceRepo{
		getByIDFn: func(ctx context.Context, id string) (*domain.Trace, error) {
			return &domain.Trace{ID: id, UserID: "other-user"}, nil
		},
	}

	uc := NewCreateMomentUseCase(traces, nil, nil, nil, nil, &mockIDGen{id: "x"})
	_, err := uc.Execute(withUserID(context.Background(), userID), CreateMomentInput{
		Content: "hello",
		TraceID: "trace-1",
	})
	if err == nil {
		t.Fatal("expected error for trace belonging to another user")
	}
}

func TestCreateMoment_EmbeddingFailure_RollsBackTrace(t *testing.T) {
	userID := "user-1"
	deleteCalled := false
	traceID := "trace-1"

	traces := &mockTraceRepo{
		createFn: func(ctx context.Context, trace *domain.Trace) error {
			trace.ID = traceID
			return nil
		},
		deleteFn: func(ctx context.Context, id string) error {
			if id == traceID {
				deleteCalled = true
			}
			return nil
		},
	}

	embedding := &mockEmbeddingGen{
		generateFn: func(ctx context.Context, content string) ([]domain.EmbeddingEntry, error) {
			return nil, errors.New("AI service unavailable")
		},
	}

	ids := &mockIDGen{id: "seq-1"}

	uc := NewCreateMomentUseCase(traces, nil, nil, embedding, nil, ids)
	_, err := uc.Execute(withUserID(context.Background(), userID), CreateMomentInput{
		Content: "hello",
	})
	if err == nil {
		t.Fatal("expected error for embedding failure")
	}
	if !deleteCalled {
		t.Fatal("expected trace to be deleted on embedding failure rollback")
	}
}

func TestCreateMoment_EchoWithNoHistory(t *testing.T) {
	userID := "user-1"

	traces := &mockTraceRepo{
		createFn: func(ctx context.Context, trace *domain.Trace) error { return nil },
		deleteFn: func(ctx context.Context, id string) error { return nil },
	}

	moments := &mockMomentRepo{
		createFn: func(ctx context.Context, moment *domain.Moment) error { return nil },
		listByUserFn: func(ctx context.Context, id string) ([]domain.Moment, error) {
			return nil, nil
		},
	}

	embedding := &mockEmbeddingGen{
		generateFn: func(ctx context.Context, content string) ([]domain.EmbeddingEntry, error) {
			return []domain.EmbeddingEntry{{Model: "test", Embedding: []float32{0.1}}}, nil
		},
	}

	ids := &mockIDGen{id: "s"}

	uc := NewCreateMomentUseCase(traces, moments, nil, embedding, nil, ids)
	output, err := uc.Execute(withUserID(context.Background(), userID), CreateMomentInput{
		Content: "first ever",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output.Echo != nil {
		t.Fatal("expected nil echo when no history exists")
	}
}

func TestCreateMoment_EchoMatchingError(t *testing.T) {
	userID := "user-1"

	traces := &mockTraceRepo{
		createFn: func(ctx context.Context, trace *domain.Trace) error { return nil },
		deleteFn: func(ctx context.Context, id string) error { return nil },
	}

	moments := &mockMomentRepo{
		createFn: func(ctx context.Context, moment *domain.Moment) error { return nil },
		listByUserFn: func(ctx context.Context, id string) ([]domain.Moment, error) {
			return []domain.Moment{{ID: "old-1"}}, nil
		},
	}

	embedding := &mockEmbeddingGen{
		generateFn: func(ctx context.Context, content string) ([]domain.EmbeddingEntry, error) {
			return []domain.EmbeddingEntry{{Model: "test", Embedding: []float32{0.1}}}, nil
		},
	}

	echo := &mockEchoMatcher{
		matchFn: func(ctx context.Context, current *domain.Moment, history []domain.Moment) ([]domain.MatchedMoment, error) {
			return nil, errors.New("match failed")
		},
	}

	ids := &mockIDGen{id: "s"}

	uc := NewCreateMomentUseCase(traces, moments, nil, embedding, echo, ids)
	_, err := uc.Execute(withUserID(context.Background(), userID), CreateMomentInput{
		Content: "test",
	})
	if err == nil {
		t.Fatal("expected error when echo matching fails")
	}
}

func TestExcludeSelf(t *testing.T) {
	moments := []domain.Moment{
		{ID: "a"},
		{ID: "b"},
		{ID: "c"},
	}
	result := excludeSelf(moments, "b")
	if len(result) != 2 {
		t.Fatalf("expected 2, got %d", len(result))
	}
	for _, m := range result {
		if m.ID == "b" {
			t.Fatal("expected 'b' to be excluded")
		}
	}
}
