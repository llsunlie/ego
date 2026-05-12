package ai

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ego-server/internal/conversation/domain"
	platformai "ego-server/internal/platform/ai"
	writingdomain "ego-server/internal/writing/domain"
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

func TestChatGenerator_GenerateOpening_Success(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Errorf("expected /chat/completions, got %s", r.URL.Path)
		}
		resp := map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{"content": "嗨，我是过去的你。关于「工作压力」，那时候我写下了很多想法，现在回头看看，还是挺有感触的。"}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	gen := NewChatGenerator(client)
	moments := []writingdomain.Moment{
		{ID: "m1", Content: "今天加班到很晚，感觉很累", CreatedAt: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)},
		{ID: "m2", Content: "老板又给我加活了", CreatedAt: time.Date(2026, 5, 3, 0, 0, 0, 0, time.UTC)},
	}

	text, refs, err := gen.GenerateOpening(context.Background(), "工作压力", moments)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "嗨，我是过去的你。关于「工作压力」，那时候我写下了很多想法，现在回头看看，还是挺有感触的。" {
		t.Fatalf("unexpected text: %q", text)
	}
	if len(refs) != 2 {
		t.Fatalf("expected 2 refs, got %d", len(refs))
	}
	if refs[0].Date != "5月1日" {
		t.Fatalf("expected '5月1日', got %q", refs[0].Date)
	}
}

func TestChatGenerator_GenerateOpening_EmptyMoments(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{"content": "嗨，关于「未来」，那时候我还没写下太多东西。但我们可以聊聊。"}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	gen := NewChatGenerator(client)
	text, refs, err := gen.GenerateOpening(context.Background(), "未来", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text == "" {
		t.Fatal("expected non-empty text")
	}
	if refs != nil {
		t.Fatalf("expected nil refs for empty moments, got %v", refs)
	}
}

func TestChatGenerator_GenerateOpening_APIError(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	gen := NewChatGenerator(client)
	_, _, err := gen.GenerateOpening(context.Background(), "test", nil)
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestChatGenerator_GenerateOpening_EmptyChoices(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"choices": []any{},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	gen := NewChatGenerator(client)
	_, _, err := gen.GenerateOpening(context.Background(), "test", nil)
	if err == nil {
		t.Fatal("expected error for empty choices")
	}
}

func TestChatGenerator_GenerateReply_Success(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{"content": "嗯，我记得那时候的感觉。每天都很疲惫，但也只能咬牙坚持。"}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	gen := NewChatGenerator(client)
	moments := []writingdomain.Moment{
		{ID: "m1", Content: "加班到很晚", CreatedAt: time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)},
	}
	input := domain.GenerateReplyInput{
		StarTopic:      "工作压力",
		ContextMoments: moments,
		History: []domain.ChatMessage{
			{Role: "user", Content: "最近感觉怎么样？"},
			{Role: "past_self", Content: "有点累，事情太多了"},
		},
		UserMessage: "还有呢？",
	}

	output, err := gen.GenerateReply(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if output.Content != "嗯，我记得那时候的感觉。每天都很疲惫，但也只能咬牙坚持。" {
		t.Fatalf("unexpected content: %q", output.Content)
	}
	if len(output.ReferencedMoments) != 1 {
		t.Fatalf("expected 1 ref, got %d", len(output.ReferencedMoments))
	}
}

func TestChatGenerator_GenerateReply_APIError(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	})

	gen := NewChatGenerator(client)
	_, err := gen.GenerateReply(context.Background(), domain.GenerateReplyInput{
		StarTopic: "test",
	})
	if err == nil {
		t.Fatal("expected error when chat API fails")
	}
}

