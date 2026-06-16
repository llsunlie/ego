package bootstrap

import (
	"ego-server/internal/writing"

	pb "ego-server/proto/ego"
)

func NewWritingHandler(p *Platform) pb.EgoServer {
	return writing.NewHandler(writing.Deps{
		DB:             p.Pool,
		AIClient:       p.AIClient,
		EmbeddingDim:   p.AIEmbeddingDim,
		EchoRecallTopK: p.EchoRecallTopK,
		EchoSparseOn:   p.EchoSparseOn,
		EchoSparseTopK: p.EchoSparseTopK,
		EchoHybridRRFK: p.EchoHybridRRFK,
	})
}
