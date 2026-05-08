package grpc

import (
	"context"
	"errors"
	"testing"

	"ego-server/internal/writing/app"
	"ego-server/internal/writing/domain"

	pb "ego-server/proto/ego"

	"github.com/google/uuid"
)

func TestHandler_CreateMoment_Delegation(t *testing.T) {
	// Verify Handler is constructible and passes through errors.
	// Full integration is tested via the app layer.

	// Create a use case that will fail with a known error.
	// We can't easily construct a real use case without all deps,
	// so test via the NewHandler constructor only.
	h := NewHandler(nil, nil)
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
	if h.createMoment != nil || h.generateInsight != nil {
		t.Fatal("expected nil use cases from constructor")
	}
}

func TestHandler_CreateMoment_ErrorPropagation(t *testing.T) {
	// Use a use case that returns an error via a real but incomplete setup.
	// Empty content triggers ErrEmptyContent.
	uc := app.NewCreateMomentUseCase(nil, nil, nil, nil, &stubIDGen{})
	h := NewHandler(uc, nil)

	ctx := context.WithValue(context.Background(), "user_id", "user-1")
	_, err := h.CreateMoment(ctx, &pb.CreateMomentReq{Content: ""})
	if err == nil {
		t.Fatal("expected error for empty content")
	}
}

func TestHandler_GenerateInsight_ErrorPropagation(t *testing.T) {
	uc := app.NewGenerateInsightUseCase(nil)
	h := NewHandler(nil, uc)

	_, err := h.GenerateInsight(context.Background(), &pb.GenerateInsightReq{CurrentContent: ""})
	if err == nil {
		t.Fatal("expected error for empty content")
	}
}

func TestHandler_GenerateInsight_Success(t *testing.T) {
	uc := app.NewGenerateInsightUseCase(&stubInsightGenerator{})
	h := NewHandler(nil, uc)

	res, err := h.GenerateInsight(context.Background(), &pb.GenerateInsightReq{
		CurrentContent: "I feel better today",
		EchoMomentId:   "echo-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Insight == nil {
		t.Fatal("expected non-nil insight")
	}
	if res.Insight.Id != "ins-1" {
		t.Fatalf("expected Id 'ins-1', got %q", res.Insight.Id)
	}
}

// --- stubs for handler tests ---

type stubIDGen struct{}

func (s *stubIDGen) New() string { return uuid.NewString() }

type stubInsightGenerator struct{}

func (s *stubInsightGenerator) Generate(ctx context.Context, currentContent string, echoMomentID string) (*domain.Insight, error) {
	if currentContent == "" {
		return nil, errors.New("empty content")
	}
	return &domain.Insight{
		ID:               "ins-1",
		Text:             "You are making progress.",
		RelatedMomentIDs: []string{"echo-1"},
	}, nil
}
