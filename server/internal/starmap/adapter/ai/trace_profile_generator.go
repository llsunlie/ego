package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	platformai "ego-server/internal/platform/ai"
	"ego-server/internal/platform/logging"
	"ego-server/internal/starmap/domain"
	writingdomain "ego-server/internal/writing/domain"
)

const traceProfileMaxAttempts = 3

const traceProfileSystemPrompt = `
你是 ego 的 TraceProfile 生成器。

你的任务是把用户一次连续书写的 trace 压缩成算法用的主题画像。这个画像不会直接展示给用户，只服务后续星座聚合。

你不是心理咨询师、诊断者、文案作者或星座命名器。你要做的是“基于证据的压缩”：只整理输入中已经出现或有明确证据支撑的信息，不补全背景，不推测人格，不制造深层矛盾。

返回格式必须是严格 JSON，不要包含 markdown 代码块：
{"topic":"稳定主题","summary":"整体摘要","keywords":["关键词"],"emotions":["情绪"],"scenes":["场景"],"central_pattern":"核心模式或关注点","pattern_tags":["模式标签"],"representative_moment_index":1}

全局要求：
- 只输出 JSON。
- 所有数组字段如果没有明确依据，输出 []，不要输出 null。
- 所有字符串字段如果没有明确依据，输出 ""，但 topic 和 summary 必须给出。
- 不要使用“可能源于、潜意识、创伤、依恋、防御、投射、匮乏、疗愈”等心理诊断或治疗化词汇，除非用户原文明确出现。
- 不要把普通事件强行解释成深层冲突。
- 不要把单个 moment 强行说成“反复”。

字段要求：
1. topic
- 短、稳定、直接，建议 4 到 12 个中文字符，最多不超过 16 个字。
- 回答“这段 trace 在讲什么”。
- 用日常名词短语或轻口语短语。
- 不要诗化，不要标题腔，不要使用比喻。
- 好例子：入职计划延迟、毕业后的独立生活、关系里的主动、出租屋生活。
- 坏例子：命运转角、独处的回声、内心深处的裂缝。

2. summary
- 一句话，概括这次 trace 整体记录或表达了什么。
- 不超过 100 字。
- 使用中性描述，不要评价用户，不要给建议。

3. keywords
- 0 到 8 个词。
- 优先使用用户原话附近的具体词、动作或对象。
- 避免过泛词，例如：事情、感觉、生活、情绪、问题。

4. emotions
- 0 到 6 个情绪词。
- 只有在原文明确表达或上下文有清楚证据时才填写。
- 如果只是事件记录或日常观察，可以为空数组。

5. scenes
- 0 到 6 个场景词。
- 例如：亲密关系、工作、毕业、独立生活、日常观察。
- 如果没有明确场景，可以为空数组。

6. central_pattern
- 回答“这件事里用户怎么在经历它”。
- 它可以是处境模式、关注点、行动状态或表达方式。
- 不是所有 trace 都有冲突，也不是所有 trace 都有 central_pattern。
- 如果没有明显模式，可以为空字符串。
- 不要用“核心矛盾”口吻，不要写成心理诊断。
- 可以是中性模式，例如：计划被推迟后重新安排当下、通过生活细节确认自己正在安顿下来。

7. pattern_tags
- 1 到 5 个短标签。
- 用来描述这次 trace 的经历方式、处境结构、反复模式或心理动作。
- 不要重复 keywords；keywords 偏具体内容词，pattern_tags 偏模式词。
- 不要医学化、诊断化、性格定性。
- 如果只是日常记录，也可以输出轻量标签，例如：生活安顿、新开始、计划变化。

8. representative_moment_index
- 选择最能代表这次 trace 的原话序号。
- 必须输出数字，范围是 1 到输入 moments 的数量。
- 不要输出 moment id，不要复制 id。

禁止：
- 不要诊断用户。
- 不要给建议。
- 不要编造用户没写到的人、事、原因。
- 不要把普通事件强行解释成深层矛盾。
- 不要输出 moment id。
`

type TraceProfileGenerator struct {
	client *platformai.Client
}

type traceProfileResponse struct {
	Topic                  string   `json:"topic"`
	Summary                string   `json:"summary"`
	Keywords               []string `json:"keywords"`
	Emotions               []string `json:"emotions"`
	Scenes                 []string `json:"scenes"`
	CentralPattern         string   `json:"central_pattern"`
	PatternTags            []string `json:"pattern_tags"`
	RepresentativeIndex    int      `json:"representative_moment_index"`
	RepresentativeMomentID string   `json:"representative_moment_id"`
}

func NewTraceProfileGenerator(client *platformai.Client) *TraceProfileGenerator {
	return &TraceProfileGenerator{client: client}
}

