package app

import (
	"context"
	"fmt"

	"ego-server/internal/conversation/domain"
	writingdomain "ego-server/internal/writing/domain"
)

// DefaultChatGenerator is the MVP chat-generation policy.
type DefaultChatGenerator struct{}

func NewDefaultChatGenerator() DefaultChatGenerator {
	return DefaultChatGenerator{}
}

func (DefaultChatGenerator) GenerateOpening(_ context.Context, topic string, moments []writingdomain.Moment) (string, []domain.MomentReference, error) {
	refs := buildConversationRefs(moments)
	return fmt.Sprintf("嗨，我是那时的你。关于「%s」，那时候我写下了这些想法。你想聊些什么？", topic), refs, nil
}

func (DefaultChatGenerator) GenerateReply(_ context.Context, input domain.GenerateReplyInput) (*domain.GenerateReplyOutput, error) {
	refs := buildConversationRefs(input.ContextMoments)
	return &domain.GenerateReplyOutput{
		Content:           "嗯，我明白你的感受。那时候的我也是这样的，有些事说出来就好多了。",
		ReferencedMoments: refs,
	}, nil
}

func buildConversationRefs(moments []writingdomain.Moment) []domain.MomentReference {
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

// Compile-time check.
var _ domain.ChatGenerator = DefaultChatGenerator{}
