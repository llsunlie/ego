package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"ego-server/internal/starmap/domain"
	writingdomain "ego-server/internal/writing/domain"
)

// --- mocks for stash_trace ---

type mockTraceReader struct {
	getByIDFn              func(ctx context.Context, id string) (*writingdomain.Trace, error)
	listMomentsByTraceIDFn func(ctx context.Context, traceID string) ([]writingdomain.Moment, error)
}

func (m *mockTraceReader) GetTraceByID(ctx context.Context, id string) (*writingdomain.Trace, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockTraceReader) ListMomentsByTraceID(ctx context.Context, traceID string) ([]writingdomain.Moment, error) {
	return m.listMomentsByTraceIDFn(ctx, traceID)
}

type mockTraceStasher struct {
	markStashedFn func(ctx context.Context, traceID string) error
}

func (m *mockTraceStasher) MarkStashed(ctx context.Context, traceID string) error {
	return m.markStashedFn(ctx, traceID)
}

type mockStarRepo struct {
	createFn       func(ctx context.Context, star *domain.Star) error
	findByTraceIDFn func(ctx context.Context, traceID string) (*domain.Star, error)
	findByIDsFn    func(ctx context.Context, ids []string) ([]domain.Star, error)
}

func (m *mockStarRepo) Create(ctx context.Context, star *domain.Star) error {
	return m.createFn(ctx, star)
}
func (m *mockStarRepo) FindByTraceID(ctx context.Context, traceID string) (*domain.Star, error) {
	return m.findByTraceIDFn(ctx, traceID)
}
func (m *mockStarRepo) FindByIDs(ctx context.Context, ids []string) ([]domain.Star, error) {
	return m.findByIDsFn(ctx, ids)
}

type mockConstellationRepo struct {
	createFn          func(ctx context.Context, c *domain.Constellation) error
	updateFn          func(ctx context.Context, c *domain.Constellation) error
	findAllByUserIDFn func(ctx context.Context, userID string) ([]domain.Constellation, error)
	findByIDFn        func(ctx context.Context, id string) (*domain.Constellation, error)
}

func (m *mockConstellationRepo) Create(ctx context.Context, c *domain.Constellation) error {
	return m.createFn(ctx, c)
}
func (m *mockConstellationRepo) Update(ctx context.Context, c *domain.Constellation) error {
	return m.updateFn(ctx, c)
}
func (m *mockConstellationRepo) FindAllByUserID(ctx context.Context, userID string) ([]domain.Constellation, error) {
	return m.findAllByUserIDFn(ctx, userID)
}
func (m *mockConstellationRepo) FindByID(ctx context.Context, id string) (*domain.Constellation, error) {
	return m.findByIDFn(ctx, id)
}
func (m *mockConstellationRepo) FindByStarID(ctx context.Context, starID string) (*domain.Constellation, error) {
	return nil, nil
}

type mockTopicGen struct {
	generateFn func(ctx context.Context, moments []writingdomain.Moment) (string, error)
}

func (m *mockTopicGen) Generate(ctx context.Context, moments []writingdomain.Moment) (string, error) {
	return m.generateFn(ctx, moments)
}

type mockConstellationMat struct {
	findMatchFn func(ctx context.Context, topic string, existing []domain.Constellation) (string, error)
}

func (m *mockConstellationMat) FindMatch(ctx context.Context, topic string, existing []domain.Constellation) (string, error) {
	return m.findMatchFn(ctx, topic, existing)
}

type mockAssetGen struct {
	generateFn func(ctx context.Context, moments []writingdomain.Moment) (string, string, []string, error)
}

func (m *mockAssetGen) Generate(ctx context.Context, moments []writingdomain.Moment) (string, string, []string, error) {
	return m.generateFn(ctx, moments)
}

type mockIDGen struct {
	id string
}

func (m *mockIDGen) New() string { return m.id }

// --- tests ---

