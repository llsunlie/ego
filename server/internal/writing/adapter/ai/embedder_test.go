package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	platformai "ego-server/internal/platform/ai"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *platformai.Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return platformai.NewClient(platformai.Config{
		BaseURL:        srv.URL,
		APIKey:         "test-key",
		EmbeddingModel: "test-embed-model",
		ChatModel:      "test-chat-model",
	})
}

func TestEmbedder_Generate_Success(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/embeddings" {
			t.Errorf("expected /embeddings, got %s", r.URL.Path)
		}
		resp := map[string]interface{}{
			"model": "test-embed-model",
			"data": []map[string]interface{}{
				{"embedding": []float32{0.1, 0.2, 0.3}},
			},
			"usage": map[string]interface{}{
				"prompt_tokens": 5,
				"total_tokens":  5,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	embedder := NewEmbedder(client)
	entries, err := embedder.Generate(context.Background(), "hello world")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Model != "test-embed-model" {
		t.Fatalf("expected model 'test-embed-model', got %q", entries[0].Model)
	}
	if len(entries[0].Embedding) != 3 {
		t.Fatalf("expected embedding dim 3, got %d", len(entries[0].Embedding))
	}
}

func TestEmbedder_Generate_APIError(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	embedder := NewEmbedder(client)
	_, err := embedder.Generate(context.Background(), "hello")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestEmbedder_Generate_EmptyData(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"model": "test-embed-model",
			"data":  []interface{}{}, // empty
			"usage": map[string]interface{}{
				"prompt_tokens": 0,
				"total_tokens":  0,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	embedder := NewEmbedder(client)
	_, err := embedder.Generate(context.Background(), "hello")
	if err == nil {
		t.Fatal("expected error for empty embedding data")
	}
}
