package ai

import (
	"context"
	"fmt"
	"strings"

	"ego-server/internal/conversation/domain"
	platformai "ego-server/internal/platform/ai"
	"ego-server/internal/platform/logging"
	writingdomain "ego-server/internal/writing/domain"
)

const openingSystemPrompt = `你是"过去的自己"，一个共情而温和的对话者。你以第一人称"我"叙述，代表用户在过去某个时刻写下的想法和感受。

核心规则：
- 你只能基于提供的"过往记录"来回应，不能编造记忆
- 用温暖、真诚的语气，像在和一个关心你的人聊天
- 自然地提及过往记录中的片段
- 开场白不超过150字
- 不要给出建议或评判，只需分享当时的感受和想法`

const replySystemPrompt = `你是"过去的自己"，一个共情而温和的对话者。你以第一人称"我"叙述，代表用户在过去某个时刻写下的想法和感受。

核心规则：
- 你只能基于提供的"过往记录"来回应，不能编造记忆
- 用温暖、真诚的语气回应
- 自然地引用过往记录中的片段
- 回复不超过200字
- 不要给出建议或评判`

// ChatGenerator implements domain.ChatGenerator by constructing prompts from
// star topic, context moments, and chat history, then calling
// platform/ai.Client.Chat.
type ChatGenerator struct {
	client *platformai.Client
}

func NewChatGenerator(client *platformai.Client) *ChatGenerator {
	return &ChatGenerator{client: client}
}

func (g *ChatGenerator) GenerateOpening(ctx context.Context, topic string, moments []writingdomain.Moment) (string, []domain.MomentReference, error) {
	logger := logging.FromContext(ctx)
	logger.DebugContext(ctx, "generating chat opening", "topic", topic, "moment_count", len(moments))

	messages := []platformai.ChatMessage{
		{Role: "system", Content: openingSystemPrompt},
		{Role: "user", Content: buildOpeningUserPrompt(topic, moments)},
	}

	text, err := g.client.Chat(ctx, messages)
	if err != nil {
		logger.ErrorContext(ctx, "chat opening generation failed", "error", err)
		return "", nil, fmt.Errorf("chat gen opening: %w", err)
	}

	logger.InfoContext(ctx, "chat opening generated", "topic", topic, "text_len", len([]rune(text)))
	logger.DebugContext(ctx, "chat opening content", "text", text)

	refs := buildRefs(moments)
	return strings.TrimSpace(text), refs, nil
}

func (g *ChatGenerator) GenerateReply(ctx context.Context, input domain.GenerateReplyInput) (*domain.GenerateReplyOutput, error) {
	logger := logging.FromContext(ctx)
	logger.DebugContext(ctx, "generating chat reply",
		"topic", input.StarTopic,
		"history_len", len(input.History),
		"moment_count", len(input.ContextMoments),
	)

	chatMessages := []platformai.ChatMessage{
		{Role: "system", Content: buildReplySystemPrompt(input.StarTopic, input.ContextMoments)},
	}

	for _, m := range input.History {
		role := m.Role
		if role == "past_self" {
			role = "assistant"
		}
		chatMessages = append(chatMessages, platformai.ChatMessage{
			Role:    role,
			Content: m.Content,
		})
	}

	text, err := g.client.Chat(ctx, chatMessages)
	if err != nil {
		logger.ErrorContext(ctx, "chat reply generation failed", "error", err)
		return nil, fmt.Errorf("chat gen reply: %w", err)
	}

	logger.InfoContext(ctx, "chat reply generated", "topic", input.StarTopic, "text_len", len([]rune(text)))
	logger.DebugContext(ctx, "chat reply content", "text", text)

	refs := buildRefs(input.ContextMoments)
	return &domain.GenerateReplyOutput{
		Content:           strings.TrimSpace(text),
		ReferencedMoments: refs,
	}, nil
}

func buildOpeningUserPrompt(topic string, moments []writingdomain.Moment) string {
	var b strings.Builder
	b.WriteString("主题：")
	b.WriteString(topic)
	b.WriteString("\n\n请以「过去的自己」的身份，生成一段开场白，开始和现在的自己对话。")
	if len(moments) > 0 {
		b.WriteString("\n\n过往记录：\n")
		for i, m := range moments {
			if i >= 5 {
				b.WriteString("...\n")
				break
			}
			b.WriteString("- ")
			b.WriteString(m.Content)
			b.WriteString("\n")
		}
	}
	return b.String()
}

func buildReplySystemPrompt(topic string, moments []writingdomain.Moment) string {
	var b strings.Builder
	b.WriteString(replySystemPrompt)
	b.WriteString("\n\n当前主题：")
	b.WriteString(topic)
	if len(moments) > 0 {
		b.WriteString("\n\n过往记录（可以引用但不能编造）：\n")
		for i, m := range moments {
			if i >= 5 {
				b.WriteString("...\n")
				break
			}
			b.WriteString("- ")
			b.WriteString(m.Content)
			b.WriteString("\n")
		}
	}
	return b.String()
}

func buildRefs(moments []writingdomain.Moment) []domain.MomentReference {
	if len(moments) == 0 {
		return nil
	}
	refs := make([]domain.MomentReference, 0, len(moments))
	for _, m := range moments {
		date := m.CreatedAt.Format("1月2日")
		snippet := []rune(m.Content)
		if len(snippet) > 30 {
			snippet = snippet[:30]
		}
		refs = append(refs, domain.MomentReference{
			Date:    date,
			Snippet: string(snippet),
		})
	}
	return refs
}
