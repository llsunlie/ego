package app

import (
	"context"
	"fmt"
	"time"

	"ego-server/internal/conversation/domain"
	writingdomain "ego-server/internal/writing/domain"
)

type SendMessageUseCase struct {
	sessions domain.ChatSessionRepository
	messages domain.ChatMessageRepository
	stars    domain.StarReader
	moments  domain.MomentReader
	chatGen  domain.ChatGenerator
	ids      IDGenerator
}

func NewSendMessageUseCase(
	sessions domain.ChatSessionRepository,
	messages domain.ChatMessageRepository,
	stars domain.StarReader,
	moments domain.MomentReader,
	chatGen domain.ChatGenerator,
	ids IDGenerator,
) *SendMessageUseCase {
	return &SendMessageUseCase{
		sessions: sessions,
		messages: messages,
		stars:    stars,
		moments:  moments,
		chatGen:  chatGen,
		ids:      ids,
	}
}

type SendMessageInput struct {
	ChatSessionID string
	Content       string
}

type SendMessageOutput struct {
	Reply *domain.ChatMessage
}

func (uc *SendMessageUseCase) Execute(ctx context.Context, input SendMessageInput) (*SendMessageOutput, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user_id not found in context")
	}

	session, err := uc.sessions.FindByID(ctx, input.ChatSessionID)
	if err != nil {
		return nil, fmt.Errorf("find session: %w", err)
	}
	if session.UserID != userID {
		return nil, domain.ErrChatSessionNotFound
	}

	now := time.Now()
	userMsg := &domain.ChatMessage{
		ID:                uc.ids.New(),
		UserID:            userID,
		SessionID:         session.ID,
		Role:              "user",
		Content:           input.Content,
		ReferencedMoments: nil,
		CreatedAt:         now,
	}

	if err := uc.messages.Create(ctx, userMsg); err != nil {
		return nil, fmt.Errorf("create user message: %w", err)
	}

	star, err := uc.stars.FindByID(ctx, session.StarID)
	if err != nil {
		return nil, fmt.Errorf("find star: %w", err)
	}

	var contextMoments []writingdomain.Moment
	if star.TraceID != "" {
		contextMoments, err = uc.moments.FindByTraceID(ctx, star.TraceID)
		if err != nil {
			return nil, fmt.Errorf("find moments by trace: %w", err)
		}
	}

	history, err := uc.messages.ListBySessionID(ctx, session.ID)
	if err != nil {
		return nil, fmt.Errorf("list history: %w", err)
	}

	replyOut, err := uc.chatGen.GenerateReply(ctx, domain.GenerateReplyInput{
		StarTopic:      star.Topic,
		ContextMoments: contextMoments,
		History:        history,
		UserMessage:    input.Content,
	})
	if err != nil {
		return nil, fmt.Errorf("generate reply: %w", err)
	}

	reply := &domain.ChatMessage{
		ID:                uc.ids.New(),
		UserID:            userID,
		SessionID:         session.ID,
		Role:              "past_self",
		Content:           replyOut.Content,
		ReferencedMoments: replyOut.ReferencedMoments,
		CreatedAt:         time.Now(),
	}

	if err := uc.messages.Create(ctx, reply); err != nil {
		return nil, fmt.Errorf("create reply message: %w", err)
	}

	return &SendMessageOutput{Reply: reply}, nil
}
