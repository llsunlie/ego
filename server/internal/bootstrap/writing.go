package bootstrap

import (
	"context"
	"fmt"

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

func (stubEmbeddingGenerator) Generate(_ context.Context, _ string) ([]float32, error) {
	return nil, nil
}

type stubEchoMatcher struct{}

func (stubEchoMatcher) Match(_ context.Context, _ *writingdomain.Moment, _ []writingdomain.Moment) (*writingdomain.Echo, error) {
	return nil, nil
}

type stubInsightGenerator struct{}

func (stubInsightGenerator) Generate(_ context.Context, _ string, _ string) (*writingdomain.Insight, error) {
	return nil, fmt.Errorf("insight generator not implemented")
}

func NewWritingHandler(p *Platform) pb.EgoServer {
	queries := sqlc.New(p.Pool)

	traceRepo := writingpostgres.NewTraceRepository(queries)
	momentRepo := writingpostgres.NewMomentRepository(queries)

	createMoment := writingapp.NewCreateMomentUseCase(
		traceRepo, momentRepo,
		stubEmbeddingGenerator{}, stubEchoMatcher{},
		uuidGenerator{},
	)

	generateInsight := writingapp.NewGenerateInsightUseCase(stubInsightGenerator{})

	return writinggrpc.NewHandler(createMoment, generateInsight)
}
