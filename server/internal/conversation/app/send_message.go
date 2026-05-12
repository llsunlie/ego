package app

import (
	"context"
	"fmt"
	"time"

	"ego-server/internal/conversation/domain"
	"ego-server/internal/platform/logging"
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
	logger := logging.FromContext(ctx)
	logger.DebugContext(ctx, "SendMessage: start", "session_id", input.ChatSessionID, "content_len", len([]rune(input.Content)))

	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user_id not found in context")
	}

	session, err := uc.sessions.FindByID(ctx, input.ChatSessionID)
	if err != nil {
		logger.ErrorContext(ctx, "SendMessage: find session failed", "session_id", input.ChatSessionID, "error", err)
		return nil, fmt.Errorf("find session: %w", err)
	}
	if session.UserID != userID {
		logger.WarnContext(ctx, "SendMessage: session ownership mismatch", "session_id", input.ChatSessionID, "session_user", session.UserID, "caller_user", userID)
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
		logger.ErrorContext(ctx, "SendMessage: create user message failed", "session_id", session.ID, "error", err)
		return nil, fmt.Errorf("create user message: %w", err)
	}
	logger.DebugContext(ctx, "SendMessage: user message saved", "message_id", userMsg.ID, "session_id", session.ID)

	star, err := uc.stars.FindByID(ctx, session.StarID)
	if err != nil {
		logger.ErrorContext(ctx, "SendMessage: find star failed", "star_id", session.StarID, "error", err)
		return nil, fmt.Errorf("find star: %w", err)
	}

	var contextMoments []writingdomain.Moment
	if star.TraceID != "" {
		contextMoments, err = uc.moments.FindByTraceID(ctx, star.TraceID)
		if err != nil {
			logger.ErrorContext(ctx, "SendMessage: find moments by trace failed", "trace_id", star.TraceID, "error", err)
			return nil, fmt.Errorf("find moments by trace: %w", err)
		}
	}
	logger.DebugContext(ctx, "SendMessage: context loaded", "star_id", star.ID, "topic", star.Topic, "trace_id", star.TraceID, "moment_count", len(contextMoments))

	history, err := uc.messages.ListBySessionID(ctx, session.ID)
	if err != nil {
		logger.ErrorContext(ctx, "SendMessage: list history failed", "session_id", session.ID, "error", err)
		return nil, fmt.Errorf("list history: %w", err)
	}
	logger.DebugContext(ctx, "SendMessage: history loaded", "session_id", session.ID, "history_len", len(history))

	logger.DebugContext(ctx, "SendMessage: generating reply", "session_id", session.ID, "topic", star.Topic, "history_len", len(history), "moment_count", len(contextMoments))
	replyOut, err := uc.chatGen.GenerateReply(ctx, domain.GenerateReplyInput{
		StarTopic:      star.Topic,
		ContextMoments: contextMoments,
		History:        history,
		UserMessage:    input.Content,
	})
	if err != nil {
		logger.ErrorContext(ctx, "SendMessage: generate reply failed", "session_id", session.ID, "error", err)
		return nil, fmt.Errorf("generate reply: %w", err)
	}
	logger.DebugContext(ctx, "SendMessage: reply generated", "session_id", session.ID, "reply_len", len([]rune(replyOut.Content)), "ref_count", len(replyOut.ReferencedMoments))

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
		logger.ErrorContext(ctx, "SendMessage: create reply message failed", "session_id", session.ID, "error", err)
		return nil, fmt.Errorf("create reply message: %w", err)
	}

	logger.InfoContext(ctx, "SendMessage: done", "session_id", session.ID, "user_msg_id", userMsg.ID, "reply_msg_id", reply.ID, "reply_len", len([]rune(reply.Content)))
	return &SendMessageOutput{Reply: reply}, nil
}
