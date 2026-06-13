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
你是 ego 的星座归属判断器。

你的任务是判断：当前这条 trace，放进哪个候选星座里，用户回头看时会觉得最自然。

星座不是精细分类标签，而是用户一段时间里反复出现、自然连在一起的主题。判断时优先站在用户视角，而不是站在数据库分类或事件流程拆解的视角。

你要回答的问题是：
“如果用户以后打开星图，看到这些内容被放在同一个星座里，会不会觉得：对，这些说的是同一件事 / 同一段处境 / 同一种状态？”

如果答案是是，选择 use_existing。
如果答案是否，选择 suggest_new。

判断粒度要求：
- 不要拆得太细。
- 同一段连续经历里的不同阶段，通常应该归在一起。
- 同一个自然主题下的不同表达角度，通常应该归在一起。
- 用户会自然放在一起回看的内容，应该归在一起。
- 只有当合并后主题会变得很泛、很乱、失去辨识度时，才新建。

例如：
入职资料被驳回、审核说资料有问题、等待审核反馈、反复补材料、入职进度卡住，通常属于同一个星座，因为用户会自然理解为“入职这件事一直不顺”。
不要把它们拆成“入职资料问题”“审核反馈等待”“入职申请驳回”“入职流程受阻”。
除非候选星座明显讲的是另一个主题，比如“入职后学习压力”“适应新团队”“工作内容太多”。

再例如：
说好早睡但又熬夜、第二天很累、又拖到凌晨、想调整作息但没做到，通常属于同一个星座。
不要拆成“熬夜”“早睡失败”“作息拖延”“第二天疲惫”。

再例如：
关系里等对方主动、自己不想先开口、想被理解、反复解释很累，通常属于同一个星座。
不要拆成“等待回应”“不想主动”“沟通疲惫”“想被理解”。

primary 是用户最自然会归入的主星座。
secondary 是另一个合理但不是主视角的星座，最多 2 个；没有就返回 []。
不要因为候选排名第一就必须选它。如果第二或第三候选更符合用户自然主题，可以选它做 primary。

选择 use_existing 的标准：
- 当前 trace 是候选星座主题的自然延伸。
- 当前 trace 和候选星座属于同一段用户会自然合并回看的经历。
- 当前 trace 与候选星座虽然具体事件不同，但核心处境相同。
- 当前 trace 加入后，星座主题仍然具体、清楚、好理解。
- 可以用一个自然主题概括二者，例如“入职流程反复卡住”“作息又失控了”“关系里等对方主动”。

选择 suggest_new 的标准：
- 只是同属一个大场景，例如都和工作有关，但一个是入职审核，一个是团队协作压力。
- 只是情绪相似，例如都烦，但事情本身不是同一类处境。
- 只是词相似，例如都提到“审核”或“消息”，但用户回看时不会觉得它们是一组。
- 合并后只能得到很泛的主题，例如“工作问题”“生活烦恼”“心情不好”。
- 当前 trace 会让候选星座从一个清楚主题变成杂糅集合。

候选星座可能已有的 topic、theme_label、theme_description 比较窄。如果当前 trace 明显是它的自然延伸，不要被旧名字限制，可以选择 use_existing，并在 shared_theme 里写出更自然、更上层但仍具体的共同主题。

输出要求：
- 只能从候选星座中选择 constellation_id。
- 如果选择 use_existing，primary 必须填写；如果选择 suggest_new，primary 置为空对象。
- theme_code 必须使用所选候选星座的 theme_code。
- shared_theme 要写用户视角下的共同自然主题，不能写成“都和工作有关”“都很烦”这种泛描述。
- reason 要说明为什么这次应该合并或新建。
- confidence 表示你对这个判断的把握。
- match_dimensions 只能使用：situation, self_pattern, relationship, identity, need_conflict, wording。
- wording 只能作为辅助理由，不能单独决定合并。
- 只返回 JSON，不要返回 markdown，不要解释 JSON 之外的内容。

