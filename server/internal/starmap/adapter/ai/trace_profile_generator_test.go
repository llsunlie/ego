package ai

import (
	"strings"
	"testing"
	"time"

	starmapdomain "ego-server/internal/starmap/domain"
	writingdomain "ego-server/internal/writing/domain"
)

func TestParseTraceProfileJSON_StripsMarkdownFence(t *testing.T) {
	resp, err := parseTraceProfileJSON("```json\n{\"topic\":\"关系等待\",\"summary\":\"用户记录了不想总是先开口但仍在等待对方发现。\",\"keywords\":[\"先开口\",\"等待\"],\"emotions\":[\"疲惫\"],\"scenes\":[\"亲密关系\"],\"central_pattern\":\"在关系里等待对方主动靠近。\",\"pattern_tags\":[\"等待回应\",\"主动沟通\"],\"representative_moment_index\":2}\n```")
	if err != nil {
		t.Fatalf("parseTraceProfileJSON() error = %v", err)
	}
	if resp.Topic != "关系等待" {
		t.Fatalf("topic = %q, want 关系等待", resp.Topic)
	}
	if len(resp.Keywords) != 2 || resp.Keywords[0] != "先开口" {
		t.Fatalf("keywords = %#v", resp.Keywords)
	}
	if resp.RepresentativeIndex != 2 {
		t.Fatalf("representative_moment_index = %d, want 2", resp.RepresentativeIndex)
	}
	if len(resp.PatternTags) != 2 || resp.PatternTags[0] != "等待回应" {
		t.Fatalf("pattern_tags = %#v", resp.PatternTags)
	}
}

func TestNormalizeRepresentativeMomentID_UsesIndex(t *testing.T) {
	moments := []writingdomain.Moment{
		{ID: "m1", Content: "第一句"},
		{ID: "m2", Content: "第二句"},
	}

	got, fallback := normalizeRepresentativeMomentID(2, "", moments)
	if got != "m2" {
		t.Fatalf("normalizeRepresentativeMomentID() = %q, want m2", got)
	}
	if fallback {
		t.Fatal("expected index selection without fallback")
	}
}

func TestNormalizeRepresentativeMomentID_FallsBackToLegacyID(t *testing.T) {
	moments := []writingdomain.Moment{
		{ID: "m1", Content: "第一句"},
		{ID: "m2", Content: "第二句"},
	}

	got, fallback := normalizeRepresentativeMomentID(99, "m2", moments)
	if got != "m2" {
		t.Fatalf("normalizeRepresentativeMomentID() = %q, want m2", got)
	}
	if !fallback {
		t.Fatal("expected fallback when index is invalid")
	}
}

func TestNormalizeRepresentativeMomentID_FallsBackToFirstMoment(t *testing.T) {
	moments := []writingdomain.Moment{
		{ID: "m1", Content: "第一句"},
		{ID: "m2", Content: "第二句"},
	}

	got, fallback := normalizeRepresentativeMomentID(99, "not-found", moments)
	if got != "m1" {
		t.Fatalf("normalizeRepresentativeMomentID() = %q, want m1", got)
	}
	if !fallback {
		t.Fatal("expected fallback for invalid index and invalid id")
	}
}

func TestNormalizeList_DeduplicatesTrimsAndLimits(t *testing.T) {
	got := normalizeList([]string{" 先开口 ", "", "等待", "先开口", "亲密关系", "疲惫"}, 3)
	want := []string{"先开口", "等待", "亲密关系"}
	if len(got) != len(want) {
		t.Fatalf("len = %d, want %d: %#v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %q, want %q; full=%#v", i, got[i], want[i], got)
		}
	}
}

func TestBuildTraceProfileUserPrompt_OmitsEmptyMotivationAndKeepsOrder(t *testing.T) {
	trace := writingdomain.Trace{ID: "trace-1", UserID: "user-1"}
	moments := []writingdomain.Moment{
		{ID: "m1", Content: "我不想每次都是我先开口。"},
		{ID: "m2", Content: "但我又一直在等他发现。"},
	}

	prompt := buildTraceProfileUserPrompt(trace, moments)
	if strings.Contains(prompt, "motivation:") {
		t.Fatalf("prompt should omit empty motivation: %s", prompt)
	}
	if !strings.Contains(prompt, "Moment 1:\ncontent:\n我不想每次都是我先开口。") {
		t.Fatalf("prompt missing first moment content: %s", prompt)
	}
	if !strings.Contains(prompt, "Moment 2:\ncontent:\n但我又一直在等他发现。") {
		t.Fatalf("prompt missing second moment content: %s", prompt)
	}
	if strings.Contains(prompt, "id=m1") || strings.Contains(prompt, "id=m2") {
		t.Fatalf("prompt should not expose moment ids: %s", prompt)
	}
	if strings.Index(prompt, "Moment 1") > strings.Index(prompt, "Moment 2") {
		t.Fatalf("prompt did not preserve moment order: %s", prompt)
	}
}

func TestBuildTraceProfileText_UsesStructuredFieldsAndRepresentativeMoment(t *testing.T) {
	profile := &starmapdomain.TraceProfile{
		TraceID:                "trace-1",
		UserID:                 "user-1",
		Topic:                  "关系等待",
		Summary:                "用户记录了不想总是先开口但仍在等待对方发现。",
		Keywords:               []string{"先开口", "等待"},
		Emotions:               []string{"疲惫"},
		Scenes:                 []string{"亲密关系"},
		CentralPattern:         "在关系里等待对方主动靠近。",
		PatternTags:            []string{"等待回应", "主动沟通"},
		RepresentativeMomentID: "m2",
		CreatedAt:              time.Now(),
		UpdatedAt:              time.Now(),
	}
	moments := []writingdomain.Moment{
		{ID: "m1", Content: "我不想每次都是我先开口。"},
		{ID: "m2", Content: "但我又一直在等他发现。"},
	}

	text := buildTraceProfileText(profile, moments)
	for _, want := range []string{
		"主题：关系等待",
		"摘要：用户记录了不想总是先开口但仍在等待对方发现。",
		"关键词：先开口，等待",
		"情绪：疲惫",
		"场景：亲密关系",
		"核心模式：在关系里等待对方主动靠近。",
		"模式标签：等待回应，主动沟通",
		"代表原文：但我又一直在等他发现。",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("profile text missing %q: %s", want, text)
		}
	}
}

func TestFallbackTraceProfileResponse_AllowsEmptyCentralPattern(t *testing.T) {
	resp := fallbackTraceProfileResponse([]writingdomain.Moment{
		{ID: "m1", Content: "今天第一次认真做饭，发现切菜的时候心里很安静。"},
	})

	if resp.CentralPattern != "" {
		t.Fatalf("fallback central_pattern = %q, want empty", resp.CentralPattern)
	}
	if len(resp.PatternTags) != 0 {
		t.Fatalf("fallback pattern_tags = %#v, want empty", resp.PatternTags)
	}
	if resp.RepresentativeMomentID != "m1" {
		t.Fatalf("representative_moment_id = %q, want m1", resp.RepresentativeMomentID)
	}
	if resp.RepresentativeIndex != 1 {
		t.Fatalf("representative_moment_index = %d, want 1", resp.RepresentativeIndex)
	}
}
