package app

import (
	"context"
	"fmt"
	"time"

	"ego-server/internal/conversation/domain"
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
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, fmt.Errorf("user_id not found in context")
	}

	if input.ChatSessionID != "" {
		return uc.resumeSession(ctx, userID, input.ChatSessionID)
	}

	return uc.newSession(ctx, userID, input.StarID)
}

func (uc *StartChatUseCase) resumeSession(ctx context.Context, userID, sessionID string) (*StartChatOutput, error) {
	session, err := uc.sessions.FindByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("find session: %w", err)
	}
	if session.UserID != userID {
		return nil, domain.ErrChatSessionNotFound
	}

	history, err := uc.messages.ListBySessionID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("list messages: %w", err)
	}

	var opening *domain.ChatMessage
	for i := len(history) - 1; i >= 0; i-- {
		if history[i].Role == "past_self" {
			opening = &history[i]
			break
		}
	}

	return &StartChatOutput{
		Session: session,
		Opening: opening,
		History: history,
	}, nil
}

func (uc *StartChatUseCase) newSession(ctx context.Context, userID, starID string) (*StartChatOutput, error) {
	star, err := uc.stars.FindByID(ctx, starID)
	if err != nil {
		return nil, fmt.Errorf("find star: %w", err)
	}
	if star.UserID != userID {
		return nil, domain.ErrStarNotFound
	}

	var contextMoments []writingdomain.Moment
	var contextMomentIDs []string
	if star.TraceID != "" {
		contextMoments, err = uc.moments.FindByTraceID(ctx, star.TraceID)
		if err != nil {
			return nil, fmt.Errorf("find moments by trace: %w", err)
		}
		contextMomentIDs = make([]string, len(contextMoments))
		for i, m := range contextMoments {
			contextMomentIDs[i] = m.ID
		}
	}

	now := time.Now()
	session := &domain.ChatSession{
		ID:        uc.ids.New(),
		UserID:    userID,
		StarID:    starID,
		CreatedAt: now,
	}

	if err := uc.sessions.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	content, refs, err := uc.chatGen.GenerateOpening(ctx, star.Topic, contextMoments)
	if err != nil {
		return nil, fmt.Errorf("generate opening: %w", err)
	}

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
		return nil, fmt.Errorf("create opening message: %w", err)
	}

	return &StartChatOutput{
		Session: session,
		Opening: opening,
		History: []domain.ChatMessage{*opening},
	}, nil
}
