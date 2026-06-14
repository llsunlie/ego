package bootstrap

import (
	"ego-server/internal/writing"

	pb "ego-server/proto/ego"
)

func NewWritingHandler(p *Platform) pb.EgoServer {
	return writing.NewHandler(writing.Deps{
		DB:       p.Pool,
		AIClient: p.AIClient,
	})
}
