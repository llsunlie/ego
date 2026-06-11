package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	platformai "ego-server/internal/platform/ai"
	"ego-server/internal/platform/logging"
	"ego-server/internal/starmap/domain"
)

const constellationProfileRefinerSystemPrompt = `
你是 ego 的 ConstellationProfile 重写器。

你的任务是在一个星座吸收足够多 trace 后，重新整理它的长期主题画像。你不判断归属，不创建新星座，只把已有星座画像写得更稳定、更短、更有辨识度。

输入会包含：
- 当前星座旧画像
- 规则合并后的候选画像
- 新加入的 TraceProfile
- 代表性原话

你要做：
- 压缩同义词，例如“入职、报到、上班第一天”可收敛成最自然的代表词。
- 去掉过泛词，例如“事情、感觉、状态、问题、生活、想法”。
- 保留能区分这个星座的具体词、场景和模式标签。
- 不要让主题变得太细，也不要变成“工作问题”“情绪不好”这类泛主题。
- 不要诊断用户，不要给建议，不要编造输入里没有的背景。

输出只能是严格 JSON，不要 markdown：
{
  "topic": "稳定主题",
  "summary": "一句话摘要",
  "keywords": ["关键词"],
  "emotions": ["情绪"],
  "scenes": ["场景"],
  "central_pattern": "用户在这类事情中反复经历的处境、关注点或行动状态",
  "pattern_tags": ["模式标签"],
  "theme_label": "中文主题标签",
  "theme_description": "主题边界",
  "theme_examples": ["代表例子"]
}

字段约束：
- topic：4 到 16 个中文字符，直接、稳定，不诗化。
- summary：不超过 100 字。
- keywords：最多 8 个。
- emotions：最多 6 个。
- scenes：最多 6 个。
- central_pattern：不超过 120 字；没有明显模式可为空字符串。
- pattern_tags：最多 6 个。
- theme_label：不超过 24 字。
- theme_description：不超过 120 字，说明“包括什么 / 不包括什么”。
- theme_examples：最多 3 条，每条不超过 40 字。
`

type ConstellationProfileRefiner struct {
	client *platformai.Client
}

type constellationProfileRefineResponse struct {
	Topic            string   `json:"topic"`
	Summary          string   `json:"summary"`
	Keywords         []string `json:"keywords"`
	Emotions         []string `json:"emotions"`
	Scenes           []string `json:"scenes"`
	CentralPattern   string   `json:"central_pattern"`
	PatternTags      []string `json:"pattern_tags"`
	ThemeLabel       string   `json:"theme_label"`
	ThemeDescription string   `json:"theme_description"`
	ThemeExamples    []string `json:"theme_examples"`
}

func NewConstellationProfileRefiner(client *platformai.Client) *ConstellationProfileRefiner {
	return &ConstellationProfileRefiner{client: client}
}

func (r *ConstellationProfileRefiner) Refine(ctx context.Context, input domain.ConstellationProfileRefineInput) (*domain.ConstellationProfileRefinement, error) {
	if r == nil || r.client == nil {
		return nil, fmt.Errorf("ai client is nil")
	}
	logger := logging.FromContext(ctx)
	messages := []platformai.ChatMessage{
		{Role: "system", Content: constellationProfileRefinerSystemPrompt},
		{Role: "user", Content: buildConstellationProfileRefinePrompt(input)},
	}
	logger.DebugContext(ctx, "starmap constellation profile refinement ai request",
		"constellation_id", input.RuleMerged.ConstellationID,
		"trigger", input.Trigger,
		"messages", chatMessagesForLog(messages),
	)
	text, err := r.client.ChatWithRetry(ctx, messages, platformai.RetryOptions{
		MaxAttempts: 2,
		Operation:   "starmap_constellation_profile_refinement",
	})
	if err != nil {
		return nil, fmt.Errorf("chat: %w", err)
	}
	logger.DebugContext(ctx, "starmap constellation profile refinement ai response",
		"constellation_id", input.RuleMerged.ConstellationID,
		"raw_response", text,
	)
	resp, err := parseConstellationProfileRefineJSON(text)
	if err != nil {
		return nil, err
	}
	profile := applyConstellationProfileRefinement(input.RuleMerged, resp)
	embedding, err := r.client.CreateEmbeddingWithRetry(ctx, profile.ProfileText, platformai.RetryOptions{
		MaxAttempts: 3,
		Operation:   "starmap_constellation_profile_refinement",
	})
	if err != nil {
		return nil, fmt.Errorf("embedding: %w", err)
	}
	return &domain.ConstellationProfileRefinement{
		Profile:          profile,
		Model:            embedding.Model,
		Dim:              len(embedding.Embedding),
		ProfileEmbedding: embedding.Embedding,
	}, nil
}

func buildConstellationProfileRefinePrompt(input domain.ConstellationProfileRefineInput) string {
	var b strings.Builder
	fmt.Fprintf(&b, "trigger_trace_count: %d\n\n", input.Trigger)
	b.WriteString("当前星座旧画像：\n")
	writeConstellationProfileForRefine(&b, input.Existing)
	b.WriteString("\n\n规则合并后的候选画像：\n")
	writeConstellationProfileForRefine(&b, input.RuleMerged)
	b.WriteString("\n\n新加入的 TraceProfile：\n")
	fmt.Fprintf(&b, "topic: %s\n", input.IncomingTraceProfile.Topic)
	fmt.Fprintf(&b, "summary: %s\n", input.IncomingTraceProfile.Summary)
	fmt.Fprintf(&b, "keywords: %s\n", strings.Join(input.IncomingTraceProfile.Keywords, "，"))
	fmt.Fprintf(&b, "emotions: %s\n", strings.Join(input.IncomingTraceProfile.Emotions, "，"))
	fmt.Fprintf(&b, "scenes: %s\n", strings.Join(input.IncomingTraceProfile.Scenes, "，"))
	fmt.Fprintf(&b, "central_pattern: %s\n", input.IncomingTraceProfile.CentralPattern)
	fmt.Fprintf(&b, "pattern_tags: %s\n", strings.Join(input.IncomingTraceProfile.PatternTags, "，"))
	if strings.TrimSpace(input.RepresentativeMoment) != "" {
		fmt.Fprintf(&b, "representative_moment: %s\n", input.RepresentativeMoment)
	}
	return b.String()
}

