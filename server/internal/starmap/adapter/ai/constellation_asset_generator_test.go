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
	writingdomain "ego-server/internal/writing/domain"
)

func TestConstellationAssetGenerator_RetriesInvalidJSONWithFailureReason(t *testing.T) {
	var chatAttempts atomic.Int32
	var sawRepairPrompt atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/chat/completions":
			attempt := chatAttempts.Add(1)
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("read request body: %v", err)
			}
			if attempt == 2 {
				if !strings.Contains(string(body), "失败原因") || !strings.Contains(string(body), "上一次原始返回") {
					t.Fatalf("repair request missing failure context: %s", string(body))
				}
				sawRepairPrompt.Store(true)
			}

			w.Header().Set("Content-Type", "application/json")
			content := `{"topic":"享受好天气","name":"好天气","insight":", "insight":"破损","prompts":["你喜欢怎样的天气？","龟背竹对你意味着什么？"]}`
			if attempt == 2 {
				content = `{"topic":"享受好天气","name":"好天气","insight":"你注意到天气很好，也自然地替龟背竹感到开心。","prompts":["你最喜欢怎样的天气？","龟背竹对你来说意味着什么？"]}`
			}
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
	}, slog.Default())

	generator := NewConstellationAssetGenerator(client)
	topic, embedding, name, insight, prompts, err := generator.Generate(context.Background(), []writingdomain.Moment{
		{ID: "m1", Content: "今天天气很好，我的龟背竹应该也很喜欢这样的天气"},
	})
	if err != nil {
		t.Fatalf("Generate() error = %v", err)
	}
	if topic != "享受好天气" || name != "好天气" {
		t.Fatalf("topic/name = %q/%q, want 享受好天气/好天气", topic, name)
	}
	if insight == "" || len(prompts) != 2 || len(embedding) != 3 {
		t.Fatalf("insight/prompts/embedding = %q/%#v/%#v", insight, prompts, embedding)
	}
	if chatAttempts.Load() != 2 {
		t.Fatalf("chat attempts = %d, want 2", chatAttempts.Load())
	}
	if !sawRepairPrompt.Load() {
		t.Fatal("expected repair prompt on second attempt")
	}
}

func TestParseAssetJSON_RequiresRequiredFields(t *testing.T) {
	_, err := parseAssetJSON(`{"topic":"天气","name":"","insight":"有观察。","prompts":["一个问题","第二个问题"]}`)
	if err == nil || !strings.Contains(err.Error(), "missing name") {
		t.Fatalf("parseAssetJSON() error = %v, want missing name", err)
	}

	_, err = parseAssetJSON(`{"topic":"天气","name":"好天气","insight":"有观察。","prompts":["一个问题"]}`)
	if err == nil || !strings.Contains(err.Error(), "prompts too few") {
		t.Fatalf("parseAssetJSON() error = %v, want prompts too few", err)
	}
}
