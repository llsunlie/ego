package app

import (
	"context"
	"errors"
	"testing"

	"ego-server/internal/writing/domain"
)

type mockInsightRepo struct {
	createFn         func(ctx context.Context, insight *domain.Insight) error
	findByMomentIDFn func(ctx context.Context, momentID string) (*domain.Insight, error)
}

func (m *mockInsightRepo) Create(ctx context.Context, insight *domain.Insight) error {
	return m.createFn(ctx, insight)
}
func (m *mockInsightRepo) FindByMomentID(ctx context.Context, momentID string) (*domain.Insight, error) {
	return m.findByMomentIDFn(ctx, momentID)
}

type mockInsightGen struct {
	generateFn func(ctx context.Context, momentID string, echoID string) (*domain.Insight, error)
}

func (m *mockInsightGen) Generate(ctx context.Context, momentID string, echoID string) (*domain.Insight, error) {
	return m.generateFn(ctx, momentID, echoID)
}

func TestGenerateInsight_EmptyMomentID(t *testing.T) {
	uc := NewGenerateInsightUseCase(nil, nil, &mockIDGen{id: "x"})
	_, err := uc.Execute(context.Background(), GenerateInsightInput{
		MomentID: "",
		EchoID:   "echo-1",
	})
	if !errors.Is(err, domain.ErrEmptyContent) {
		t.Fatalf("expected ErrEmptyContent, got %v", err)
	}
}

func TestGenerateInsight_Success(t *testing.T) {
	insightRepo := &mockInsightRepo{
		createFn: func(ctx context.Context, insight *domain.Insight) error { return nil },
	}

	insightGen := &mockInsightGen{
		generateFn: func(ctx context.Context, momentID string, echoID string) (*domain.Insight, error) {
			if momentID != "moment-1" {
				t.Fatalf("expected momentID 'moment-1', got %q", momentID)
			}
			if echoID != "echo-1" {
				t.Fatalf("expected echoID 'echo-1', got %q", echoID)
			}
			return &domain.Insight{
				ID:               "insight-1",
				Text:             "You seem to be rediscovering hope.",
				RelatedMomentIDs: []string{"echo-1"},
				MomentID:         "moment-1",
				EchoID:           "echo-1",
			}, nil
		},
	}

	uc := NewGenerateInsightUseCase(insightRepo, insightGen, &mockIDGen{id: "id-1"})
	output, err := uc.Execute(context.Background(), GenerateInsightInput{
		MomentID: "moment-1",
		EchoID:   "echo-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output.ID != "id-1" {
		t.Fatalf("expected ID 'id-1' (from IDGen), got %q", output.ID)
	}
	if output.Text != "You seem to be rediscovering hope." {
		t.Fatalf("unexpected insight text: %q", output.Text)
	}
	if output.MomentID != "moment-1" {
		t.Fatalf("expected MomentID 'moment-1', got %q", output.MomentID)
	}
}

func TestGenerateInsight_GeneratorError(t *testing.T) {
	insightRepo := &mockInsightRepo{
		createFn: func(ctx context.Context, insight *domain.Insight) error { return nil },
	}

	insightGen := &mockInsightGen{
		generateFn: func(ctx context.Context, momentID string, echoID string) (*domain.Insight, error) {
			return nil, errors.New("AI timeout")
		},
	}

	ids := &mockIDGen{id: "x"}

	uc := NewGenerateInsightUseCase(insightRepo, insightGen, ids)
	_, err := uc.Execute(context.Background(), GenerateInsightInput{
		MomentID: "moment-1",
		EchoID:   "echo-1",
	})
	if err == nil {
		t.Fatal("expected error from insight generator")
	}
}
