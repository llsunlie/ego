package grpc

import (
	"ego-server/internal/conversation/domain"

	pb "ego-server/proto/ego"
)

func chatMessageToProto(m domain.ChatMessage) *pb.ChatMessage {
	role := pb.ChatRole_USER
	if m.Role == "past_self" {
		role = pb.ChatRole_PAST_SELF
	}

	refs := make([]*pb.MomentReference, len(m.ReferencedMoments))
	for i, r := range m.ReferencedMoments {
		refs[i] = &pb.MomentReference{
			Date:    r.Date,
			Snippet: r.Snippet,
		}
	}

	return &pb.ChatMessage{
		Id:         m.ID,
		Role:       role,
		Content:    m.Content,
		Referenced: refs,
		Timestamp:  m.CreatedAt.UnixMilli(),
	}
}