func TestChatGenerator_GenerateReply_TrimsWhitespace(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{"content": "\n  那些日子确实不容易。  \n"}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	gen := NewChatGenerator(client)
	output, err := gen.GenerateReply(context.Background(), domain.GenerateReplyInput{
		StarTopic: "test",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(output.Content) != output.Content {
		t.Fatalf("expected trimmed content, got %q", output.Content)
	}
}

func TestBuildOpeningUserPrompt_WithMoments(t *testing.T) {
	moments := []writingdomain.Moment{
		{Content: "今天加班到很晚"},
		{Content: "老板又给我加活了"},
	}
	result := buildOpeningUserPrompt("工作", moments)

	if !strings.Contains(result, "主题：工作") {
		t.Fatalf("expected topic in prompt, got %q", result)
	}
	if !strings.Contains(result, "过往记录") {
		t.Fatalf("expected '过往记录' in prompt, got %q", result)
	}
	if !strings.Contains(result, "今天加班到很晚") {
		t.Fatalf("expected moment content in prompt, got %q", result)
	}
}

func TestBuildOpeningUserPrompt_NoMoments(t *testing.T) {
	result := buildOpeningUserPrompt("未来", nil)

	if !strings.Contains(result, "主题：未来") {
		t.Fatalf("expected topic in prompt, got %q", result)
	}
	if strings.Contains(result, "过往记录") {
		t.Fatalf("should not contain '过往记录' when no moments, got %q", result)
	}
}

func TestBuildReplySystemPrompt(t *testing.T) {
	moments := []writingdomain.Moment{
		{Content: "今天心情不错"},
	}
	result := buildReplySystemPrompt("生活", moments)

	if !strings.Contains(result, "当前主题：生活") {
		t.Fatalf("expected topic in prompt, got %q", result)
	}
	if !strings.Contains(result, "今天心情不错") {
		t.Fatalf("expected moment content in prompt, got %q", result)
	}
	if !strings.Contains(result, "过去的自己") {
		t.Fatalf("expected system prompt intro in result, got %q", result)
	}
}

func TestBuildRefs_Empty(t *testing.T) {
	refs := buildRefs(nil)
	if refs != nil {
		t.Fatalf("expected nil refs, got %v", refs)
	}

	refs = buildRefs([]writingdomain.Moment{})
	if refs != nil {
		t.Fatalf("expected nil refs for empty slice, got %v", refs)
	}
}

func TestBuildRefs_TruncatesLongContent(t *testing.T) {
	longContent := "这是一段很长很长很长很长很长很长很长很长很长很长很长很长很长很长很长的内容"
	moments := []writingdomain.Moment{
		{Content: longContent, CreatedAt: time.Date(2026, 5, 11, 0, 0, 0, 0, time.UTC)},
	}

	refs := buildRefs(moments)
	if len(refs) != 1 {
		t.Fatalf("expected 1 ref, got %d", len(refs))
	}
	if len([]rune(refs[0].Snippet)) > 30 {
		t.Fatalf("expected snippet <= 30 runes, got %d", len([]rune(refs[0].Snippet)))
	}
	if refs[0].Date != "5月11日" {
		t.Fatalf("expected '5月11日', got %q", refs[0].Date)
	}
}

func TestChatGenerator_GenerateReply_PastSelfRoleMapping(t *testing.T) {
	// Verify that past_self role in history is mapped to assistant for the API.
	var receivedRoles []string
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Messages []struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"messages"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		for _, m := range req.Messages {
			receivedRoles = append(receivedRoles, m.Role)
		}
		resp := map[string]any{
			"choices": []map[string]any{
				{"message": map[string]any{"content": "是的，那就是我。"}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})

	gen := NewChatGenerator(client)
	input := domain.GenerateReplyInput{
		StarTopic: "test",
		History: []domain.ChatMessage{
			{Role: "user", Content: "你好"},
			{Role: "past_self", Content: "你好啊"},
			{Role: "user", Content: "最近怎么样"},
		},
	}

	_, err := gen.GenerateReply(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(receivedRoles) < 4 {
		t.Fatalf("expected at least 4 messages (system + 3 history), got %d", len(receivedRoles))
	}
	// system + user + assistant + user
	if receivedRoles[1] != "user" {
		t.Fatalf("expected role 'user' at index 1, got %q", receivedRoles[1])
	}
	if receivedRoles[2] != "assistant" {
		t.Fatalf("expected role 'assistant' at index 2 (mapped from past_self), got %q", receivedRoles[2])
	}
	if receivedRoles[3] != "user" {
		t.Fatalf("expected role 'user' at index 3, got %q", receivedRoles[3])
	}
}
