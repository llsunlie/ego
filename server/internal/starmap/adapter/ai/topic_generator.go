package ai

import (
	"context"
	"fmt"
	"strings"

	platformai "ego-server/internal/platform/ai"
	"ego-server/internal/platform/logging"
	writingdomain "ego-server/internal/writing/domain"
)

const topicSystemPrompt = `你是一个善于提炼的观察者。用户写下了几段话，请从中提取一个简短的主题短语。只返回主题文本，不超过15个字，不要加引号、标点或其他修饰。`

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
