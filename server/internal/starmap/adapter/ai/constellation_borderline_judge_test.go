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

func TestParseConstellationBorderlineJudgeJSON_PrimaryAndSecondary(t *testing.T) {
	resp, err := parseConstellationBorderlineJudgeJSON(`{
  "decision": "use_existing",
  "primary": {
    "constellation_id": "c-main",
    "theme_code": "theme_main",
    "confidence": 0.82,
    "shared_theme": "入职流程反复卡住",
    "match_dimensions": ["situation"],
    "reason": "这是主归属"
  },
  "secondary": [
    {
      "constellation_id": "c-side",
      "theme_code": "theme_side",
      "confidence": 0.72,
      "shared_theme": "也有被审核的位置感",
      "match_dimensions": ["identity"],
      "reason": "这是副视角"
    }
  ],
  "suggested_theme_code": "",
  "suggested_theme_label": "",
  "suggested_theme_description": ""
}`)
	if err != nil {
		t.Fatalf("parseConstellationBorderlineJudgeJSON() error = %v", err)
	}
	if resp.Decision != "use_existing" {
		t.Fatalf("decision = %q", resp.Decision)
	}
	if resp.Primary == nil || resp.Primary.ConstellationID != "c-main" {
		t.Fatalf("primary = %#v", resp.Primary)
	}
	if len(resp.Secondary) != 1 || resp.Secondary[0].ConstellationID != "c-side" {
		t.Fatalf("secondary = %#v", resp.Secondary)
	}
}

func TestParseConstellationBorderlineJudgeJSON_LegacyShape(t *testing.T) {
	resp, err := parseConstellationBorderlineJudgeJSON(`{
  "decision": "use_existing",
  "constellation_id": "c-main",
  "theme_code": "theme_main",
  "confidence": 0.82,
  "shared_situation": "入职流程反复卡住",
  "match_dimensions": ["situation"],
  "reason": "这是主归属"
}`)
	if err != nil {
		t.Fatalf("parseConstellationBorderlineJudgeJSON() error = %v", err)
	}
	if resp.Decision != "use_existing" || resp.ConstellationID != "c-main" {
		t.Fatalf("legacy response = %#v", resp)
	}
}

func TestConstellationBorderlineJudge_RetriesInvalidJSONWithFailureReason(t *testing.T) {
	var chatAttempts atomic.Int32
	var sawRepairPrompt atomic.Bool
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat/completions" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		attempt := chatAttempts.Add(1)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read chat body: %v", err)
		}
		if attempt == 2 {
			if !strings.Contains(string(body), "失败原因") || !strings.Contains(string(body), "candidate-1") {
				t.Fatalf("repair request missing failure context: %s", string(body))
			}
			sawRepairPrompt.Store(true)
		}

		content := `{"decision":"use_existing","primary":{"constellation_id":"not-candidate","theme_code":"theme_1","confidence":0.8,"shared_theme":"入职等待","match_dimensions":["situation"],"reason":"相似"},"secondary":[]}`
		if attempt == 2 {
			content = `{"decision":"use_existing","primary":{"constellation_id":"candidate-1","theme_code":"theme_1","confidence":0.8,"shared_theme":"入职等待反馈","match_dimensions":["situation"],"reason":"都在记录入职反馈等待。"},"secondary":[],"suggested_theme_code":"","suggested_theme_label":"","suggested_theme_description":""}`
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{"content": content}},
			},
			"usage": map[string]int{"prompt_tokens": 10, "completion_tokens": 20, "total_tokens": 30},
		})
	}))
	defer server.Close()

	client := platformai.NewClient(platformai.Config{
		ChatAPIKey:  "test-key",
		ChatBaseURL: server.URL,
		ChatModel:   "chat-test",
	}, slog.New(slog.NewTextHandler(io.Discard, nil)))
	judge := NewConstellationBorderlineJudge(client)

	result, err := judge.Judge(context.Background(), domain.ConstellationBorderlineJudgeInput{
		TraceProfile: domain.TraceProfile{
			TraceID: "trace-1",
			Topic:   "入职反馈",
			Summary: "用户还在等待入职反馈。",
		},
		Candidates: []domain.ConstellationBorderlineCandidate{
			{
				ConstellationID: "candidate-1",
				ThemeCode:       "theme_1",
				Topic:           "入职等待",
				Summary:         "记录入职前反馈等待。",
			},
		},
	})
	if err != nil {
		t.Fatalf("Judge() error = %v", err)
	}
	if result.Primary == nil || result.Primary.ConstellationID != "candidate-1" {
		t.Fatalf("primary = %#v", result.Primary)
	}
	if chatAttempts.Load() != 2 {
		t.Fatalf("chat attempts = %d, want 2", chatAttempts.Load())
	}
	if !sawRepairPrompt.Load() {
		t.Fatal("expected repair prompt on second attempt")
	}
}