func writeConstellationProfileForRefine(b *strings.Builder, profile domain.ConstellationProfile) {
	fmt.Fprintf(b, "topic: %s\n", profile.Topic)
	fmt.Fprintf(b, "summary: %s\n", profile.Summary)
	fmt.Fprintf(b, "keywords: %s\n", strings.Join(profile.Keywords, "，"))
	fmt.Fprintf(b, "emotions: %s\n", strings.Join(profile.Emotions, "，"))
	fmt.Fprintf(b, "scenes: %s\n", strings.Join(profile.Scenes, "，"))
	fmt.Fprintf(b, "central_pattern: %s\n", profile.CentralPattern)
	fmt.Fprintf(b, "pattern_tags: %s\n", strings.Join(profile.PatternTags, "，"))
	fmt.Fprintf(b, "theme_label: %s\n", profile.ThemeLabel)
	fmt.Fprintf(b, "theme_description: %s\n", profile.ThemeDescription)
	fmt.Fprintf(b, "theme_examples: %s\n", strings.Join(profile.ThemeExamples, "；"))
	fmt.Fprintf(b, "trace_count: %.2f\n", profile.TraceCount)
	fmt.Fprintf(b, "moment_count: %.2f\n", profile.MomentCount)
}

func parseConstellationProfileRefineJSON(text string) (constellationProfileRefineResponse, error) {
	cleaned := strings.TrimSpace(text)
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)
	var resp constellationProfileRefineResponse
	if err := json.Unmarshal([]byte(cleaned), &resp); err != nil {
		return resp, fmt.Errorf("parse constellation profile refinement json: %w", err)
	}
	if strings.TrimSpace(resp.Topic) == "" || strings.TrimSpace(resp.Summary) == "" {
		return resp, fmt.Errorf("refinement missing topic or summary")
	}
	return resp, nil
}

func applyConstellationProfileRefinement(base domain.ConstellationProfile, resp constellationProfileRefineResponse) domain.ConstellationProfile {
	result := base
	result.Topic = truncateRunes(strings.TrimSpace(resp.Topic), 32)
	result.Summary = truncateRunes(strings.TrimSpace(resp.Summary), 160)
	result.Keywords = normalizeList(resp.Keywords, 8)
	result.Emotions = normalizeList(resp.Emotions, 6)
	result.Scenes = normalizeList(resp.Scenes, 6)
	result.CentralPattern = truncateRunes(strings.TrimSpace(resp.CentralPattern), 180)
	result.PatternTags = normalizeList(resp.PatternTags, 6)
	result.ThemeLabel = truncateRunes(strings.TrimSpace(resp.ThemeLabel), 32)
	result.ThemeDescription = truncateRunes(strings.TrimSpace(resp.ThemeDescription), 120)
	result.ThemeExamples = normalizeList(resp.ThemeExamples, 3)
	if result.ThemeLabel == "" {
		result.ThemeLabel = base.ThemeLabel
	}
	if result.ThemeDescription == "" {
		result.ThemeDescription = base.ThemeDescription
	}
	result.ProfileText = buildRefinedConstellationProfileText(result)
	return result
}

func buildRefinedConstellationProfileText(profile domain.ConstellationProfile) string {
	var b strings.Builder
	if profile.ThemeCode != "" {
		fmt.Fprintf(&b, "主题码：%s\n", profile.ThemeCode)
	}
	if profile.ThemeLabel != "" {
		fmt.Fprintf(&b, "主题标签：%s\n", profile.ThemeLabel)
	}
	if profile.ThemeDescription != "" {
		fmt.Fprintf(&b, "主题边界：%s\n", profile.ThemeDescription)
	}
	if profile.Topic != "" {
		fmt.Fprintf(&b, "主题：%s\n", profile.Topic)
	}
	if profile.Summary != "" {
		fmt.Fprintf(&b, "摘要：%s\n", profile.Summary)
	}
	if len(profile.Keywords) > 0 {
		fmt.Fprintf(&b, "关键词：%s\n", strings.Join(profile.Keywords, "，"))
	}
	if len(profile.Scenes) > 0 {
		fmt.Fprintf(&b, "场景：%s\n", strings.Join(profile.Scenes, "，"))
	}
	if len(profile.Emotions) > 0 {
		fmt.Fprintf(&b, "情绪：%s\n", strings.Join(profile.Emotions, "，"))
	}
	if profile.CentralPattern != "" {
		fmt.Fprintf(&b, "核心模式：%s\n", profile.CentralPattern)
	}
	if len(profile.PatternTags) > 0 {
		fmt.Fprintf(&b, "模式标签：%s\n", strings.Join(profile.PatternTags, "，"))
	}
	if len(profile.ThemeExamples) > 0 {
		fmt.Fprintf(&b, "代表例子：%s\n", strings.Join(profile.ThemeExamples, "；"))
	}
	return strings.TrimSpace(b.String())
}

var _ domain.ConstellationProfileRefiner = (*ConstellationProfileRefiner)(nil)
