package ai

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestClient_CreateEmbeddingWithRetry_RetriesRetryableError(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/embeddings" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if attempts.Add(1) == 1 {
			http.Error(w, "temporary unavailable", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"model": "embed-test",
			"data": []map[string]any{
				{"embedding": []float32{0.1, 0.2}},
			},
			"usage": map[string]int{"prompt_tokens": 1, "total_tokens": 1},
		})
	}))
	defer server.Close()

	client := NewClient(Config{
		EmbeddingBaseURL: server.URL,
		EmbeddingAPIKey:  "test-key",
		EmbeddingModel:   "embed-test",
	}, slog.Default())

	result, err := client.CreateEmbeddingWithRetry(context.Background(), "hello", RetryOptions{
		MaxAttempts: 2,
		Backoff:     func(int) time.Duration { return 0 },
		Operation:   "test_embedding",
	})
	if err != nil {
		t.Fatalf("CreateEmbeddingWithRetry() error = %v", err)
	}
	if attempts.Load() != 2 {
		t.Fatalf("attempts = %d, want 2", attempts.Load())
	}
	if result.Model != "embed-test" || len(result.Embedding) != 2 {
		t.Fatalf("result = %#v", result)
	}
}

func TestClient_CreateEmbeddingWithRetry_DoesNotRetryClientError(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts.Add(1)
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer server.Close()

	client := NewClient(Config{
		EmbeddingBaseURL: server.URL,
		EmbeddingAPIKey:  "test-key",
		EmbeddingModel:   "embed-test",
	}, slog.Default())

	_, err := client.CreateEmbeddingWithRetry(context.Background(), "hello", RetryOptions{
		MaxAttempts: 3,
		Backoff:     func(int) time.Duration { return 0 },
		Operation:   "test_embedding",
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if attempts.Load() != 1 {
		t.Fatalf("attempts = %d, want 1", attempts.Load())
	}
}

func TestClient_ChatWithRetry_RetriesRetryableError(t *testing.T) {
	var attempts atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if attempts.Add(1) == 1 {
			http.Error(w, "temporary unavailable", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"content": "ok"}},
			},
			"usage": map[string]int{"prompt_tokens": 1, "completion_tokens": 1, "total_tokens": 2},
		})
	}))
	defer server.Close()

	client := NewClient(Config{
		ChatBaseURL: server.URL,
		ChatAPIKey:  "test-key",
		ChatModel:   "chat-test",
	}, slog.Default())

	result, err := client.ChatWithRetry(context.Background(), []ChatMessage{{Role: "user", Content: "hello"}}, RetryOptions{
		MaxAttempts: 2,
		Backoff:     func(int) time.Duration { return 0 },
		Operation:   "test_chat",
	})
	if err != nil {
		t.Fatalf("ChatWithRetry() error = %v", err)
	}
	if attempts.Load() != 2 {
		t.Fatalf("attempts = %d, want 2", attempts.Load())
	}
	if result != "ok" {
		t.Fatalf("result = %q, want ok", result)
	}
}