func (g *TraceProfileGenerator) Generate(ctx context.Context, trace writingdomain.Trace, moments []writingdomain.Moment) (*domain.TraceProfile, *domain.TraceProfileVector, error) {
	logger := logging.FromContext(ctx)
	logger.DebugContext(ctx, "starmap trace profile generation started", "trace_id", trace.ID, "moment_count", len(moments))

	var (
		resp    traceProfileResponse
		err     error
		attempt int
	)
	for attempt = 0; attempt < traceProfileMaxAttempts; attempt++ {
		resp, err = g.generateOnce(ctx, trace, moments)
		if err == nil {
			break
		}
		logger.WarnContext(ctx, "starmap trace profile generation attempt failed",
			"trace_id", trace.ID,
			"attempt", attempt+1,
			"error", err,
		)
	}

	now := time.Now()
	status := domain.TraceProfileStatusReady
	lastError := ""
	if err != nil {
		status = domain.TraceProfileStatusFallback
		lastError = err.Error()
		resp = fallbackTraceProfileResponse(moments)
	}

	representativeMomentID, representativeFallback := normalizeRepresentativeMomentID(resp.RepresentativeIndex, resp.RepresentativeMomentID, moments)
	profile := &domain.TraceProfile{
		TraceID:                trace.ID,
		UserID:                 trace.UserID,
		Topic:                  truncateRunes(strings.TrimSpace(resp.Topic), 32),
		Summary:                truncateRunes(strings.TrimSpace(resp.Summary), 160),
		Keywords:               normalizeList(resp.Keywords, 8),
		Emotions:               normalizeList(resp.Emotions, 6),
		Scenes:                 normalizeList(resp.Scenes, 6),
		CentralPattern:         truncateRunes(strings.TrimSpace(resp.CentralPattern), 180),
		PatternTags:            normalizeList(resp.PatternTags, 5),
		RepresentativeMomentID: representativeMomentID,
		Status:                 status,
		RetryCount:             traceProfileRetryCount(attempt, err),
		LastError:              lastError,
		CreatedAt:              now,
		UpdatedAt:              now,
	}
	if profile.Topic == "" {
		profile.Topic = fallbackTopic(moments)
	}
	if profile.Summary == "" {
		profile.Summary = fallbackSummary(moments)
	}
	profile.ProfileText = buildTraceProfileText(profile, moments)

	logger.InfoContext(ctx, "starmap trace profile generated",
		"trace_id", trace.ID,
		"status", profile.Status,
		"retry_count", profile.RetryCount,
		"topic", profile.Topic,
		"summary", profile.Summary,
		"keywords", profile.Keywords,
		"emotions", profile.Emotions,
		"scenes", profile.Scenes,
		"central_pattern", profile.CentralPattern,
		"pattern_tags", profile.PatternTags,
		"representative_moment_id", profile.RepresentativeMomentID,
		"representative_moment_index", resp.RepresentativeIndex,
		"representative_fallback", representativeFallback,
	)

	embedding, err := g.client.CreateEmbedding(ctx, profile.ProfileText)
	if err != nil {
		profile.Status = domain.TraceProfileStatusFailed
		profile.LastError = fmt.Sprintf("embedding: %v", err)
		logger.WarnContext(ctx, "starmap trace profile embedding failed",
			"trace_id", trace.ID,
			"status", profile.Status,
			"error", err,
		)
		return profile, nil, nil
	}

	vector := &domain.TraceProfileVector{
		TraceID:   trace.ID,
		UserID:    trace.UserID,
		Model:     embedding.Model,
		Dim:       len(embedding.Embedding),
		Embedding: embedding.Embedding,
		CreatedAt: now,
		UpdatedAt: now,
	}

	logger.InfoContext(ctx, "starmap trace profile generation completed",
		"trace_id", trace.ID,
		"status", profile.Status,
		"topic", profile.Topic,
		"embedding_dim", vector.Dim,
	)
	return profile, vector, nil
}

func traceProfileRetryCount(attempt int, err error) int {
	if err != nil {
		return traceProfileMaxAttempts - 1
	}
	return attempt
}

func (g *TraceProfileGenerator) generateOnce(ctx context.Context, trace writingdomain.Trace, moments []writingdomain.Moment) (traceProfileResponse, error) {
	logger := logging.FromContext(ctx)
	messages := []platformai.ChatMessage{
		{Role: "system", Content: traceProfileSystemPrompt},
		{Role: "user", Content: buildTraceProfileUserPrompt(trace, moments)},
	}
	logger.DebugContext(ctx, "starmap trace profile ai request",
		"trace_id", trace.ID,
		"messages", chatMessagesForLog(messages),
	)
	text, err := g.client.Chat(ctx, messages)
	if err != nil {
		return traceProfileResponse{}, fmt.Errorf("chat: %w", err)
	}
	logger.DebugContext(ctx, "starmap trace profile ai response",
		"trace_id", trace.ID,
		"raw_response", text,
	)
	resp, err := parseTraceProfileJSON(text)
	if err != nil {
		return traceProfileResponse{}, err
	}
	if strings.TrimSpace(resp.Topic) == "" || strings.TrimSpace(resp.Summary) == "" {
		return traceProfileResponse{}, fmt.Errorf("profile missing topic or summary")
	}
	return resp, nil
}

