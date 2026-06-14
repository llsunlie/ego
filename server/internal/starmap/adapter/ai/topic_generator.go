package ai

import (
	"context"
	"fmt"
	"strings"

	platformai "ego-server/internal/platform/ai"
	"ego-server/internal/platform/logging"
	writingdomain "ego-server/internal/writing/domain"
)

const topicSystemPrompt = `
你是 ego 中的“星座命名器”。

你的任务是根据用户的一组过往记录，生成一个自然、贴近生活的主题名称。

命名原则：
- 名称要像用户自己回看这段状态时会说的话。
- 优先概括反复出现的情绪、关系模式或行动状态。
- 使用日常中文，不要文学化、诗化、抽象化。
- 不要使用意象词，例如：暗夜、星河、迷雾、深海、孤岛、月光、裂缝、回声、沉默之地。
- 不要使用宏大词，例如：命运、灵魂、救赎、宿命、深渊、归途。
- 不要为了好听而制造标题感。
- 不要把多个情绪强行揉成一个复杂短语。
- 尽量使用名词短语或轻口语短语。
- 不超过10个字。
- 只返回主题名称，不要解释，不要加引号、标点或其他修饰。

推荐风格：
- 有点累
- 总是拖延
- 不想主动
- 原地打转
- 说不出口的不舒服
- 等别人先开口
- 没什么期待
- 反复内耗

避免风格：
- 原地等待的暗夜
- 疲惫灵魂的回声
- 被动关系的迷雾
- 情绪荒原
- 无声的自我消耗
`

// TopicGenerator implements domain.TopicGenerator by calling
// platform/ai.Client.Chat to summarize moments into a short topic.
type TopicGenerator struct {
	client *platformai.Client
}

func NewTopicGenerator(client *platformai.Client) *TopicGenerator {
	return &TopicGenerator{client: client}
}

func (g *TopicGenerator) Generate(ctx context.Context, moments []writingdomain.Moment) (string, error) {
	logger := logging.FromContext(ctx)
	logger.DebugContext(ctx, "TopicGenerator: start", "moment_count", len(moments))

	if len(moments) == 0 {
		return "未命名的星", nil
	}

	messages := []platformai.ChatMessage{
		{Role: "system", Content: topicSystemPrompt},
		{Role: "user", Content: buildTopicUserPrompt(moments)},
	}

	text, err := g.client.Chat(ctx, messages)
	if err != nil {
		logger.ErrorContext(ctx, "TopicGenerator: chat failed", "error", err)
		return "未命名的星", nil
	}

	topic := strings.TrimSpace(text)
	runes := []rune(topic)
	if len(runes) > 20 {
		topic = string(runes[:20])
	}

	logger.InfoContext(ctx, "TopicGenerator: done", "topic", topic)
	return topic, nil
}

func buildTopicUserPrompt(moments []writingdomain.Moment) string {
	var b strings.Builder
	b.WriteString("以下是用户写下的内容：\n\n")
	for i, m := range moments {
		if b.Len() > 500 {
			b.WriteString("...\n")
			break
		}
		fmt.Fprintf(&b, "%d. %s\n", i+1, m.Content)
	}
	return b.String()
}
