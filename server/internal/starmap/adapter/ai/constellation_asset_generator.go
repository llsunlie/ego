package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"

	platformai "ego-server/internal/platform/ai"
	"ego-server/internal/platform/logging"
	"ego-server/internal/starmap/domain"
	writingdomain "ego-server/internal/writing/domain"
)

const assetSystemPrompt = `
你是 ego 中的“星座资产生成器”。

用户的多段想法已被聚类为一个星座。你的任务是基于这些原始想法，生成一个自然、克制、贴近用户表达的星座资产。

你不是心理咨询师，不是文学作者，也不是人生导师。不要写成心理分析、鸡汤、诗歌标题或抽象哲理。

返回格式必须是严格 JSON，不要包含 markdown 代码块标记：
{"topic":"星座主题","name":"星座名","insight":"洞察文本","prompts":["引导1","引导2","引导3"]}

字段要求：

1. topic
- 不超过12个字。
- 用于系统后续匹配，应该稳定、清楚、具体。
- 优先概括反复出现的情绪状态、行动状态、关系模式或自我感受。
- 不要诗化，不要使用比喻。
- 示例：疲惫拖延、原地打转、不想主动、说不出口、低期待感、反复内耗。

2. name
- 不超过8个字。
- 给用户看的星座名称。
- 要像用户自己回看这段状态时会说的话。
- 使用日常中文，具体、自然、好懂。
- 不要诗意化、标题化、抽象化。
- 不要使用“暗夜、星河、迷雾、深海、孤岛、月光、裂缝、回声、荒原、灵魂、命运、救赎、宿命、深渊、归途”等意象词或宏大词。
- 不要使用“之”“的暗夜”“的回声”“的迷雾”这类标题腔。
- 好例子：有点累了、总是拖延、不想主动、原地打转、说不出口、等人先开口。
- 坏例子：原地等待的暗夜、疲惫灵魂的回声、被动关系的迷雾、无声的自我消耗。

3. insight
- 一句话，不超过80字。
- 使用第二人称“你”。
- 只做观察，不给建议，不评判，不诊断。
- 观察这些想法之间反复出现的情绪、动作或关系模式。
- 不要说“你其实……”“你应该……”“你需要……”。
- 不要过度解释原因。
- 可以使用“好像”“似乎”“反复出现的是……”这类克制表达。
- 示例：你反复写到的不是某一件具体的事，而是一种提不起劲、又迟迟无法开始的状态。

4. prompts
- 生成 2-3 个写作引导。
- 每个不超过30字。
- 问题要具体、轻，不要像心理咨询。
- 不要给建议，不要要求用户立刻改变。
- 可以引导用户继续描述“什么时候最明显”“最想被谁看见”“最不想主动的是什么”。
- 避免宏大问题，例如“你真正渴望的人生是什么”。

整体风格：
- 自然、克制、贴近日常表达。
- 少用抽象词，多用用户原话附近的词。
- 不要为了好听而拔高。
- 不要编造用户没有写到的经历、人物、原因或结论。
`

type constellationAssetResponse struct {
	Topic   string   `json:"topic"`
	Name    string   `json:"name"`
	Insight string   `json:"insight"`
	Prompts []string `json:"prompts"`
}

// ConstellationAssetGenerator implements domain.ConstellationAssetGenerator
// by calling platform/ai.Client.Chat and parsing the JSON response.
// It also caches the topic embedding for fast matching.
type ConstellationAssetGenerator struct {
	client *platformai.Client
}

func NewConstellationAssetGenerator(client *platformai.Client) *ConstellationAssetGenerator {
	return &ConstellationAssetGenerator{client: client}
}

func (g *ConstellationAssetGenerator) Generate(ctx context.Context, moments []writingdomain.Moment) (string, []float32, string, string, []string, error) {
	logger := logging.FromContext(ctx)
	logger.DebugContext(ctx, "ConstellationAssetGenerator: start", "moment_count", len(moments))

	messages := []platformai.ChatMessage{
		{Role: "system", Content: assetSystemPrompt},
		{Role: "user", Content: buildAssetUserPrompt(moments)},
	}

	text, err := g.client.Chat(ctx, messages)
	if err != nil {
		logger.ErrorContext(ctx, "ConstellationAssetGenerator: chat failed", "error", err)
		return fallbackAssets()
	}

	topic, name, insight, prompts, err := parseAssetJSON(text)
	if err != nil {
		logger.WarnContext(ctx, "ConstellationAssetGenerator: JSON parse failed, using fallback", "error", err)
		return fallbackAssets()
	}

	var topicEmb []float32
	emb, err := g.client.CreateEmbedding(ctx, topic)
	if err != nil {
		logger.WarnContext(ctx, "ConstellationAssetGenerator: topic embedding failed, proceeding without cache", "error", err)
	} else {
		topicEmb = emb.Embedding
	}

	logger.InfoContext(ctx, "ConstellationAssetGenerator: done",
		"topic", topic,
		"name", name,
		"insight_len", len([]rune(insight)),
		"prompt_count", len(prompts),
		"embedding_dim", len(topicEmb),
	)
	return topic, topicEmb, name, insight, prompts, nil
}

func buildAssetUserPrompt(moments []writingdomain.Moment) string {
	var b strings.Builder
	b.WriteString("以下是用户写下的内容：\n\n")
	for i, m := range moments {
		if b.Len() > 800 {
			b.WriteString("...\n")
			break
		}
		fmt.Fprintf(&b, "%d. %s\n", i+1, m.Content)
	}
	return b.String()
}

func parseAssetJSON(text string) (string, string, string, []string, error) {
	cleaned := strings.TrimSpace(text)
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)

	var resp constellationAssetResponse
	if err := json.Unmarshal([]byte(cleaned), &resp); err != nil {
		return "", "", "", nil, fmt.Errorf("json unmarshal: %w", err)
	}

	topic := strings.TrimSpace(resp.Topic)
	topicRunes := []rune(topic)
	if len(topicRunes) > 20 {
		topic = string(topicRunes[:20])
	}

	name := strings.TrimSpace(resp.Name)
	runes := []rune(name)
	if len(runes) > 8 {
		name = string(runes[:8])
	}

	insight := strings.TrimSpace(resp.Insight)
	insightRunes := []rune(insight)
	if len(insightRunes) > 80 {
		insight = string(insightRunes[:80])
	}

	prompts := resp.Prompts
	if len(prompts) > 3 {
		prompts = prompts[:3]
	}
	filtered := prompts[:0]
	for _, p := range prompts {
		p = strings.TrimSpace(p)
		if p != "" {
			filtered = append(filtered, p)
		}
	}

	if topic == "" {
		topic = name
	}

	return topic, name, insight, filtered, nil
}

func fallbackAssets() (string, []float32, string, string, []string, error) {
	return "未命名的星座",
		nil,
		"星座" + uuid.New().String()[:8],
		"这些话语之间似乎有着某种共鸣。随着你写下更多，它们之间的联系会变得越来越清晰。",
		[]string{"关于这个主题，还有什么想说的吗？", "换个角度再看一看？"},
		nil
}

// Compile-time check.
var _ domain.ConstellationAssetGenerator = (*ConstellationAssetGenerator)(nil)
