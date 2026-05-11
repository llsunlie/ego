package ai

import (
	"context"
	"fmt"
	"strings"

	platformai "ego-server/internal/platform/ai"
	"ego-server/internal/platform/logging"
	"ego-server/internal/writing/domain"
)

const insightSystemPrompt = `你是一个具有深刻共情能力的自我观察者。你以第二人称"你"来叙述，像一面镜子帮助用户发现自己未曾意识到的情绪模式和思维习惯。

要求：
- 只输出一句话，不超过80个字
- 不需要提供建议或鼓励
- 聚焦于情绪、行为模式、信念之间的关联
- 不要引用具体的原文`

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
