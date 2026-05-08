package domain

import (
	"context"

	starmapdomain "ego-server/internal/starmap/domain"
	writingdomain "ego-server/internal/writing/domain"
)

type ChatSessionRepository interface {
	Create(ctx context.Context, session *ChatSession) error
	FindByID(ctx context.Context, id string) (*ChatSession, error)
}

type ChatMessageRepository interface {
	Create(ctx context.Context, msg *ChatMessage) error
	ListBySessionID(ctx context.Context, sessionID string) ([]ChatMessage, error)
}

type StarReader interface {
	FindByID(ctx context.Context, id string) (*starmapdomain.Star, error)
}

type MomentReader interface {
	FindByIDs(ctx context.Context, ids []string) ([]writingdomain.Moment, error)
}

type ChatGenerator interface {
	GenerateOpening(ctx context.Context, topic string, moments []writingdomain.Moment) (string, []MomentReference, error)
	GenerateReply(ctx context.Context, input GenerateReplyInput) (*GenerateReplyOutput, error)
}

type GenerateReplyInput struct {
	StarTopic      string
	ContextMoments []writingdomain.Moment
	History        []ChatMessage
	UserMessage    string
}

type GenerateReplyOutput struct {
	Content           string
	ReferencedMoments []MomentReference
}
