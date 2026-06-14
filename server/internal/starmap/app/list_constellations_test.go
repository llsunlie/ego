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
	starRepo := &mockStarRepo{
		findAllByUserIDFn: func(ctx context.Context, userID string) ([]domain.Star, error) {
			return nil, nil
		},
	}

	uc := NewListConstellationsUseCase(constellationRepo, starRepo)
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
	starRepo := &mockStarRepo{
		findAllByUserIDFn: func(ctx context.Context, userID string) ([]domain.Star, error) {
			return []domain.Star{
				{ID: "s1", UserID: "user-1"},
				{ID: "s2", UserID: "user-1"},
				{ID: "s3", UserID: "user-1"},
			}, nil
		},
	}

	uc := NewListConstellationsUseCase(constellationRepo, starRepo)
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

func TestListConstellations_CountsMultiMembershipStarsOnce(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user_id", "user-1")

	constellationRepo := &mockConstellationRepo{
		findAllByUserIDFn: func(ctx context.Context, userID string) ([]domain.Constellation, error) {
			return []domain.Constellation{
				{ID: "c1", StarIDs: []string{"s1", "s2"}},
				{ID: "c2", StarIDs: []string{"s1", "s3"}},
			}, nil
		},
	}
	starRepo := &mockStarRepo{
		findAllByUserIDFn: func(ctx context.Context, userID string) ([]domain.Star, error) {
			return []domain.Star{
				{ID: "s1", UserID: "user-1"},
				{ID: "s2", UserID: "user-1"},
				{ID: "s3", UserID: "user-1"},
			}, nil
		},
	}

	uc := NewListConstellationsUseCase(constellationRepo, starRepo)
	out, err := uc.Execute(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.TotalStarCount != 3 {
		t.Fatalf("expected 3 unique stars, got %d", out.TotalStarCount)
	}
	if len(out.Constellations) != 2 {
		t.Fatalf("expected 2 constellations, got %d", len(out.Constellations))
	}
}

func TestListConstellations_WithUnclusteredStars(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user_id", "user-1")

	constellationRepo := &mockConstellationRepo{
		findAllByUserIDFn: func(ctx context.Context, userID string) ([]domain.Constellation, error) {
			return []domain.Constellation{
				{ID: "c1", StarIDs: []string{"s1"}},
			}, nil
		},
	}
	starRepo := &mockStarRepo{
		findAllByUserIDFn: func(ctx context.Context, userID string) ([]domain.Star, error) {
			return []domain.Star{
				{ID: "s1", UserID: "user-1", Topic: "已聚合"},
				{ID: "s2", UserID: "user-1", Topic: "聚合中"},
			}, nil
		},
	}

	uc := NewListConstellationsUseCase(constellationRepo, starRepo)
	out, err := uc.Execute(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.TotalStarCount != 2 {
		t.Fatalf("expected 2 total stars, got %d", out.TotalStarCount)
	}
	// 1 real constellation + 1 fake for unclustered star
	if len(out.Constellations) != 2 {
		t.Fatalf("expected 2 constellations (1 real + 1 fake), got %d", len(out.Constellations))
	}
	// Find the fake constellation for s2
	var found bool
	for _, c := range out.Constellations {
		if c.ID == "s2" && c.Name == "聚合中" && len(c.StarIDs) == 1 && c.StarIDs[0] == "s2" {
			found = true
		}
	}
	if !found {
		t.Fatal("expected a fake constellation wrapping unclustered star s2")
	}
}
