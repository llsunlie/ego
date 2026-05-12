package bootstrap

import (
	"ego-server/internal/conversation"

	pb "ego-server/proto/ego"
)

func NewChatHandler(p *Platform) pb.EgoServer {
	return conversation.NewHandler(conversation.Deps{
		DB:       p.Pool,
		AIClient: p.AIClient,
	})
}
