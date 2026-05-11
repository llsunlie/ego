package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"ego-server/internal/conversation/domain"
	starmapdomain "ego-server/internal/starmap/domain"
	writingdomain "ego-server/internal/writing/domain"
)

func TestSendMessage_Success(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user_id", "user-1")
	now := time.Now()

	savedMessages := []domain.ChatMessage{}

	sessionRepo := &mockSessionRepo{
		findByIDFn: func(ctx context.Context, id string) (*domain.ChatSession, error) {
			return &domain.ChatSession{
				ID: "session-1", UserID: "user-1", StarID: "star-1",
			}, nil
		},
	}

	messageRepo := &mockMessageRepo{
		createFn: func(ctx context.Context, msg *domain.ChatMessage) error {
			savedMessages = append(savedMessages, *msg)
			return nil
		},
		listBySessionIDFn: func(ctx context.Context, sessionID string) ([]domain.ChatMessage, error) {
			return savedMessages, nil
		},
	}

	starReader := &mockStarReader{
		findByIDFn: func(ctx context.Context, id string) (*starmapdomain.Star, error) {
			return &starmapdomain.Star{Topic: "关于情绪管理"}, nil
		},
	}

	momentReader := &mockMomentReader{
		findByIDsFn: func(ctx context.Context, ids []string) ([]writingdomain.Moment, error) {
			return []writingdomain.Moment{
				{ID: "mom-1", Content: "今天心情不好", CreatedAt: now},
			}, nil
		},
	}

	chatGen := &mockChatGenerator{
		generateReplyFn: func(ctx context.Context, input domain.GenerateReplyInput) (*domain.GenerateReplyOutput, error) {
			return &domain.GenerateReplyOutput{
				Content:           "嗯，我明白你的感受",
				ReferencedMoments: []domain.MomentReference{{Date: "5月8日", Snippet: "今天心情不好"}},
			}, nil
		},
	}

	ids := mockIDGenerator{id: "msg-1"}
	uc := NewSendMessageUseCase(sessionRepo, messageRepo, starReader, momentReader, chatGen, ids)

	out, err := uc.Execute(ctx, SendMessageInput{
		ChatSessionID: "session-1",
		Content:       "你好呀",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Reply == nil {
		t.Fatal("expected reply, got nil")
	}
	if out.Reply.Content != "嗯，我明白你的感受" {
		t.Fatalf("expected reply content, got %q", out.Reply.Content)
	}
	if out.Reply.Role != "past_self" {
		t.Fatalf("expected role 'past_self', got %q", out.Reply.Role)
	}

	if len(savedMessages) != 2 {
		t.Fatalf("expected 2 messages saved (user + reply), got %d", len(savedMessages))
	}
	if savedMessages[0].Role != "user" {
		t.Fatalf("expected first saved message to be user, got %q", savedMessages[0].Role)
	}
}

func TestSendMessage_SessionNotFound(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user_id", "user-1")

	sessionRepo := &mockSessionRepo{
		findByIDFn: func(ctx context.Context, id string) (*domain.ChatSession, error) {
			return nil, domain.ErrChatSessionNotFound
		},
	}

	uc := NewSendMessageUseCase(sessionRepo, nil, nil, nil, nil, nil)

	_, err := uc.Execute(ctx, SendMessageInput{ChatSessionID: "nonexistent"})
	if !errors.Is(err, domain.ErrChatSessionNotFound) {
		t.Fatalf("expected ErrChatSessionNotFound, got %v", err)
	}
}

func TestSendMessage_WrongUser(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user_id", "user-1")

	sessionRepo := &mockSessionRepo{
		findByIDFn: func(ctx context.Context, id string) (*domain.ChatSession, error) {
			return &domain.ChatSession{
				ID: "session-1", UserID: "user-2",
			}, nil
		},
	}

	uc := NewSendMessageUseCase(sessionRepo, nil, nil, nil, nil, nil)

	_, err := uc.Execute(ctx, SendMessageInput{ChatSessionID: "session-1"})
	if !errors.Is(err, domain.ErrChatSessionNotFound) {
		t.Fatalf("expected ErrChatSessionNotFound, got %v", err)
	}
}
