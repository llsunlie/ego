package ai

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	platformai "ego-server/internal/platform/ai"
	"ego-server/internal/starmap/domain"
)

func TestConstellationProfileRefiner_RefinesAndEmbedsProfile(t *testing.T) {
	var sawRepresentativeMoment bool
	var embeddingInput string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/chat/completions":
			var body struct {
				Messages []struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"messages"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode chat body: %v", err)
			}
			for _, msg := range body.Messages {
				if strings.Contains(msg.Content, "入职流程突然卡住了") {
					sawRepresentativeMoment = true
				}
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"choices": []map[string]any{
					{"message": map[string]string{"content": `{"topic":"入职适应","summary":"用户持续记录入职阶段的等待和流程不确定。","keywords":["入职","反馈","流程"],"emotions":["焦虑"],"scenes":["工作"],"central_pattern":"新阶段开始前反复等待外部流程确认。","pattern_tags":["新阶段过渡","等待确认"],"theme_label":"入职过渡","theme_description":"包括入职前后的流程等待和反馈不确定，不包括一般工作任务压力。","theme_examples":["等待入职反馈","流程突然卡住"]}`}},
				},
				"usage": map[string]int{"prompt_tokens": 10, "completion_tokens": 20, "total_tokens": 30},
			})
		case "/embeddings":
			var body struct {
				Input string `json:"input"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Fatalf("decode embedding body: %v", err)
			}
			embeddingInput = body.Input
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": []map[string]any{
					{"embedding": []float32{0.1, 0.2, 0.3}},
				},
				"model": "embed-test",
				"usage": map[string]int{"prompt_tokens": 5, "total_tokens": 5},
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	client := platformai.NewClient(platformai.Config{
		EmbeddingAPIKey:  "test-key",
		EmbeddingBaseURL: server.URL,
		EmbeddingModel:   "embed-test",
		ChatAPIKey:       "test-key",
		ChatBaseURL:      server.URL,
		ChatModel:        "chat-test",
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	refiner := NewConstellationProfileRefiner(client)

	result, err := refiner.Refine(context.Background(), domain.ConstellationProfileRefineInput{
		Existing: domain.ConstellationProfile{
			ConstellationID: "constellation-1",
			UserID:          "user-1",
			Topic:           "入职等待",
			Summary:         "用户在等待入职反馈。",
			TraceCount:      2,
			MomentCount:     2,
		},
		RuleMerged: domain.ConstellationProfile{
			ConstellationID: "constellation-1",
			UserID:          "user-1",
			Topic:           "入职过程",
			Summary:         "用户记录入职阶段的等待和流程不确定。",
			TraceCount:      3,
			MomentCount:     3,
			Keywords:        []string{"入职", "反馈"},
			Scenes:          []string{"工作"},
		},
		IncomingTraceProfile: domain.TraceProfile{
			TraceID:     "trace-3",
			UserID:      "user-1",
			Topic:       "入职受阻",
			Summary:     "入职流程突然卡住。",
			Keywords:    []string{"入职", "流程"},
			Scenes:      []string{"工作"},
			PatternTags: []string{"等待确认"},
		},
		RepresentativeMoment: "入职流程突然卡住了。",
		Trigger:              3,
	})
	if err != nil {
		t.Fatalf("Refine() error = %v", err)
	}
	if !sawRepresentativeMoment {
		t.Fatal("expected prompt to include representative moment")
	}
	if result.Profile.Topic != "入职适应" {
		t.Fatalf("topic = %q, want 入职适应", result.Profile.Topic)
	}
	if result.Model != "embed-test" || result.Dim != 3 {
		t.Fatalf("model/dim = %s/%d, want embed-test/3", result.Model, result.Dim)
	}
	if len(result.ProfileEmbedding) != 3 {
		t.Fatalf("embedding = %#v", result.ProfileEmbedding)
	}
	for _, want := range []string{"主题：入职适应", "核心模式：新阶段开始前反复等待外部流程确认。", "模式标签：新阶段过渡，等待确认"} {
		if !strings.Contains(embeddingInput, want) {
			t.Fatalf("embedding input missing %q: %s", want, embeddingInput)
		}
	}
}
