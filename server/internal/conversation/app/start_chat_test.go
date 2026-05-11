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

// --- mocks ---

type mockSessionRepo struct {
	createFn   func(ctx context.Context, session *domain.ChatSession) error
	findByIDFn func(ctx context.Context, id string) (*domain.ChatSession, error)
}

func (m *mockSessionRepo) Create(ctx context.Context, session *domain.ChatSession) error {
	return m.createFn(ctx, session)
}
func (m *mockSessionRepo) FindByID(ctx context.Context, id string) (*domain.ChatSession, error) {
	return m.findByIDFn(ctx, id)
}

type mockMessageRepo struct {
	createFn          func(ctx context.Context, msg *domain.ChatMessage) error
	listBySessionIDFn func(ctx context.Context, sessionID string) ([]domain.ChatMessage, error)
}

func (m *mockMessageRepo) Create(ctx context.Context, msg *domain.ChatMessage) error {
	return m.createFn(ctx, msg)
}
func (m *mockMessageRepo) ListBySessionID(ctx context.Context, sessionID string) ([]domain.ChatMessage, error) {
	return m.listBySessionIDFn(ctx, sessionID)
}

type mockStarReader struct {
	findByIDFn func(ctx context.Context, id string) (*starmapdomain.Star, error)
}

func (m *mockStarReader) FindByID(ctx context.Context, id string) (*starmapdomain.Star, error) {
	return m.findByIDFn(ctx, id)
}

type mockMomentReader struct {
	findByIDsFn      func(ctx context.Context, ids []string) ([]writingdomain.Moment, error)
	findByTraceIDFn  func(ctx context.Context, traceID string) ([]writingdomain.Moment, error)
}

func (m *mockMomentReader) FindByIDs(ctx context.Context, ids []string) ([]writingdomain.Moment, error) {
	return m.findByIDsFn(ctx, ids)
}
func (m *mockMomentReader) FindByTraceID(ctx context.Context, traceID string) ([]writingdomain.Moment, error) {
	if m.findByTraceIDFn != nil {
		return m.findByTraceIDFn(ctx, traceID)
	}
	return nil, nil
}

type mockChatGenerator struct {
	generateOpeningFn func(ctx context.Context, topic string, moments []writingdomain.Moment) (string, []domain.MomentReference, error)
	generateReplyFn   func(ctx context.Context, input domain.GenerateReplyInput) (*domain.GenerateReplyOutput, error)
}

func (m *mockChatGenerator) GenerateOpening(ctx context.Context, topic string, moments []writingdomain.Moment) (string, []domain.MomentReference, error) {
	return m.generateOpeningFn(ctx, topic, moments)
}
func (m *mockChatGenerator) GenerateReply(ctx context.Context, input domain.GenerateReplyInput) (*domain.GenerateReplyOutput, error) {
	return m.generateReplyFn(ctx, input)
}

type mockIDGenerator struct {
	id string
}

func (m mockIDGenerator) New() string { return m.id }

// --- tests ---

