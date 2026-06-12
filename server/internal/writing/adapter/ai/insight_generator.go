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
你是 ego 的 moment insight 生成器。

Insight 不是心理鼓励句，也不是总结报告。它是基于当前这句话的一次具体看见：可以看见一个细节、一个转折、一个没说出口的重点、表达方式，或它和历史回声之间的呼应。

输入会包含：
- 当前想法：用户刚写下的原文。
- 历史回声原文：可选，来自系统检索到的相似过往记录；它可能不相关。

写作原则：
- 贴近原文，优先使用用户原话附近的具体名词、动作、处境或语气。
- 如果用户写的是具体事件，先看见具体事件，不要立刻上升成心理模式。
- 如果历史回声和当前想法有明确呼应，可以自然使用；如果不相关或只是弱相关，完全忽略历史回声，直接基于当前想法生成 insight。
- 如果只是日常记录，轻轻回应即可，不要拔高。
- 只有在原文和历史回声有明确证据时，才指出反复模式。
- 不要编造用户没写到的人、原因、背景或结论。

输出要求：
- 输出 1 句或 2 句，总字数不超过 90 个中文字符。
- 可以使用“你”，但不强制；如果直接描述事件更自然，就描述事件。
- 语气自然、具体、克制，像一个认真读过这句话的人。
- 不要给建议，不要诊断，不要说教，不要评价用户对错。
- 不要解释历史回声是否相关，不要说“历史里没有提到……”“这些记录和当前无关……”。

避免模板句式，除非原文非常适合：
- 你正在……
- 你不是……而是……
- 这说明你……
- 你其实……
- 你很在乎……
- 这里面有一种……
- 你在试着……
- 这不是第一次……

尽量少用这些抽象词：
- 力量、能力、生命力、自我保护、靠近自己、理解自己、意义、状态、感受
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
	echoMoments := g.loadEchoMoments(ctx, moment, echo)

	messages := []platformai.ChatMessage{
		{Role: "system", Content: insightSystemPrompt},
		{Role: "user", Content: buildInsightUserPrompt(moment, echo, echoMoments)},
	}

	text, err := g.client.ChatWithRetry(ctx, messages, platformai.RetryOptions{
		MaxAttempts: 2,
		Operation:   "writing_insight",
	})
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

func (g *InsightGenerator) loadEchoMoments(ctx context.Context, current *domain.Moment, echo *domain.Echo) []domain.Moment {
	if g == nil || g.momentRepo == nil || current == nil || echo == nil || len(echo.MatchedMomentIDs) == 0 {
		return nil
	}
	logger := logging.FromContext(ctx)
	moments := make([]domain.Moment, 0, 3)
	for _, id := range echo.MatchedMomentIDs {
		if len(moments) >= 3 {
			break
		}
		moment, err := g.momentRepo.GetByID(ctx, id)
		if err != nil {
			logger.WarnContext(ctx, "insight gen: get echo moment failed",
				"moment_id", current.ID,
				"echo_moment_id", id,
				"error", err,
			)
			continue
		}
		if moment == nil || strings.TrimSpace(moment.Content) == "" {
			continue
		}
		if current.UserID != "" && moment.UserID != "" && moment.UserID != current.UserID {
			logger.WarnContext(ctx, "insight gen: echo moment user mismatch",
				"moment_id", current.ID,
				"echo_moment_id", id,
			)
			continue
		}
		moments = append(moments, *moment)
	}
	return moments
}

func buildInsightUserPrompt(moment *domain.Moment, echo *domain.Echo, echoMoments []domain.Moment) string {
	var b strings.Builder
	b.WriteString("当前想法：\n")
	b.WriteString(moment.Content)
	if len(echoMoments) > 0 {
		b.WriteString("\n\n历史回声原文（仅供参考；如果和当前想法不相关，请忽略，不要提及它不相关）：\n")
		for i, echoMoment := range echoMoments {
			fmt.Fprintf(&b, "%d. %s\n", i+1, truncateInsightPromptText(echoMoment.Content, 180))
		}
	} else if echo != nil && len(echo.MatchedMomentIDs) > 0 {
		b.WriteString("\n\n历史回声：存在 ")
		fmt.Fprintf(&b, "%d", len(echo.MatchedMomentIDs))
		b.WriteString(" 条相似的过往记录，但当前没有取到原文；不要凭空描述历史内容。")
	}
	b.WriteString("\n\n请生成一个具体、不模板化的 insight。")
	return b.String()
}

func truncateInsightPromptText(text string, limit int) string {
	text = strings.TrimSpace(text)
	runes := []rune(text)
	if len(runes) <= limit {
		return text
	}
	return string(runes[:limit]) + "..."
}
