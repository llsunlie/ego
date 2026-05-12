package app

import (
	"context"
	"fmt"
	"time"

	"ego-server/internal/conversation/domain"
	"ego-server/internal/platform/logging"
	writingdomain "ego-server/internal/writing/domain"
)

type StartChatUseCase struct {
	sessions domain.ChatSessionRepository
	messages domain.ChatMessageRepository
	stars    domain.StarReader
	moments  domain.MomentReader
	chatGen  domain.ChatGenerator
	ids      IDGenerator
}

func NewStartChatUseCase(
	sessions domain.ChatSessionRepository,
	messages domain.ChatMessageRepository,
	stars domain.StarReader,
	moments domain.MomentReader,
	chatGen domain.ChatGenerator,
	ids IDGenerator,
) *StartChatUseCase {
	return &StartChatUseCase{
		sessions: sessions,
		messages: messages,
		stars:    stars,
		moments:  moments,
		chatGen:  chatGen,
		ids:      ids,
	}
}

type StartChatInput struct {
	StarID        string
	ChatSessionID string // optional: resume existing session
}

type StartChatOutput struct {
	Session *domain.ChatSession
	Opening *domain.ChatMessage
	History []domain.ChatMessage
}

func (uc *StartChatUseCase) Execute(ctx context.Context, input StartChatInput) (*StartChatOutput, error) {
	logger := logging.FromContext(ctx)

	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user_id not found in context")
	}

	if input.ChatSessionID != "" {
		logger.DebugContext(ctx, "StartChat: resuming session", "session_id", input.ChatSessionID, "star_id", input.StarID)
		out, err := uc.resumeSession(ctx, userID, input.ChatSessionID)
		if err != nil {
			logger.ErrorContext(ctx, "StartChat: resume failed", "session_id", input.ChatSessionID, "error", err)
			return nil, err
		}
		logger.InfoContext(ctx, "StartChat: session resumed", "session_id", out.Session.ID, "history_len", len(out.History))
		return out, nil
	}

	logger.DebugContext(ctx, "StartChat: creating new session", "star_id", input.StarID)
	out, err := uc.newSession(ctx, userID, input.StarID)
	if err != nil {
		logger.ErrorContext(ctx, "StartChat: new session failed", "star_id", input.StarID, "error", err)
		return nil, err
	}
	logger.InfoContext(ctx, "StartChat: new session created", "session_id", out.Session.ID, "star_id", input.StarID, "opening_len", len([]rune(out.Opening.Content)))
	return out, nil
}

func (uc *StartChatUseCase) resumeSession(ctx context.Context, userID, sessionID string) (*StartChatOutput, error) {
	logger := logging.FromContext(ctx)

	session, err := uc.sessions.FindByID(ctx, sessionID)
	if err != nil {
		logger.ErrorContext(ctx, "StartChat: find session failed", "session_id", sessionID, "error", err)
		return nil, fmt.Errorf("find session: %w", err)
	}
	if session.UserID != userID {
		logger.WarnContext(ctx, "StartChat: session ownership mismatch", "session_id", sessionID, "session_user", session.UserID, "caller_user", userID)
		return nil, domain.ErrChatSessionNotFound
	}

	history, err := uc.messages.ListBySessionID(ctx, sessionID)
	if err != nil {
		logger.ErrorContext(ctx, "StartChat: list messages failed", "session_id", sessionID, "error", err)
		return nil, fmt.Errorf("list messages: %w", err)
	}

	var opening *domain.ChatMessage
	for i := len(history) - 1; i >= 0; i-- {
		if history[i].Role == "past_self" {
			opening = &history[i]
			break
		}
	}

	logger.DebugContext(ctx, "StartChat: session resumed", "session_id", sessionID, "history_len", len(history), "has_opening", opening != nil)
	return &StartChatOutput{
		Session: session,
		Opening: opening,
		History: history,
	}, nil
}

func (uc *StartChatUseCase) newSession(ctx context.Context, userID, starID string) (*StartChatOutput, error) {
	logger := logging.FromContext(ctx)

	star, err := uc.stars.FindByID(ctx, starID)
	if err != nil {
		logger.ErrorContext(ctx, "StartChat: find star failed", "star_id", starID, "error", err)
		return nil, fmt.Errorf("find star: %w", err)
	}
	if star.UserID != userID {
		logger.WarnContext(ctx, "StartChat: star ownership mismatch", "star_id", starID, "star_user", star.UserID, "caller_user", userID)
		return nil, domain.ErrStarNotFound
	}

	var contextMoments []writingdomain.Moment
	if star.TraceID != "" {
		contextMoments, err = uc.moments.FindByTraceID(ctx, star.TraceID)
		if err != nil {
			logger.ErrorContext(ctx, "StartChat: find moments by trace failed", "trace_id", star.TraceID, "error", err)
			return nil, fmt.Errorf("find moments by trace: %w", err)
		}
	}
	logger.DebugContext(ctx, "StartChat: context loaded", "star_id", starID, "topic", star.Topic, "trace_id", star.TraceID, "moment_count", len(contextMoments))

	now := time.Now()
	session := &domain.ChatSession{
		ID:        uc.ids.New(),
		UserID:    userID,
		StarID:    starID,
		CreatedAt: now,
	}

	if err := uc.sessions.Create(ctx, session); err != nil {
		logger.ErrorContext(ctx, "StartChat: create session failed", "error", err)
		return nil, fmt.Errorf("create session: %w", err)
	}

	logger.DebugContext(ctx, "StartChat: generating opening", "session_id", session.ID, "topic", star.Topic, "moment_count", len(contextMoments))
	content, refs, err := uc.chatGen.GenerateOpening(ctx, star.Topic, contextMoments)
	if err != nil {
		logger.ErrorContext(ctx, "StartChat: generate opening failed", "session_id", session.ID, "error", err)
		return nil, fmt.Errorf("generate opening: %w", err)
	}
	logger.DebugContext(ctx, "StartChat: opening generated", "session_id", session.ID, "opening_len", len([]rune(content)), "ref_count", len(refs))

	opening := &domain.ChatMessage{
		ID:                uc.ids.New(),
		UserID:            userID,
		SessionID:         session.ID,
		Role:              "past_self",
		Content:           content,
		ReferencedMoments: refs,
		CreatedAt:         now,
	}

	if err := uc.messages.Create(ctx, opening); err != nil {
		logger.ErrorContext(ctx, "StartChat: create opening message failed", "session_id", session.ID, "error", err)
		return nil, fmt.Errorf("create opening message: %w", err)
	}

	logger.DebugContext(ctx, "StartChat: session ready", "session_id", session.ID, "opening_msg_id", opening.ID)
	return &StartChatOutput{
		Session: session,
		Opening: opening,
		History: []domain.ChatMessage{*opening},
	}, nil
}
