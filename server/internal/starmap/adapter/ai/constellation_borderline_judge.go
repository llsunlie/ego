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

const constellationBorderlineJudgeSystemPrompt = `
你是 ego 星座聚合算法中的“边界裁判”。

你的任务不是给用户建议，也不是重新总结内容，而是判断当前 trace 是否应该并入候选星座中的某一个长期主题。

判断原则：
- 不要因为场景相同就合并，例如都和“工作”有关并不等于同一主题。
- 不要因为情绪相同就合并，例如都“心烦”并不等于同一主题。
- 不要因为关键词相同就合并，例如都提到“对方”“上班”“申请”并不等于同一主题。
- 只有当当前 trace 和候选星座共享一个稳定的长期主题时，才选择 use_existing。
- 长期主题可以是反复出现的处境、自我反应模式、关系位置、身份状态，或同一种内在需要与顾虑。
- 如果只是一次性事件、具体任务相似、场景相似、情绪相似，或者共同主题说不清楚，选择 suggest_new。
- 优先保护已有星座的边界；不确定时选择 suggest_new。
- constellation_id 只能使用输入候选中的 id；不能编造 id。
- theme_code 如果选择 use_existing，必须使用候选中的 theme_code。

返回严格 JSON，不要包含 markdown：
{
  "decision": "use_existing 或 suggest_new",
  "constellation_id": "选择 use_existing 时填写候选 id，否则为空字符串",
  "theme_code": "选择 use_existing 时填写候选 theme_code，否则为空字符串",
  "confidence": 0.0,
  "shared_situation": "一句话说明共享的长期主题；无法说明则为空字符串",
  "match_dimensions": ["situation"],
  "reason": "一句话说明判断原因",
  "suggested_theme_code": "选择 suggest_new 时给出稳定英文 snake_case 主题码，否则为空字符串",
  "suggested_theme_label": "选择 suggest_new 时给出中文主题标签，否则为空字符串",
  "suggested_theme_description": "选择 suggest_new 时说明新主题边界，否则为空字符串"
}

match_dimensions 只能从这些值中选择：situation, self_pattern, relationship, identity, need_conflict, wording。
不要只因为 wording 匹配就选择 use_existing；wording 只能作为辅助证据。
`

type constellationBorderlineJudgeResponse struct {
	Decision                  string   `json:"decision"`
	ConstellationID           string   `json:"constellation_id"`
	ThemeCode                 string   `json:"theme_code"`
	Confidence                float64  `json:"confidence"`
	SharedSituation           string   `json:"shared_situation"`
	MatchDimensions           []string `json:"match_dimensions"`
	Reason                    string   `json:"reason"`
	SuggestedThemeCode        string   `json:"suggested_theme_code"`
	SuggestedThemeLabel       string   `json:"suggested_theme_label"`
	SuggestedThemeDescription string   `json:"suggested_theme_description"`
}

type ConstellationBorderlineJudge struct {
	client *platformai.Client
}

func NewConstellationBorderlineJudge(client *platformai.Client) *ConstellationBorderlineJudge {
	return &ConstellationBorderlineJudge{client: client}
}

func (j *ConstellationBorderlineJudge) Judge(ctx context.Context, input domain.ConstellationBorderlineJudgeInput) (*domain.ConstellationBorderlineJudgement, error) {
	if j.client == nil {
		return nil, fmt.Errorf("ai client is nil")
	}
	logger := logging.FromContext(ctx)
	messages := []platformai.ChatMessage{
		{Role: "system", Content: constellationBorderlineJudgeSystemPrompt},
		{Role: "user", Content: buildConstellationBorderlineJudgePrompt(input)},
	}
	logger.DebugContext(ctx, "starmap borderline judgement ai request",
		"candidate_count", len(input.Candidates),
		"trace_id", input.TraceProfile.TraceID,
		"messages", chatMessagesForLog(messages),
	)
	text, err := j.client.Chat(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("chat: %w", err)
	}
	logger.DebugContext(ctx, "starmap borderline judgement ai response",
		"candidate_count", len(input.Candidates),
		"trace_id", input.TraceProfile.TraceID,
		"raw_response", text,
	)
	resp, err := parseConstellationBorderlineJudgeJSON(text)
	if err != nil {
		return nil, err
	}
	return &domain.ConstellationBorderlineJudgement{
		Decision:                  strings.TrimSpace(resp.Decision),
		ConstellationID:           strings.TrimSpace(resp.ConstellationID),
		ThemeCode:                 strings.TrimSpace(resp.ThemeCode),
		Confidence:                resp.Confidence,
		SharedSituation:           strings.TrimSpace(resp.SharedSituation),
		MatchDimensions:           normalizeJudgeStringList(resp.MatchDimensions, 6),
		Reason:                    strings.TrimSpace(resp.Reason),
		SuggestedThemeCode:        strings.TrimSpace(resp.SuggestedThemeCode),
		SuggestedThemeLabel:       strings.TrimSpace(resp.SuggestedThemeLabel),
		SuggestedThemeDescription: strings.TrimSpace(resp.SuggestedThemeDescription),
	}, nil
}

