package bootstrap

import (
	"context"

	"ego-server/internal/writing"
	writingdomain "ego-server/internal/writing/domain"

	pb "ego-server/proto/ego"
)

type stubEmbeddingGenerator struct{}

func (stubEmbeddingGenerator) Generate(_ context.Context, _ string) ([]writingdomain.EmbeddingEntry, error) {
	return []writingdomain.EmbeddingEntry{
		{Model: "stub", Embedding: []float32{0.1, 0.2, 0.3}},
	}, nil
}

func NewWritingHandler(p *Platform) pb.EgoServer {
	return writing.NewHandler(writing.Deps{
		DB:                 p.Pool,
		EmbeddingGenerator: stubEmbeddingGenerator{},
	})
}
