package ai

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"ego-server/internal/writing/domain"
)

type stubMomentRepo struct {
	getByIDFn func(ctx context.Context, id string) (*domain.Moment, error)
}

func (m *stubMomentRepo) Create(ctx context.Context, moment *domain.Moment) error { return nil }
func (m *stubMomentRepo) GetByID(ctx context.Context, id string) (*domain.Moment, error) {
	return m.getByIDFn(ctx, id)
}
func (m *stubMomentRepo) ListByTraceID(ctx context.Context, traceID string) ([]domain.Moment, error) {
	return nil, nil
}
func (m *stubMomentRepo) ListByUserID(ctx context.Context, userID string) ([]domain.Moment, error) {
	return nil, nil
}

type stubEchoRepo struct {
	findByMomentIDFn func(ctx context.Context, momentID string) (*domain.Echo, error)
}

func (e *stubEchoRepo) Create(ctx context.Context, echo *domain.Echo) error { return nil }
func (e *stubEchoRepo) FindByMomentID(ctx context.Context, momentID string) (*domain.Echo, error) {
	return e.findByMomentIDFn(ctx, momentID)
}

func TestInsightGenerator_Success(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{"content": "你在反复寻找某种确定性，却忽略了已经拥有的答案。"}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	momentRepo := &stubMomentRepo{
		getByIDFn: func(ctx context.Context, id string) (*domain.Moment, error) {
			return &domain.Moment{ID: "moment-1", Content: "今天又陷入了迷茫"}, nil
		},
	}
	echoRepo := &stubEchoRepo{
		findByMomentIDFn: func(ctx context.Context, momentID string) (*domain.Echo, error) {
			return &domain.Echo{MatchedMomentIDs: []string{"old-1", "old-2"}}, nil
		},
	}

	gen := NewInsightGenerator(client, momentRepo, echoRepo)
	insight, err := gen.Generate(context.Background(), "moment-1", "echo-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if insight.Text != "你在反复寻找某种确定性，却忽略了已经拥有的答案。" {
		t.Fatalf("unexpected insight text: %q", insight.Text)
	}
	if len(insight.RelatedMomentIDs) != 2 {
		t.Fatalf("expected 2 related moment IDs, got %d", len(insight.RelatedMomentIDs))
	}
}

func TestInsightGenerator_Success_NoEcho(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{"content": "你总是在独自面对问题，这未必是坚强。"}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	momentRepo := &stubMomentRepo{
		getByIDFn: func(ctx context.Context, id string) (*domain.Moment, error) {
			return &domain.Moment{ID: "moment-1", Content: "test"}, nil
		},
	}
	echoRepo := &stubEchoRepo{
		findByMomentIDFn: func(ctx context.Context, momentID string) (*domain.Echo, error) {
			return nil, errors.New("echo not found")
		},
	}

	gen := NewInsightGenerator(client, momentRepo, echoRepo)
	insight, err := gen.Generate(context.Background(), "moment-1", "echo-1")
	if err != nil {
		t.Fatalf("unexpected error (should tolerate missing echo): %v", err)
	}
	if insight.Text != "你总是在独自面对问题，这未必是坚强。" {
		t.Fatalf("unexpected insight text: %q", insight.Text)
	}
	if len(insight.RelatedMomentIDs) != 0 {
		t.Fatalf("expected empty related moment IDs when no echo, got %v", insight.RelatedMomentIDs)
	}
}

func TestInsightGenerator_MomentNotFound(t *testing.T) {
	momentRepo := &stubMomentRepo{
		getByIDFn: func(ctx context.Context, id string) (*domain.Moment, error) {
			return nil, domain.ErrMomentNotFound
		},
	}
	echoRepo := &stubEchoRepo{
		findByMomentIDFn: func(ctx context.Context, momentID string) (*domain.Echo, error) {
			return nil, nil
		},
	}

	gen := NewInsightGenerator(nil, momentRepo, echoRepo)
	_, err := gen.Generate(context.Background(), "moment-1", "echo-1")
	if err == nil {
		t.Fatal("expected error when moment not found")
	}
}

func TestInsightGenerator_ChatAPIError(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	})

	momentRepo := &stubMomentRepo{
		getByIDFn: func(ctx context.Context, id string) (*domain.Moment, error) {
			return &domain.Moment{ID: "moment-1", Content: "test"}, nil
		},
	}
	echoRepo := &stubEchoRepo{
		findByMomentIDFn: func(ctx context.Context, momentID string) (*domain.Echo, error) {
			return &domain.Echo{}, nil
		},
	}

	gen := NewInsightGenerator(client, momentRepo, echoRepo)
	_, err := gen.Generate(context.Background(), "moment-1", "echo-1")
	if err == nil {
		t.Fatal("expected error when chat API fails")
	}
}

func TestInsightGenerator_TrimsWhitespace(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{"content": "\n  你在用忙碌逃避某个不愿面对的问题。  \n"}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	momentRepo := &stubMomentRepo{
		getByIDFn: func(ctx context.Context, id string) (*domain.Moment, error) {
			return &domain.Moment{ID: "moment-1", Content: "test"}, nil
		},
	}
	echoRepo := &stubEchoRepo{
		findByMomentIDFn: func(ctx context.Context, momentID string) (*domain.Echo, error) {
			return nil, nil
		},
	}

	gen := NewInsightGenerator(client, momentRepo, echoRepo)
	insight, err := gen.Generate(context.Background(), "moment-1", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(insight.Text) != insight.Text {
		t.Fatalf("expected trimmed insight text, got %q", insight.Text)
	}
}

func TestBuildInsightUserPrompt_WithEcho(t *testing.T) {
	moment := &domain.Moment{Content: "今天感到特别焦虑"}
	echo := &domain.Echo{MatchedMomentIDs: []string{"a", "b", "c"}}

	result := buildInsightUserPrompt(moment, echo)

	if !strings.Contains(result, "当前想法：今天感到特别焦虑") {
		t.Fatalf("expected moment content in prompt, got %q", result)
	}
	if !strings.Contains(result, "3") {
		t.Fatalf("expected matched count in prompt, got %q", result)
	}
}

func TestBuildInsightUserPrompt_NoEcho(t *testing.T) {
	moment := &domain.Moment{Content: "只是随便写写"}

	result := buildInsightUserPrompt(moment, nil)

	if result != "当前想法：只是随便写写" {
		t.Fatalf("expected only moment content, got %q", result)
	}
}
