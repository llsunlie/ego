package ai_test

import (
	"context"
	"os"
	"testing"

	"ego-server/internal/config"
	"ego-server/internal/platform/ai"
)

func newTestClient(t *testing.T) *ai.Client {
	t.Helper()

	// Trigger .env loading so devs can put AI_API_KEY in server/.env
	// instead of exporting it manually.
	_ = config.Load()

	apiKey := os.Getenv("AI_API_KEY")
	if apiKey == "" {
		t.Skip("AI_API_KEY not set in env or .env, skipping integration test")
	}

	return ai.NewClient(ai.Config{
		APIKey:         apiKey,
		BaseURL:        envOrDefault("AI_BASE_URL", "https://api.siliconflow.cn/v1"),
		EmbeddingModel: envOrDefault("AI_EMBEDDING_MODEL", "Qwen/Qwen3-VL-Embedding-8B"),
		ChatModel:      envOrDefault("AI_CHAT_MODEL", "deepseek-ai/DeepSeek-V3"),
	})
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func TestClient_CreateEmbedding(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	result, err := client.CreateEmbedding(ctx, "Hello, world!")
	if err != nil {
		t.Fatalf("CreateEmbedding: %v", err)
	}

	if result.Model == "" {
		t.Error("expected non-empty Model in result")
	}
	if len(result.Embedding) == 0 {
		t.Error("expected non-empty Embedding in result")
	}

	t.Logf("model=%s tokens=%d dim=%d", result.Model, result.Usage.TotalTokens, len(result.Embedding))
}

func TestClient_Chat(t *testing.T) {
	client := newTestClient(t)
	ctx := context.Background()

	messages := []ai.ChatMessage{
		{Role: "system", Content: "用一句话回答用户的问题。"},
		{Role: "user", Content: "你好，请介绍一下你自己。"},
	}

	reply, err := client.Chat(ctx, messages)
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if reply == "" {
		t.Error("expected non-empty chat reply")
	}

	t.Logf("reply=%s", reply)
}
