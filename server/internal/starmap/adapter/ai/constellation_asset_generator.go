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

const assetSystemPrompt = `你是一个善于洞察的观察者。用户的多段想法已被聚类为一个星座。请基于这些想法，生成星座资产。

返回格式必须是严格的 JSON，不要包含 markdown 代码块标记：
{"topic":"星座主题","name":"星座名","insight":"洞察文本","prompts":["引导1","引导2","引导3"]}

要求：
- topic: 不超过15个字，概括这个星座的核心主题（类似标签），用于后续匹配
- name: 不超过8个字，给这个星座起一个诗意的名字
- insight: 一句话（不超过80字），以第二人称"你"叙述，发现这些想法之间的情绪或思维关联。不要给建议，只做观察。
- prompts: 2-3个写作引导提示，帮助用户继续探索这个主题`

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