func TestStartChat_NewSession(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user_id", "user-1")
	now := time.Now()

	starReader := &mockStarReader{
		findByIDFn: func(ctx context.Context, id string) (*starmapdomain.Star, error) {
			return &starmapdomain.Star{ID: "star-1", UserID: "user-1", Topic: "关于情绪管理"}, nil
		},
	}

	momentReader := &mockMomentReader{
		findByIDsFn: func(ctx context.Context, ids []string) ([]writingdomain.Moment, error) {
			return []writingdomain.Moment{
				{ID: "mom-1", Content: "今天心情不好", CreatedAt: now},
			}, nil
		},
	}

	sessionRepo := &mockSessionRepo{
		createFn: func(ctx context.Context, session *domain.ChatSession) error {
			return nil
		},
	}

	messageRepo := &mockMessageRepo{
		createFn: func(ctx context.Context, msg *domain.ChatMessage) error {
			return nil
		},
	}

	chatGen := &mockChatGenerator{
		generateOpeningFn: func(ctx context.Context, topic string, moments []writingdomain.Moment) (string, []domain.MomentReference, error) {
			return "嗨，我是那时的你", []domain.MomentReference{{Date: "5月8日", Snippet: "今天心情不好"}}, nil
		},
	}

	ids := mockIDGenerator{id: "session-1"}
	uc := NewStartChatUseCase(sessionRepo, messageRepo, starReader, momentReader, chatGen, ids)

	out, err := uc.Execute(ctx, StartChatInput{
		StarID:           "star-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Session.ID != "session-1" {
		t.Fatalf("expected session id 'session-1', got %q", out.Session.ID)
	}
	if out.Opening == nil {
		t.Fatal("expected opening message, got nil")
	}
	if out.Opening.Content != "嗨，我是那时的你" {
		t.Fatalf("expected opening content, got %q", out.Opening.Content)
	}
	if out.Opening.Role != "past_self" {
		t.Fatalf("expected role 'past_self', got %q", out.Opening.Role)
	}
	if len(out.History) != 1 {
		t.Fatalf("expected 1 history message, got %d", len(out.History))
	}
}

func TestStartChat_ResumeSession(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user_id", "user-1")

	sessionRepo := &mockSessionRepo{
		findByIDFn: func(ctx context.Context, id string) (*domain.ChatSession, error) {
			return &domain.ChatSession{
				ID: "session-1", UserID: "user-1", StarID: "star-1",
			}, nil
		},
	}

	messageRepo := &mockMessageRepo{
		listBySessionIDFn: func(ctx context.Context, sessionID string) ([]domain.ChatMessage, error) {
			return []domain.ChatMessage{
				{ID: "msg-1", Role: "past_self", Content: "嗨，我是那时的你"},
				{ID: "msg-2", Role: "user", Content: "你好呀"},
				{ID: "msg-3", Role: "past_self", Content: "你想聊什么？"},
			}, nil
		},
	}

	uc := NewStartChatUseCase(sessionRepo, messageRepo, nil, nil, nil, nil)

	out, err := uc.Execute(ctx, StartChatInput{ChatSessionID: "session-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Session.ID != "session-1" {
		t.Fatalf("expected session id 'session-1', got %q", out.Session.ID)
	}
	if out.Opening == nil {
		t.Fatal("expected opening (last AI message), got nil")
	}
	if out.Opening.Content != "你想聊什么？" {
		t.Fatalf("expected last AI message, got %q", out.Opening.Content)
	}
	if len(out.History) != 3 {
		t.Fatalf("expected 3 history messages, got %d", len(out.History))
	}
}

func TestStartChat_StarNotFound(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user_id", "user-1")

	starReader := &mockStarReader{
		findByIDFn: func(ctx context.Context, id string) (*starmapdomain.Star, error) {
			return nil, domain.ErrStarNotFound
		},
	}

	uc := NewStartChatUseCase(nil, nil, starReader, nil, nil, nil)

	_, err := uc.Execute(ctx, StartChatInput{StarID: "nonexistent"})
	if !errors.Is(err, domain.ErrStarNotFound) {
		t.Fatalf("expected ErrStarNotFound, got %v", err)
	}
}

func TestStartChat_WrongUser(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user_id", "user-1")

	sessionRepo := &mockSessionRepo{
		findByIDFn: func(ctx context.Context, id string) (*domain.ChatSession, error) {
			return &domain.ChatSession{
				ID: "session-1", UserID: "user-2",
			}, nil
		},
	}

	uc := NewStartChatUseCase(sessionRepo, nil, nil, nil, nil, nil)

	_, err := uc.Execute(ctx, StartChatInput{ChatSessionID: "session-1"})
	if !errors.Is(err, domain.ErrChatSessionNotFound) {
		t.Fatalf("expected ErrChatSessionNotFound, got %v", err)
	}
}

func TestStartChat_NewSession_StarWrongUser(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user_id", "user-1")

	starReader := &mockStarReader{
		findByIDFn: func(ctx context.Context, id string) (*starmapdomain.Star, error) {
			return &starmapdomain.Star{ID: "star-1", UserID: "user-2"}, nil
		},
	}

	uc := NewStartChatUseCase(nil, nil, starReader, nil, nil, nil)

	_, err := uc.Execute(ctx, StartChatInput{StarID: "star-1"})
	if !errors.Is(err, domain.ErrStarNotFound) {
		t.Fatalf("expected ErrStarNotFound for cross-user star, got %v", err)
	}
}

func TestStartChat_ResumeSession_NotFound(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user_id", "user-1")

	sessionRepo := &mockSessionRepo{
		findByIDFn: func(ctx context.Context, id string) (*domain.ChatSession, error) {
			return nil, domain.ErrChatSessionNotFound
		},
	}

	uc := NewStartChatUseCase(sessionRepo, nil, nil, nil, nil, nil)

	_, err := uc.Execute(ctx, StartChatInput{ChatSessionID: "nonexistent"})
	if !errors.Is(err, domain.ErrChatSessionNotFound) {
		t.Fatalf("expected ErrChatSessionNotFound, got %v", err)
	}
}

func TestStartChat_NewSession_EmptyContextMoments(t *testing.T) {
	ctx := context.WithValue(context.Background(), "user_id", "user-1")

	starReader := &mockStarReader{
		findByIDFn: func(ctx context.Context, id string) (*starmapdomain.Star, error) {
			return &starmapdomain.Star{ID: "star-1", UserID: "user-1", Topic: "关于情绪管理"}, nil
		},
	}

	momentReader := &mockMomentReader{
		findByIDsFn: func(ctx context.Context, ids []string) ([]writingdomain.Moment, error) {
			return nil, nil
		},
	}

	sessionRepo := &mockSessionRepo{
		createFn: func(ctx context.Context, session *domain.ChatSession) error { return nil },
	}

	messageRepo := &mockMessageRepo{
		createFn: func(ctx context.Context, msg *domain.ChatMessage) error { return nil },
	}

	chatGen := &mockChatGenerator{
		generateOpeningFn: func(ctx context.Context, topic string, moments []writingdomain.Moment) (string, []domain.MomentReference, error) {
			return "嗨，我是那时的你", nil, nil
		},
	}

	ids := mockIDGenerator{id: "session-1"}
	uc := NewStartChatUseCase(sessionRepo, messageRepo, starReader, momentReader, chatGen, ids)

	out, err := uc.Execute(ctx, StartChatInput{
		StarID:           "star-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Opening == nil {
		t.Fatal("expected opening, got nil")
	}
	if out.Opening.Content != "嗨，我是那时的你" {
		t.Fatalf("expected opening content, got %q", out.Opening.Content)
	}
}
