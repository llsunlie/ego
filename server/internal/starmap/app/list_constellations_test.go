package app

import (
	"context"
	"testing"

	"ego-server/internal/starmap/domain"
)

func TestListConstellations_Empty(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user_id", "user-1")

	constellationRepo := &mockConstellationRepo{
		findAllByUserIDFn: func(ctx context.Context, userID string) ([]domain.Constellation, error) {
			return nil, nil
		},
	}

	uc := NewListConstellationsUseCase(constellationRepo)
	out, err := uc.Execute(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.TotalStarCount != 0 {
		t.Fatalf("expected 0 stars, got %d", out.TotalStarCount)
	}
	if len(out.Constellations) != 0 {
		t.Fatalf("expected 0 constellations, got %d", len(out.Constellations))
	}
}

func TestListConstellations_WithData(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user_id", "user-1")

	constellationRepo := &mockConstellationRepo{
		findAllByUserIDFn: func(ctx context.Context, userID string) ([]domain.Constellation, error) {
			return []domain.Constellation{
				{ID: "c1", StarIDs: []string{"s1", "s2"}},
				{ID: "c2", StarIDs: []string{"s3"}},
			}, nil
		},
	}

	uc := NewListConstellationsUseCase(constellationRepo)
	out, err := uc.Execute(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.TotalStarCount != 3 {
		t.Fatalf("expected 3 stars, got %d", out.TotalStarCount)
	}
	if len(out.Constellations) != 2 {
		t.Fatalf("expected 2 constellations, got %d", len(out.Constellations))
	}
}