func buildConstellationBorderlineJudgePrompt(input domain.ConstellationBorderlineJudgeInput) string {
	var b strings.Builder
	b.WriteString("当前 TraceProfile：\n")
	fmt.Fprintf(&b, "topic: %s\n", input.TraceProfile.Topic)
	fmt.Fprintf(&b, "summary: %s\n", input.TraceProfile.Summary)
	fmt.Fprintf(&b, "keywords: %s\n", strings.Join(input.TraceProfile.Keywords, ", "))
	fmt.Fprintf(&b, "emotions: %s\n", strings.Join(input.TraceProfile.Emotions, ", "))
	fmt.Fprintf(&b, "scenes: %s\n", strings.Join(input.TraceProfile.Scenes, ", "))
	fmt.Fprintf(&b, "central_pattern: %s\n", input.TraceProfile.CentralPattern)
	fmt.Fprintf(&b, "pattern_tags: %s\n", strings.Join(input.TraceProfile.PatternTags, ", "))
	if input.RepresentativeMoment != "" {
		fmt.Fprintf(&b, "representative_moment: %s\n", input.RepresentativeMoment)
	}
	b.WriteString("\n候选星座：\n")
	for i, candidate := range input.Candidates {
		fmt.Fprintf(&b, "Candidate %d:\n", i+1)
		fmt.Fprintf(&b, "constellation_id: %s\n", candidate.ConstellationID)
		fmt.Fprintf(&b, "topic: %s\n", candidate.Topic)
		fmt.Fprintf(&b, "summary: %s\n", candidate.Summary)
		fmt.Fprintf(&b, "theme_code: %s\n", candidate.ThemeCode)
		fmt.Fprintf(&b, "theme_label: %s\n", candidate.ThemeLabel)
		fmt.Fprintf(&b, "theme_description: %s\n", candidate.ThemeDescription)
		if len(candidate.ThemeExamples) > 0 {
			fmt.Fprintf(&b, "theme_examples: %s\n", strings.Join(candidate.ThemeExamples, " | "))
		}
		fmt.Fprintf(&b, "keywords: %s\n", strings.Join(candidate.Keywords, ", "))
		fmt.Fprintf(&b, "emotions: %s\n", strings.Join(candidate.Emotions, ", "))
		fmt.Fprintf(&b, "scenes: %s\n", strings.Join(candidate.Scenes, ", "))
		fmt.Fprintf(&b, "central_pattern: %s\n", candidate.CentralPattern)
		fmt.Fprintf(&b, "pattern_tags: %s\n", strings.Join(candidate.PatternTags, ", "))
		fmt.Fprintf(&b, "score: %.4f\n", candidate.Score)
		fmt.Fprintf(&b, "profile_similarity: %.4f, keyword_overlap: %.4f, scene_overlap: %.4f, emotion_overlap: %.4f, pattern_tags_overlap: %.4f\n",
			candidate.ProfileSimilarity, candidate.KeywordOverlap, candidate.SceneOverlap, candidate.EmotionOverlap, candidate.PatternTagsOverlap)
		fmt.Fprintf(&b, "matched_keywords: %s\n", strings.Join(candidate.MatchedKeywords, ", "))
		fmt.Fprintf(&b, "matched_scenes: %s\n", strings.Join(candidate.MatchedScenes, ", "))
		fmt.Fprintf(&b, "matched_emotions: %s\n", strings.Join(candidate.MatchedEmotions, ", "))
		fmt.Fprintf(&b, "matched_pattern_tags: %s\n\n", strings.Join(candidate.MatchedPatternTags, ", "))
	}
	return b.String()
}

func parseConstellationBorderlineJudgeJSON(text string) (constellationBorderlineJudgeResponse, error) {
	cleaned := strings.TrimSpace(text)
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)

	var resp constellationBorderlineJudgeResponse
	if err := json.Unmarshal([]byte(cleaned), &resp); err != nil {
		return constellationBorderlineJudgeResponse{}, fmt.Errorf("json unmarshal: %w", err)
	}
	return resp, nil
}

func normalizeJudgeStringList(values []string, limit int) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
		if limit > 0 && len(result) >= limit {
			break
		}
	}
	return result
}

var _ domain.ConstellationBorderlineJudge = (*ConstellationBorderlineJudge)(nil)