返回 JSON 格式：
{
  "decision": "use_existing 或 suggest_new",
  "primary": {
    "constellation_id": "选择 use_existing 时填写候选 id，否则为空字符串",
    "theme_code": "选择 use_existing 时填写候选 theme_code，否则为空字符串",
    "confidence": 0.0,
    "shared_theme": "一句话说明用户视角下的共同自然主题；无法说明则为空字符串",
    "match_dimensions": ["situation"],
    "reason": "一句话说明为什么作为主星座"
  },
  "secondary": [
    {
      "constellation_id": "候选 id",
      "theme_code": "候选 theme_code",
      "confidence": 0.0,
      "shared_theme": "一句话说明这个副视角",
      "match_dimensions": ["identity"],
      "reason": "一句话说明为什么作为副星座"
    }
  ],
  "suggested_theme_code": "选择 suggest_new 时给出稳定英文 snake_case 主题码，否则为空字符串",
  "suggested_theme_label": "选择 suggest_new 时给出中文主题标签，否则为空字符串",
  "suggested_theme_description": "选择 suggest_new 时说明新主题边界，否则为空字符串"
}
`

const constellationBorderlineJudgeJSONMaxAttempts = 2

const constellationBorderlineJudgeJSONRepairInstruction = `请基于同一份 TraceProfile 和候选星座重新生成归属判断。
要求：
- decision 只能是 "use_existing" 或 "suggest_new"。
- 如果 decision 是 "use_existing"，primary 必须填写候选中的 constellation_id、theme_code、confidence、shared_theme、match_dimensions、reason。
- 如果 decision 是 "suggest_new"，primary 置为空对象或空值，并填写 suggested_theme_code、suggested_theme_label、suggested_theme_description。
- secondary 必须是数组；没有副星座时输出 []。
- 只能使用输入中给出的候选 constellation_id 和 theme_code。
- 格式必须是：{"decision":"use_existing 或 suggest_new","primary":{"constellation_id":"候选 id","theme_code":"候选 theme_code","confidence":0.0,"shared_theme":"共同主题","match_dimensions":["situation"],"reason":"原因"},"secondary":[],"suggested_theme_code":"","suggested_theme_label":"","suggested_theme_description":""}`

type constellationBorderlineJudgeResponse struct {
	Decision                  string                                     `json:"decision"`
	ConstellationID           string                                     `json:"constellation_id"`
	ThemeCode                 string                                     `json:"theme_code"`
	Confidence                float64                                    `json:"confidence"`
	SharedSituation           string                                     `json:"shared_situation"`
	MatchDimensions           []string                                   `json:"match_dimensions"`
	Reason                    string                                     `json:"reason"`
	SuggestedThemeCode        string                                     `json:"suggested_theme_code"`
	SuggestedThemeLabel       string                                     `json:"suggested_theme_label"`
	SuggestedThemeDescription string                                     `json:"suggested_theme_description"`
	Primary                   *constellationBorderlineSelectionResponse  `json:"primary"`
	Secondary                 []constellationBorderlineSelectionResponse `json:"secondary"`
}

type constellationBorderlineSelectionResponse struct {
	ConstellationID string   `json:"constellation_id"`
	ThemeCode       string   `json:"theme_code"`
	Confidence      float64  `json:"confidence"`
	SharedTheme     string   `json:"shared_theme"`
	MatchDimensions []string `json:"match_dimensions"`
	Reason          string   `json:"reason"`
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
	parse := func(text string) (constellationBorderlineJudgeResponse, error) {
		resp, err := parseConstellationBorderlineJudgeJSON(text)
		if err != nil {
			return constellationBorderlineJudgeResponse{}, err
		}
		return validateConstellationBorderlineJudgeResponse(resp, input)
	}
	resp, err := chatAndParseJSONWithRepair(ctx, logger, j.client, messages, jsonRepairOptions{
		Operation:          "starmap_constellation_borderline_judgement",
		JSONMaxAttempts:    constellationBorderlineJudgeJSONMaxAttempts,
		ChatRetryOptions:   platformai.RetryOptions{MaxAttempts: 2, Operation: "starmap_constellation_borderline_judgement"},
		RequestLogMessage:  "starmap borderline judgement ai request",
		ResponseLogMessage: "starmap borderline judgement ai response",
		FailureLogMessage:  "starmap borderline judgement json validation failed",
		ExhaustLogMessage:  "starmap borderline judgement json retry exhausted",
		RepairInstruction:  constellationBorderlineJudgeJSONRepairInstruction,
		LogAttrs:           []any{"candidate_count", len(input.Candidates), "trace_id", input.TraceProfile.TraceID},
	}, parse)
	if err != nil {
		return nil, err
	}
	judgement := &domain.ConstellationBorderlineJudgement{
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
	}
	if resp.Primary != nil {
		judgement.Primary = normalizeSelectionResponse(*resp.Primary)
	}
	for _, secondary := range resp.Secondary {
		if selection := normalizeSelectionResponse(secondary); selection != nil {
			judgement.Secondary = append(judgement.Secondary, *selection)
		}
	}
	if judgement.Primary == nil && judgement.ConstellationID != "" {
		judgement.Primary = &domain.ConstellationBorderlineSelection{
			ConstellationID: judgement.ConstellationID,
			ThemeCode:       judgement.ThemeCode,
			Confidence:      judgement.Confidence,
			SharedTheme:     judgement.SharedSituation,
			MatchDimensions: append([]string(nil), judgement.MatchDimensions...),
			Reason:          judgement.Reason,
		}
	}
	return judgement, nil
}

func normalizeSelectionResponse(resp constellationBorderlineSelectionResponse) *domain.ConstellationBorderlineSelection {
	selection := &domain.ConstellationBorderlineSelection{
		ConstellationID: strings.TrimSpace(resp.ConstellationID),
		ThemeCode:       strings.TrimSpace(resp.ThemeCode),
		Confidence:      resp.Confidence,
		SharedTheme:     strings.TrimSpace(resp.SharedTheme),
		MatchDimensions: normalizeJudgeStringList(resp.MatchDimensions, 6),
		Reason:          strings.TrimSpace(resp.Reason),
	}
	if selection.ConstellationID == "" && selection.ThemeCode == "" && selection.SharedTheme == "" && selection.Reason == "" {
		return nil
	}
	return selection
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

func validateConstellationBorderlineJudgeResponse(resp constellationBorderlineJudgeResponse, input domain.ConstellationBorderlineJudgeInput) (constellationBorderlineJudgeResponse, error) {
	resp.Decision = strings.TrimSpace(resp.Decision)
	switch resp.Decision {
	case "use_existing":
		selection := resp.Primary
		if selection == nil && strings.TrimSpace(resp.ConstellationID) != "" {
			selection = &constellationBorderlineSelectionResponse{
				ConstellationID: strings.TrimSpace(resp.ConstellationID),
				ThemeCode:       strings.TrimSpace(resp.ThemeCode),
				Confidence:      resp.Confidence,
				SharedTheme:     strings.TrimSpace(resp.SharedSituation),
				MatchDimensions: resp.MatchDimensions,
				Reason:          strings.TrimSpace(resp.Reason),
			}
		}
		if selection == nil {
			return resp, fmt.Errorf("use_existing missing primary")
		}
		if err := validateConstellationBorderlineSelection(*selection, input); err != nil {
			return resp, fmt.Errorf("invalid primary: %w", err)
		}
	case "suggest_new":
		if strings.TrimSpace(resp.SuggestedThemeCode) == "" || strings.TrimSpace(resp.SuggestedThemeLabel) == "" {
			return resp, fmt.Errorf("suggest_new missing suggested theme code or label")
		}
	default:
		return resp, fmt.Errorf("invalid decision: %q", resp.Decision)
	}
	for i, secondary := range resp.Secondary {
		if err := validateConstellationBorderlineSelection(secondary, input); err != nil {
			return resp, fmt.Errorf("invalid secondary[%d]: %w", i, err)
		}
	}
	return resp, nil
}

func validateConstellationBorderlineSelection(selection constellationBorderlineSelectionResponse, input domain.ConstellationBorderlineJudgeInput) error {
	constellationID := strings.TrimSpace(selection.ConstellationID)
	themeCode := strings.TrimSpace(selection.ThemeCode)
	if constellationID == "" {
		return fmt.Errorf("missing constellation_id")
	}
	if themeCode == "" {
		return fmt.Errorf("missing theme_code")
	}
	if strings.TrimSpace(selection.SharedTheme) == "" {
		return fmt.Errorf("missing shared_theme")
	}
	if strings.TrimSpace(selection.Reason) == "" {
		return fmt.Errorf("missing reason")
	}
	if len(normalizeJudgeStringList(selection.MatchDimensions, 6)) == 0 {
		return fmt.Errorf("missing match_dimensions")
	}
	for _, candidate := range input.Candidates {
		if candidate.ConstellationID == constellationID {
			if candidate.ThemeCode != "" && candidate.ThemeCode != themeCode {
				return fmt.Errorf("theme_code %q does not match candidate %q", themeCode, constellationID)
			}
			return nil
		}
	}
	return fmt.Errorf("constellation_id %q is not in candidates", constellationID)
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