func TestStashTrace_Success(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user_id", "user-1")
	now := time.Now()

	trace := &writingdomain.Trace{
		ID: "tr-1", UserID: "user-1", Motivation: "direct",
		Stashed: false, CreatedAt: now,
	}
	moments := []writingdomain.Moment{
		{ID: "mom-1", TraceID: "tr-1", UserID: "user-1", Content: "一些内容", CreatedAt: now},
	}

	starCreated := false
	traceStashed := false
	constellationCreated := false

	traceReader := &mockTraceReader{
		getByIDFn: func(ctx context.Context, id string) (*writingdomain.Trace, error) {
			return trace, nil
		},
		listMomentsByTraceIDFn: func(ctx context.Context, traceID string) ([]writingdomain.Moment, error) {
			return moments, nil
		},
	}

	traceStasher := &mockTraceStasher{
		markStashedFn: func(ctx context.Context, traceID string) error {
			traceStashed = true
			return nil
		},
	}

	starRepo := &mockStarRepo{
		createFn: func(ctx context.Context, star *domain.Star) error {
			starCreated = true
			if star.TraceID != "tr-1" {
				t.Errorf("expected traceID 'tr-1', got %q", star.TraceID)
			}
			return nil
		},
	}

	constellationRepo := &mockConstellationRepo{
		findAllByUserIDFn: func(ctx context.Context, userID string) ([]domain.Constellation, error) {
			return nil, nil
		},
		createFn: func(ctx context.Context, c *domain.Constellation) error {
			constellationCreated = true
			if len(c.StarIDs) != 1 {
				t.Errorf("expected 1 star in constellation, got %d", len(c.StarIDs))
			}
			return nil
		},
	}

	topicGen := &mockTopicGen{
		generateFn: func(ctx context.Context, moments []writingdomain.Moment) (string, error) {
			return "关于一些内容…", nil
		},
	}

	constellationMat := &mockConstellationMat{
		findMatchFn: func(ctx context.Context, topic string, existing []domain.Constellation) (string, error) {
			return "", nil // no match → lone-star constellation
		},
	}

	assetGen := &mockAssetGen{
		generateFn: func(ctx context.Context, moments []writingdomain.Moment) (string, string, []string, error) {
			return "测试星座", "一些洞察", []string{"提示1", "提示2"}, nil
		},
	}

	idGen := &mockIDGen{id: "star-1"}

	uc := NewStashTraceUseCase(
		traceReader, traceStasher, starRepo, constellationRepo,
		topicGen, constellationMat, assetGen, idGen,
	)

	star, err := uc.Execute(ctx, StashTraceInput{TraceID: "tr-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if star.ID != "star-1" {
		t.Fatalf("expected star id 'star-1', got %q", star.ID)
	}
	if !starCreated {
		t.Fatal("star was not created")
	}
	if !traceStashed {
		t.Fatal("trace was not marked stashed")
	}
	if !constellationCreated {
		t.Fatal("constellation was not created")
	}
}

func TestStashTrace_TraceNotFound(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user_id", "user-1")

	traceReader := &mockTraceReader{
		getByIDFn: func(ctx context.Context, id string) (*writingdomain.Trace, error) {
			return nil, writingdomain.ErrTraceNotFound
		},
	}

	uc := NewStashTraceUseCase(
		traceReader, nil, nil, nil, nil, nil, nil, nil,
	)

	_, err := uc.Execute(ctx, StashTraceInput{TraceID: "nonexistent"})
	if !errors.Is(err, writingdomain.ErrTraceNotFound) {
		t.Fatalf("expected ErrTraceNotFound, got %v", err)
	}
}

func TestStashTrace_AlreadyStashed(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user_id", "user-1")
	now := time.Now()

	trace := &writingdomain.Trace{
		ID: "tr-1", UserID: "user-1", Stashed: true, CreatedAt: now,
	}

	traceReader := &mockTraceReader{
		getByIDFn: func(ctx context.Context, id string) (*writingdomain.Trace, error) {
			return trace, nil
		},
	}

	uc := NewStashTraceUseCase(
		traceReader, nil, nil, nil, nil, nil, nil, nil,
	)

	_, err := uc.Execute(ctx, StashTraceInput{TraceID: "tr-1"})
	if !errors.Is(err, domain.ErrTraceAlreadyStashed) {
		t.Fatalf("expected ErrTraceAlreadyStashed, got %v", err)
	}
}

func TestStashTrace_WrongUser(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user_id", "user-1")
	now := time.Now()

	trace := &writingdomain.Trace{
		ID: "tr-1", UserID: "user-2", Stashed: false, CreatedAt: now,
	}

	traceReader := &mockTraceReader{
		getByIDFn: func(ctx context.Context, id string) (*writingdomain.Trace, error) {
			return trace, nil
		},
	}

	uc := NewStashTraceUseCase(
		traceReader, nil, nil, nil, nil, nil, nil, nil,
	)

	_, err := uc.Execute(ctx, StashTraceInput{TraceID: "tr-1"})
	if !errors.Is(err, domain.ErrTraceNotFound) {
		t.Fatalf("expected ErrTraceNotFound, got %v", err)
	}
}
