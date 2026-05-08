package app

import (
	"context"
	"errors"
	"testing"

	"ego-server/internal/writing/domain"
)

type mockInsightGen struct {
	generateFn func(ctx context.Context, currentContent string, echoMomentID string) (*domain.Insight, error)
}

func (m *mockInsightGen) Generate(ctx context.Context, currentContent string, echoMomentID string) (*domain.Insight, error) {
	return m.generateFn(ctx, currentContent, echoMomentID)
}

func TestGenerateInsight_EmptyContent(t *testing.T) {
	uc := NewGenerateInsightUseCase(nil)
	_, err := uc.Execute(context.Background(), GenerateInsightInput{
		CurrentContent: "",
		EchoMomentID:   "echo-1",
	})
	if !errors.Is(err, domain.ErrEmptyContent) {
		t.Fatalf("expected ErrEmptyContent, got %v", err)
	}
}

func TestGenerateInsight_Success(t *testing.T) {
	insight := &mockInsightGen{
		generateFn: func(ctx context.Context, currentContent string, echoMomentID string) (*domain.Insight, error) {
			if currentContent != "I feel hopeful" {
				t.Fatalf("expected content 'I feel hopeful', got %q", currentContent)
			}
			if echoMomentID != "echo-1" {
				t.Fatalf("expected echoMomentID 'echo-1', got %q", echoMomentID)
			}
			return &domain.Insight{
				ID:               "insight-1",
				Text:             "You seem to be rediscovering hope.",
				RelatedMomentIDs: []string{"echo-1"},
			}, nil
		},
	}

	uc := NewGenerateInsightUseCase(insight)
	output, err := uc.Execute(context.Background(), GenerateInsightInput{
		CurrentContent: "I feel hopeful",
		EchoMomentID:   "echo-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output.ID != "insight-1" {
		t.Fatalf("expected ID 'insight-1', got %q", output.ID)
	}
	if output.Text != "You seem to be rediscovering hope." {
		t.Fatalf("unexpected insight text: %q", output.Text)
	}
}

func TestGenerateInsight_GeneratorError(t *testing.T) {
	insight := &mockInsightGen{
		generateFn: func(ctx context.Context, currentContent string, echoMomentID string) (*domain.Insight, error) {
			return nil, errors.New("AI timeout")
		},
	}

	uc := NewGenerateInsightUseCase(insight)
	_, err := uc.Execute(context.Background(), GenerateInsightInput{
		CurrentContent: "test",
		EchoMomentID:   "echo-1",
	})
	if err == nil {
		t.Fatal("expected error from insight generator")
	}
}
