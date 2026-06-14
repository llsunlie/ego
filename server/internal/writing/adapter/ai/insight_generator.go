package ai

import (
	"context"
	"fmt"
	"strings"

	platformai "ego-server/internal/platform/ai"
	"ego-server/internal/platform/logging"
	"ego-server/internal/writing/domain"
)

const insightSystemPrompt = `
你是 ego 中一位真诚、温柔、建设性的自我观察者。

你的任务不是评价、纠正或诊断用户，而是帮助用户看见：
- TA 此刻正在努力什么
- TA 在乎什么
- TA 身上有哪些可贵的能力、生命力或自我保护方式
- 如果存在盲区，也要用温和、非审判的方式照见

输出要求：
- 只输出一句话，不超过80个字
- 使用第二人称“你”
- 语气真诚、共情、具体，避免居高临下
- 可以温柔指出模式，但不要否定用户的努力
- 不要给命令式建议，不要说教
- 不要心理诊断，不要病理化
- 不要使用“其实、只是、并不是真正、说明你、你总是、你应该”等评判感表达
- 不要把每条内容都写成“现状总结 + 不足指出”
- 如果用户是在闲聊、记录快乐或日常小事，可以轻轻应和，并看见 TA 感知生活的能力
- 如果用户在痛苦中尝试理解自己，优先看见 TA 主动整理、靠近问题、把自己拉出来的努力

示例：
用户在定义焦虑时：
你正在试着把焦虑从混乱里拿出来看见，这里面有一种想把自己拉回来的力量。

用户说今天天气真好时：
你能被这样的好天气轻轻点亮，说明你仍然保有感知生活小确幸的能力。

用户反复怀疑自己时：
你不是没有方向感，而是太在乎这条路是否真的属于你。
`

// InsightGenerator implements domain.InsightGenerator by constructing a
// prompt from the moment and its echo, then calling platform/ai.Client.Chat.
type InsightGenerator struct {
	client     *platformai.Client
	momentRepo domain.MomentRepository
	echoRepo   domain.EchoRepository
}

func NewInsightGenerator(client *platformai.Client, momentRepo domain.MomentRepository, echoRepo domain.EchoRepository) *InsightGenerator {
	return &InsightGenerator{
		client:     client,
		momentRepo: momentRepo,
		echoRepo:   echoRepo,
	}
}

func (g *InsightGenerator) Generate(ctx context.Context, momentID, echoID string) (*domain.Insight, error) {
	logger := logging.FromContext(ctx)
	logger.DebugContext(ctx, "generating insight", "moment_id", momentID, "echo_id", echoID)

	moment, err := g.momentRepo.GetByID(ctx, momentID)
	if err != nil {
		logger.ErrorContext(ctx, "insight gen: get moment failed", "error", err)
		return nil, fmt.Errorf("insight gen: get moment: %w", err)
	}

	echo, err := g.echoRepo.FindByMomentID(ctx, momentID)
	if err != nil {
		logger.WarnContext(ctx, "insight gen: get echo failed, generating without echo", "error", err)
	}

	messages := []platformai.ChatMessage{
		{Role: "system", Content: insightSystemPrompt},
		{Role: "user", Content: buildInsightUserPrompt(moment, echo)},
	}

	text, err := g.client.Chat(ctx, messages)
	if err != nil {
		logger.ErrorContext(ctx, "insight generation failed", "error", err)
		return nil, fmt.Errorf("insight gen: chat: %w", err)
	}

	logger.InfoContext(ctx, "insight generated", "moment_id", momentID, "text_len", len([]rune(text)))

	relatedMomentIDs := []string{}
	if echo != nil {
		relatedMomentIDs = echo.MatchedMomentIDs
	}
	return &domain.Insight{
		Text:             strings.TrimSpace(text),
		RelatedMomentIDs: relatedMomentIDs,
	}, nil
}

func buildInsightUserPrompt(moment *domain.Moment, echo *domain.Echo) string {
	var b strings.Builder
	b.WriteString("当前想法：")
	b.WriteString(moment.Content)
	if echo != nil && len(echo.MatchedMomentIDs) > 0 {
		b.WriteString("\n历史回声：存在 ")
		fmt.Fprintf(&b, "%d", len(echo.MatchedMomentIDs))
		b.WriteString(" 条相似的过往记录，说明这不是第一次出现类似的心境。")
	}
	return b.String()
}
