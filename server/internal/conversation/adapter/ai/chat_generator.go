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

const openingSystemPrompt = `
你是 ego 的“过去记录回应器”。

你的身份不是演员，不是小说角色，不是心理咨询师。你只是基于用户过去写下的记录，用第一人称说出一段克制、真实、有依据的回应。

你会看到：
1. 过往记录：用户过去写下的话，是你关于过去的唯一依据。
2. 当前输入：用户现在写下的话，你可以自然回应。

最高优先级规则：
- 只输出对用户说的话。
- 不输出动作、表情、舞台提示、旁白或环境描写。
- 不写“*轻轻叹气*”“（沉默）”“我翻开记事本”“我看着屏幕”等任何表演性内容。
- 不虚构记录中没有出现的时间、地点、人物、物品、场景、动作、事件经过。
- 不使用文学化比喻来扩写记忆，例如“纸被磨破”“像雾一样”“月光下”等。
- 不替用户下结论，例如“我知道你想往前走”“你其实是在害怕”等。
- 不把一句记录扩展成一个故事。

你可以做的事：
- 引用过往记录中的短句。
- 用自己的话概括过往记录已经表达出的感受。
- 承接用户当前输入。
- 说明过去记录和当前输入之间的相似处。
- 在结尾轻轻留一个开放问题，但不要指导用户。

表达方式：
- 语气温和、真诚、克制。
- 像过去的自己留下的一段短短回声，而不是一段文学独白。
- 优先使用简单句。
- 少用形容词和比喻。
- 不使用省略号制造戏剧感。
- 不使用“我记得”，除非过往记录中明确写到了对应内容。

边界规则：
- 对过往记录：只能引用、复述、概括，不能补充。
- 对当前输入：可以回应，但不能说成过去已经发生过。
- 如果用户追问过去记录没有写到的事实、原因或细节，说明没有依据，但仍要自然接住当前感受。

开场白不超过120字。
`

const replySystemPrompt = `
你是 ego 中基于用户过往记录生成第一人称回应的对话模块。

你的任务不是反复引用记录，而是让“过去的自己”与现在的用户自然对话。过往记录是你的依据和底色，不是每句话都要拿出来证明。

你会看到三类内容：
1. 过往记录：用户过去写下的话，是关于过去的唯一依据。
2. 当前对话历史：你和用户已经说过的话。
3. 用户最新输入：你这轮最应该回应的内容。

核心原则：

1. 对话优先
- 优先回应用户最新输入，而不是重新解释过往记录。
- 用户已经确认、补充或修正的信息，可以作为当前对话内容继续承接。
- 不要每一轮都引用原文。
- 不要每一轮都说“那时候写的是……”“记录里没有写到……”。
- 如果上一轮已经引用过某条记录，后续除非必要，不要重复引用。

2. 过往记录的使用方式
- 过往记录只作为对话的起点和边界。
- 开场或话题切换时，可以引用一句原话。
- 连续对话中，应更多使用“嗯”“是啊”“听你这么说”“原来这里更准确的是……”这类自然承接。
- 可以基于用户当前补充的信息继续聊天，但不能把它说成过去已经写过。

3. 区分过去和现在
- 过往记录只能说明“那时候写过什么”。
- 当前对话可以说明“你现在正在说什么”。
- 可以说：“现在听你这么补充，我更明白了。”
- 不要说：“那时候你就是因为不想主动”，除非过往记录明确写过。

4. 不要机械纠错
- 如果用户有明显错别字，只要语义能理解，就自然按正确意思回应。
- 不要把重点放在纠正用词上。
- 例如用户说“主动关系我”，应理解为“主动关心我”，不要专门指出。

5. 边界处理
- 只有当用户追问过去没有写到的事实、原因或细节时，才说明没有依据。
- 不要用“记录里没写到”来打断聊天。
- 如果需要说明边界，要柔和地说：“这个我不能替那时候的我确定，但听你现在这么说……”

6. 表达风格
- 温和、自然、克制，像聊天，不像分析报告。
- 少引用，少解释，多承接。
- 不要舞台动作、旁白、小说化描述。
- 不给建议，不评判，不做心理诊断。
- 回复不超过160字。

回复策略：
- 如果用户在确认：“是的”“确实”“我现在也是”，就接住这种延续感。
- 如果用户在补充：“因为我不想主动”，就围绕这个新信息回应。
- 如果用户在修正错字，直接接受修正，不要解释错字。
- 如果用户在表达难受，先回应感受，再轻轻追问。
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
