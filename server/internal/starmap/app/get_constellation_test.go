package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"ego-server/internal/starmap/domain"
	writingdomain "ego-server/internal/writing/domain"
)

func TestGetConstellation_Success(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user_id", "user-1")
	now := time.Now()

	constellationRepo := &mockConstellationRepo{
		findByIDFn: func(ctx context.Context, id string) (*domain.Constellation, error) {
			return &domain.Constellation{
				ID:      "c1",
				UserID:  "user-1",
				Name:    "测试星座",
				StarIDs: []string{"s1", "s2"},
				CreatedAt: now,
				UpdatedAt: now,
			}, nil
		},
	}

	starRepo := &mockStarRepo{
		findByIDsFn: func(ctx context.Context, ids []string) ([]domain.Star, error) {
			return []domain.Star{
				{ID: "s1", TraceID: "tr-1"},
				{ID: "s2", TraceID: "tr-2"},
			}, nil
		},
	}

	traceReader := &mockTraceReader{
		listMomentsByTraceIDFn: func(ctx context.Context, traceID string) ([]writingdomain.Moment, error) {
			if traceID == "tr-1" {
				return []writingdomain.Moment{{ID: "mom-1", TraceID: "tr-1"}}, nil
			}
			return []writingdomain.Moment{{ID: "mom-2", TraceID: "tr-2"}}, nil
		},
	}

	uc := NewGetConstellationUseCase(constellationRepo, starRepo, traceReader)
	out, err := uc.Execute(ctx, GetConstellationInput{ConstellationID: "c1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Constellation.ID != "c1" {
		t.Fatalf("expected constellation 'c1', got %q", out.Constellation.ID)
	}
	if len(out.Stars) != 2 {
		t.Fatalf("expected 2 stars, got %d", len(out.Stars))
	}
	if len(out.Moments) != 2 {
		t.Fatalf("expected 2 moments, got %d", len(out.Moments))
	}
}

func TestGetConstellation_NotFound(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user_id", "user-1")

	constellationRepo := &mockConstellationRepo{
		findByIDFn: func(ctx context.Context, id string) (*domain.Constellation, error) {
			return nil, domain.ErrConstellationNotFound
		},
	}

	uc := NewGetConstellationUseCase(constellationRepo, nil, nil)
	_, err := uc.Execute(ctx, GetConstellationInput{ConstellationID: "nonexistent"})
	if !errors.Is(err, domain.ErrConstellationNotFound) {
		t.Fatalf("expected ErrConstellationNotFound, got %v", err)
	}
}

func TestGetConstellation_WrongUser(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user_id", "user-1")

	constellationRepo := &mockConstellationRepo{
		findByIDFn: func(ctx context.Context, id string) (*domain.Constellation, error) {
			return &domain.Constellation{
				ID:     "c1",
				UserID: "user-2", // different user
			}, nil
		},
	}

	uc := NewGetConstellationUseCase(constellationRepo, nil, nil)
	_, err := uc.Execute(ctx, GetConstellationInput{ConstellationID: "c1"})
	if !errors.Is(err, domain.ErrConstellationNotFound) {
		t.Fatalf("expected ErrConstellationNotFound, got %v", err)
	}
}
