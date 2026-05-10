package writing

import (
	"ego-server/internal/platform/postgres/sqlc"
	writinggrpc "ego-server/internal/writing/adapter/grpc"
	writingid "ego-server/internal/writing/adapter/id"
	writingpostgres "ego-server/internal/writing/adapter/postgres"
	writingapp "ego-server/internal/writing/app"
	writingdomain "ego-server/internal/writing/domain"
)

// Deps contains process-level resources and external capabilities needed to
// assemble the writing bounded context.
type Deps struct {
	DB                 sqlc.DBTX
	EmbeddingGenerator writingdomain.EmbeddingGenerator
}

// NewHandler wires the writing module's adapters and use cases.
func NewHandler(deps Deps) *writinggrpc.Handler {
	queries := sqlc.New(deps.DB)

	traceRepo := writingpostgres.NewTraceRepository(queries)
	momentRepo := writingpostgres.NewMomentRepository(queries)
	echoRepo := writingpostgres.NewEchoRepository(queries)
	insightRepo := writingpostgres.NewInsightRepository(queries)
	reader := writingpostgres.NewReader(queries)
	ids := writingid.NewUUIDGenerator()
	echoMatcher := writingapp.NewDefaultEchoMatcher()
	insightGenerator := writingapp.NewDefaultInsightGenerator()

	createMoment := writingapp.NewCreateMomentUseCase(
		traceRepo, momentRepo, echoRepo,
		deps.EmbeddingGenerator, echoMatcher,
		ids,
	)

	generateInsight := writingapp.NewGenerateInsightUseCase(
		insightRepo, insightGenerator, ids,
	)

	return writinggrpc.NewHandler(createMoment, generateInsight, reader)
}