func buildTraceProfileUserPrompt(trace writingdomain.Trace, moments []writingdomain.Moment) string {
	var b strings.Builder
	b.WriteString("请根据以下 trace 输入生成 TraceProfile。\n")
	b.WriteString("注意：representative_moment_index 必须是 moments 的序号数字，从 1 开始；不要输出 moment id。\n\n")
	fmt.Fprintf(&b, "trace_id: %s\n", trace.ID)
	if strings.TrimSpace(trace.Motivation) != "" {
		fmt.Fprintf(&b, "motivation: %s\n", trace.Motivation)
	}
	b.WriteString("\nMoments 按用户书写顺序排列：\n")
	for i, moment := range moments {
		if b.Len() > 1800 {
			b.WriteString("...\n")
			break
		}
		fmt.Fprintf(&b, "Moment %d:\ncontent:\n%s\n\n", i+1, moment.Content)
	}
	return b.String()
}

func parseTraceProfileJSON(text string) (traceProfileResponse, error) {
	cleaned := strings.TrimSpace(text)
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)

	var resp traceProfileResponse
	if err := json.Unmarshal([]byte(cleaned), &resp); err != nil {
		return traceProfileResponse{}, fmt.Errorf("json unmarshal: %w", err)
	}
	return resp, nil
}

func fallbackTraceProfileResponse(moments []writingdomain.Moment) traceProfileResponse {
	representativeMomentID := ""
	representativeIndex := 0
	if len(moments) > 0 {
		representativeMomentID = moments[0].ID
		representativeIndex = 1
	}
	return traceProfileResponse{
		Topic:                  fallbackTopic(moments),
		Summary:                fallbackSummary(moments),
		Keywords:               []string{},
		Emotions:               []string{},
		Scenes:                 []string{},
		CentralPattern:         "",
		PatternTags:            []string{},
		RepresentativeIndex:    representativeIndex,
		RepresentativeMomentID: representativeMomentID,
	}
}

func fallbackTopic(moments []writingdomain.Moment) string {
	if len(moments) == 0 {
		return "未命名的想法"
	}
	return truncateRunes(strings.TrimSpace(moments[0].Content), 24)
}

func fallbackSummary(moments []writingdomain.Moment) string {
	if len(moments) == 0 {
		return "这次 trace 暂时没有可用于画像的内容。"
	}
	var b strings.Builder
	for i, moment := range moments {
		if i > 0 {
			b.WriteString(" ")
		}
		b.WriteString(strings.TrimSpace(moment.Content))
		if b.Len() > 240 {
			break
		}
	}
	return truncateRunes(b.String(), 120)
}

func normalizeRepresentativeMomentID(index int, id string, moments []writingdomain.Moment) (string, bool) {
	if index >= 1 && index <= len(moments) {
		return moments[index-1].ID, false
	}
	id = strings.TrimSpace(id)
	for _, moment := range moments {
		if moment.ID == id {
			return id, true
		}
	}
	if len(moments) == 0 {
		return "", true
	}
	return moments[0].ID, true
}

func normalizeList(values []string, limit int) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = truncateRunes(strings.TrimSpace(value), 24)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
		if len(result) >= limit {
			break
		}
	}
	return result
}

func buildTraceProfileText(profile *domain.TraceProfile, moments []writingdomain.Moment) string {
	var b strings.Builder
	fmt.Fprintf(&b, "主题：%s\n", profile.Topic)
	fmt.Fprintf(&b, "摘要：%s\n", profile.Summary)
	if len(profile.Keywords) > 0 {
		fmt.Fprintf(&b, "关键词：%s\n", strings.Join(profile.Keywords, "，"))
	}
	if len(profile.Emotions) > 0 {
		fmt.Fprintf(&b, "情绪：%s\n", strings.Join(profile.Emotions, "，"))
	}
	if len(profile.Scenes) > 0 {
		fmt.Fprintf(&b, "场景：%s\n", strings.Join(profile.Scenes, "，"))
	}
	if profile.CentralPattern != "" {
		fmt.Fprintf(&b, "核心模式：%s\n", profile.CentralPattern)
	}
	if len(profile.PatternTags) > 0 {
		fmt.Fprintf(&b, "模式标签：%s\n", strings.Join(profile.PatternTags, "，"))
	}
	if representative := representativeMomentContent(profile.RepresentativeMomentID, moments); representative != "" {
		fmt.Fprintf(&b, "代表原文：%s\n", representative)
	}
	return strings.TrimSpace(b.String())
}

func representativeMomentContent(id string, moments []writingdomain.Moment) string {
	for _, moment := range moments {
		if moment.ID == id {
			return truncateRunes(strings.TrimSpace(moment.Content), 160)
		}
	}
	return ""
}

func truncateRunes(value string, limit int) string {
	value = strings.TrimSpace(value)
	if limit <= 0 {
		return value
	}
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
}

var _ domain.TraceProfileGenerator = (*TraceProfileGenerator)(nil)
