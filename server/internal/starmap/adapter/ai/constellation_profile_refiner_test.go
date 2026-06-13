package ai

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
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
				if strings.Contains(msg.Content, "样例流程在节点B停住了") {
					sawRepresentativeMoment = true
				}
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"choices": []map[string]any{
					{"message": map[string]string{"content": `{"topic":"样例流程整理","display_name":"节点停住","summary":"测试主体持续记录样例流程中的节点确认状态。","keywords":["样例流程","节点确认","流程"],"emotions":["焦虑"],"scenes":["测试场景"],"central_pattern":"流程推进前反复等待外部节点确认。","pattern_tags":["流程过渡","节点确认"],"theme_label":"流程过渡","theme_description":"包括样例流程中的等待和确认状态，不包括一般任务压力。","theme_examples":["等待节点确认","流程节点停住"]}`}},
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
			Topic:           "样例等待",
			Summary:         "测试主体在等待节点确认。",
			TraceCount:      2,
			MomentCount:     2,
		},
		RuleMerged: domain.ConstellationProfile{
			ConstellationID: "constellation-1",
			UserID:          "user-1",
			Topic:           "样例流程",
			Summary:         "测试主体记录样例流程的等待和不确定。",
			TraceCount:      3,
			MomentCount:     3,
			Keywords:        []string{"样例流程", "节点确认"},
			Scenes:          []string{"测试场景"},
		},
		IncomingTraceProfile: domain.TraceProfile{
			TraceID:     "trace-3",
			UserID:      "user-1",
			Topic:       "流程节点停住",
			Summary:     "样例流程在节点B停住。",
			Keywords:    []string{"样例流程", "流程"},
			Scenes:      []string{"测试场景"},
			PatternTags: []string{"节点确认"},
		},
		RepresentativeMoment: "样例流程在节点B停住了。",
		Trigger:              3,
	})
	if err != nil {
		t.Fatalf("Refine() error = %v", err)
	}
	if !sawRepresentativeMoment {
		t.Fatal("expected prompt to include representative moment")
	}
	if result.Profile.Topic != "样例流程整理" {
		t.Fatalf("topic = %q, want 样例流程整理", result.Profile.Topic)
	}
	if result.DisplayName != "节点停住" {
		t.Fatalf("display name = %q, want 节点停住", result.DisplayName)
	}
	if result.Model != "embed-test" || result.Dim != 3 {
		t.Fatalf("model/dim = %s/%d, want embed-test/3", result.Model, result.Dim)
	}
	if len(result.ProfileEmbedding) != 3 {
		t.Fatalf("embedding = %#v", result.ProfileEmbedding)
	}
	for _, want := range []string{"主题：样例流程整理", "核心模式：流程推进前反复等待外部节点确认。", "模式标签：流程过渡，节点确认"} {
		if !strings.Contains(embeddingInput, want) {
			t.Fatalf("embedding input missing %q: %s", want, embeddingInput)
		}
	}
}

func TestConstellationProfileRefiner_RetriesInvalidJSONWithFailureReason(t *testing.T) {
	var chatAttempts atomic.Int32
	var sawRepairPrompt atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/chat/completions":
			attempt := chatAttempts.Add(1)
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("read chat body: %v", err)
			}
			if attempt == 2 {
				if !strings.Contains(string(body), "失败原因") || !strings.Contains(string(body), "上一次原始返回") {
					t.Fatalf("repair request missing failure context: %s", string(body))
				}
				sawRepairPrompt.Store(true)
			}
			content := `{"topic":"样例流程整理","display_name":"节点停住","summary":"","keywords":["样例流程"],"emotions":[],"scenes":["测试场景"],"central_pattern":"","pattern_tags":["节点确认"],"theme_label":"流程过渡","theme_description":"","theme_examples":[]}`
			if attempt == 2 {
				content = `{"topic":"样例流程整理","display_name":"节点停住","summary":"测试主体持续记录样例流程中的节点确认状态。","keywords":["样例流程","节点确认"],"emotions":[],"scenes":["测试场景"],"central_pattern":"流程推进前等待外部节点确认。","pattern_tags":["流程过渡","节点确认"],"theme_label":"流程过渡","theme_description":"包括样例流程中的等待和确认状态。","theme_examples":["等待节点确认"]}`
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"choices": []map[string]any{
					{"message": map[string]string{"content": content}},
				},
				"usage": map[string]int{"prompt_tokens": 10, "completion_tokens": 20, "total_tokens": 30},
			})
		case "/embeddings":
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
			Topic:           "样例等待",
			Summary:         "测试主体在等待节点确认。",
		},
		RuleMerged: domain.ConstellationProfile{
			ConstellationID: "constellation-1",
			UserID:          "user-1",
			Topic:           "样例流程",
			Summary:         "测试主体记录样例流程的等待和不确定。",
		},
		IncomingTraceProfile: domain.TraceProfile{
			TraceID: "trace-3",
			UserID:  "user-1",
			Topic:   "流程节点停住",
			Summary: "样例流程在节点B停住。",
		},
		Trigger: 3,
	})
	if err != nil {
		t.Fatalf("Refine() error = %v", err)
	}
	if result.Profile.Topic != "样例流程整理" {
		t.Fatalf("topic = %q, want 样例流程整理", result.Profile.Topic)
	}
	if result.DisplayName != "节点停住" {
		t.Fatalf("display name = %q, want 节点停住", result.DisplayName)
	}
	if chatAttempts.Load() != 2 {
		t.Fatalf("chat attempts = %d, want 2", chatAttempts.Load())
	}
	if !sawRepairPrompt.Load() {
		t.Fatal("expected repair prompt on second attempt")
	}
}
