package app

import (
	"context"

	"ego-server/internal/starmap/domain"
	writingdomain "ego-server/internal/writing/domain"
)

// DefaultTopicGenerator is the MVP topic-generation policy. It uses the
// leading characters of the first moment to produce a short summary topic.
type DefaultTopicGenerator struct{}

func NewDefaultTopicGenerator() DefaultTopicGenerator {
	return DefaultTopicGenerator{}
}

func (DefaultTopicGenerator) Generate(_ context.Context, moments []writingdomain.Moment) (string, error) {
	if len(moments) > 0 {
		content := []rune(moments[0].Content)
		if len(content) > 20 {
			content = content[:20]
		}
		return "关于" + string(content) + "…", nil
	}
	return "未命名的星", nil
}

// Compile-time check.
var _ domain.TopicGenerator = DefaultTopicGenerator{}
