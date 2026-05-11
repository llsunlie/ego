package grpc

import (
	"context"

	"ego-server/internal/conversation/app"

	pb "ego-server/proto/ego"
)

type StartChatUseCase interface {
	Execute(ctx context.Context, input app.StartChatInput) (*app.StartChatOutput, error)
}

type SendMessageUseCase interface {
	Execute(ctx context.Context, input app.SendMessageInput) (*app.SendMessageOutput, error)
}

type Handler struct {
	pb.UnimplementedEgoServer
	startChat   StartChatUseCase
	sendMessage SendMessageUseCase
}

func NewHandler(startChat StartChatUseCase, sendMessage SendMessageUseCase) *Handler {
	return &Handler{
		startChat:   startChat,
		sendMessage: sendMessage,
	}
}

func (h *Handler) StartChat(ctx context.Context, req *pb.StartChatReq) (*pb.StartChatRes, error) {
	out, err := h.startChat.Execute(ctx, app.StartChatInput{
		StarID:        req.StarId,
		ChatSessionID: req.ChatSessionId,
	})
	if err != nil {
		return nil, err
	}

	history := make([]*pb.ChatMessage, len(out.History))
	for i, m := range out.History {
		history[i] = chatMessageToProto(m)
	}

	var opening *pb.ChatMessage
	if out.Opening != nil {
		opening = chatMessageToProto(*out.Opening)
	}

	return &pb.StartChatRes{
		ChatSessionId: out.Session.ID,
		Opening:       opening,
		History:       history,
	}, nil
}

func (h *Handler) SendMessage(ctx context.Context, req *pb.SendMessageReq) (*pb.SendMessageRes, error) {
	out, err := h.sendMessage.Execute(ctx, app.SendMessageInput{
		ChatSessionID: req.ChatSessionId,
		Content:       req.Content,
	})
	if err != nil {
		return nil, err
	}

	return &pb.SendMessageRes{
		Reply: chatMessageToProto(*out.Reply),
	}, nil
}
