package bootstrap

import (
	"context"

	writingapp "ego-server/internal/writing/app"
	writingdomain "ego-server/internal/writing/domain"
	writinggrpc "ego-server/internal/writing/adapter/grpc"
	writingpostgres "ego-server/internal/writing/adapter/postgres"
	"ego-server/internal/platform/postgres/sqlc"

	"github.com/google/uuid"

	pb "ego-server/proto/ego"
)

type uuidGenerator struct{}

func (uuidGenerator) New() string {
	return uuid.New().String()
}

type stubEmbeddingGenerator struct{}

func (stubEmbeddingGenerator) Generate(_ context.Context, _ string) ([]writingdomain.EmbeddingEntry, error) {
	return []writingdomain.EmbeddingEntry{
		{Model: "stub", Embedding: []float32{0.1, 0.2, 0.3}},
	}, nil
}

type stubEchoMatcher struct{}

func (stubEchoMatcher) Match(_ context.Context, _ *writingdomain.Moment, history []writingdomain.Moment) ([]writingdomain.MatchedMoment, error) {
	if len(history) == 0 {
		return nil, nil
	}
	matches := make([]writingdomain.MatchedMoment, 0, len(history))
	for _, h := range history {
		matches = append(matches, writingdomain.MatchedMoment{
			MomentID:   h.ID,
			Similarity: 0.85,
		})
	}
	return matches, nil
}

type stubInsightGenerator struct{}

func (stubInsightGenerator) Generate(_ context.Context, momentID, _ string) (*writingdomain.Insight, error) {
	return &writingdomain.Insight{
		MomentID:         momentID,
		Text:             "你似乎在反复思考与自尊相关的话题。当你感到被否定时，童年时期形成的防御模式会被激活。",
		RelatedMomentIDs: []string{},
	}, nil
}

func NewWritingHandler(p *Platform) pb.EgoServer {
	queries := sqlc.New(p.Pool)

	traceRepo := writingpostgres.NewTraceRepository(queries)
	momentRepo := writingpostgres.NewMomentRepository(queries)
	echoRepo := writingpostgres.NewEchoRepository(queries)
	insightRepo := writingpostgres.NewInsightRepository(queries)
	reader := writingpostgres.NewReader(queries)

	createMoment := writingapp.NewCreateMomentUseCase(
		traceRepo, momentRepo, echoRepo,
		stubEmbeddingGenerator{}, stubEchoMatcher{},
		uuidGenerator{},
	)

	generateInsight := writingapp.NewGenerateInsightUseCase(
		insightRepo, stubInsightGenerator{}, uuidGenerator{},
	)

	return writinggrpc.NewHandler(createMoment, generateInsight, reader)
}
