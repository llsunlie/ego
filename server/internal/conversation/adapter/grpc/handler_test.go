package grpc

import (
	"context"
	"errors"
	"testing"
	"time"

	"ego-server/internal/conversation/app"
	"ego-server/internal/conversation/domain"

	pb "ego-server/proto/ego"
)

func userCtx(userID string) context.Context {
	return context.WithValue(context.Background(), "user_id", userID)
}

// --- fake use cases ---

type fakeStartChatUC struct {
	out *app.StartChatOutput
	err error
}

func (f *fakeStartChatUC) Execute(ctx context.Context, input app.StartChatInput) (*app.StartChatOutput, error) {
	return f.out, f.err
}

type fakeSendMessageUC struct {
	out *app.SendMessageOutput
	err error
}

func (f *fakeSendMessageUC) Execute(ctx context.Context, input app.SendMessageInput) (*app.SendMessageOutput, error) {
	return f.out, f.err
}

// --- tests ---

func TestHandler_StartChat(t *testing.T) {
	now := time.Now()
	fake := &fakeStartChatUC{
		out: &app.StartChatOutput{
			Session: &domain.ChatSession{
				ID: "session-1", UserID: "user-1", StarID: "star-1",
			},
			Opening: &domain.ChatMessage{
				ID: "msg-1", UserID: "user-1", SessionID: "session-1",
				Role: "past_self", Content: "嗨，我是那时的你",
				ReferencedMoments: []domain.MomentReference{{Date: "5月8日", Snippet: "今天心情不好"}},
				CreatedAt: now,
			},
			History: []domain.ChatMessage{
				{ID: "msg-1", Role: "past_self", Content: "嗨，我是那时的你", CreatedAt: now},
			},
		},
	}
	sendUC := &fakeSendMessageUC{}
	h := NewHandler(fake, sendUC)

	res, err := h.StartChat(userCtx("user-1"), &pb.StartChatReq{
		StarId:           "star-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.ChatSessionId != "session-1" {
		t.Fatalf("expected session id 'session-1', got %q", res.ChatSessionId)
	}
	if res.Opening == nil {
		t.Fatal("expected opening, got nil")
	}
	if res.Opening.Content != "嗨，我是那时的你" {
		t.Fatalf("expected opening content, got %q", res.Opening.Content)
	}
	if res.Opening.Role != pb.ChatRole_PAST_SELF {
		t.Fatalf("expected PAST_SELF role, got %v", res.Opening.Role)
	}
	if len(res.Opening.Referenced) != 1 {
		t.Fatalf("expected 1 reference, got %d", len(res.Opening.Referenced))
	}
	if len(res.History) != 1 {
		t.Fatalf("expected 1 history, got %d", len(res.History))
	}
}

func TestHandler_StartChat_Error(t *testing.T) {
	fake := &fakeStartChatUC{err: domain.ErrStarNotFound}
	sendUC := &fakeSendMessageUC{}
	h := NewHandler(fake, sendUC)

	_, err := h.StartChat(userCtx("user-1"), &pb.StartChatReq{StarId: "nonexistent"})
	if !errors.Is(err, domain.ErrStarNotFound) {
		t.Fatalf("expected ErrStarNotFound, got %v", err)
	}
}

func TestHandler_SendMessage(t *testing.T) {
	now := time.Now()
	fake := &fakeSendMessageUC{
		out: &app.SendMessageOutput{
			Reply: &domain.ChatMessage{
				ID: "msg-2", UserID: "user-1", SessionID: "session-1",
				Role: "past_self", Content: "嗯，我明白你的感受",
				ReferencedMoments: []domain.MomentReference{{Date: "5月8日", Snippet: "今天心情不好"}},
				CreatedAt: now,
			},
		},
	}
	startUC := &fakeStartChatUC{}
	h := NewHandler(startUC, fake)

	res, err := h.SendMessage(userCtx("user-1"), &pb.SendMessageReq{
		ChatSessionId: "session-1",
		Content:       "你好呀",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Reply == nil {
		t.Fatal("expected reply, got nil")
	}
	if res.Reply.Content != "嗯，我明白你的感受" {
		t.Fatalf("expected reply content, got %q", res.Reply.Content)
	}
	if res.Reply.Role != pb.ChatRole_PAST_SELF {
		t.Fatalf("expected PAST_SELF role, got %v", res.Reply.Role)
	}
}

func TestHandler_SendMessage_Error(t *testing.T) {
	fake := &fakeSendMessageUC{err: domain.ErrChatSessionNotFound}
	startUC := &fakeStartChatUC{}
	h := NewHandler(startUC, fake)

	_, err := h.SendMessage(userCtx("user-1"), &pb.SendMessageReq{ChatSessionId: "nonexistent"})
	if !errors.Is(err, domain.ErrChatSessionNotFound) {
		t.Fatalf("expected ErrChatSessionNotFound, got %v", err)
	}
}
