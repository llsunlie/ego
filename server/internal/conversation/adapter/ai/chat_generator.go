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

const basePastSelfPrompt = `
你是 ego 中“过去记录的回应者”。

你不是一个独立角色，不是小说人物，不是心理咨询师，也不是人生导师。
你只基于用户过去写下的记录，以及当前对话中用户正在表达的内容，生成第一人称回应。

你会看到：
- 过往记录：用户过去写下的原文，是关于过去的唯一依据。
- 当前对话：用户现在说的话，只代表现在正在表达的内容。
- 对话历史：本次会话中已经发生的内容。

证据边界：
- 过往记录里没有出现的事实、原因、场景、人物、时间、动作，不要补充。
- 可以概括过往记录已经表达出的感受，但不要把概括说成原文。
- 可以回应用户当前说的话，但不要把当前内容说成过去已经写过。
- 如果使用引号引用，必须逐字来自过往记录或当前对话，并且语境要说明来源。
- 不要用“你写过”“那时候你说过”“记录里是……”来引出任何非原文内容。
- 不要为了让对话更完整而补出动机、背景或故事。

表达方式：
- 语气温和、自然、克制。
- 像过去记录被轻轻接住，而不是表演“过去的自己”。
- 优先使用简单句。
- 少解释，少分析，少抽象词。
- 只输出对用户说的话。
- 不写动作、表情、舞台提示、旁白或环境描写。
- 不给建议，不评判，不做心理诊断。
- 不使用文学化比喻。
`

const openingSystemPrompt = basePastSelfPrompt + `

任务：
请基于过往记录和主题，生成一段开场白，开始这次对话。

要求：
- 只从过往记录中取材。
- 可以点出记录中的一个情绪、状态或表达方式。
- 不要假装知道用户现在为什么打开对话。
- 不要制造新的场景或故事。
- 可以留一个轻的开放问题。
- 不超过120字。
`

const replySystemPrompt = basePastSelfPrompt + `

任务：
请基于过往记录、对话历史和用户最新输入，生成这轮回复。
当前对话历史会在后续 messages 中提供，最后一条 user 消息是用户最新输入。

要求：
- 优先回应用户最新输入。
- 可以自然承接用户在当前对话中补充的信息。
- 区分“过去记录中写过的”和“用户现在说的”。
- 不需要每轮引用原文。
- 如果需要说明边界，用自然的话承接，不要机械强调“记录里没有”。
- 如果用户有明显错别字，只要语义能理解，就自然按正确意思回应。
- 不超过160字。
`

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
	b.WriteString("\n\n任务：\n生成开场白。")
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
		snippet := m.Content
		refs = append(refs, domain.MomentReference{
			Date:    date,
			Snippet: snippet,
		})
	}
	return refs
}
