package writing

import (
	platformai "ego-server/internal/platform/ai"
	"ego-server/internal/platform/postgres/sqlc"
	writingai "ego-server/internal/writing/adapter/ai"
	writinggrpc "ego-server/internal/writing/adapter/grpc"
	writingid "ego-server/internal/writing/adapter/id"
	writingpostgres "ego-server/internal/writing/adapter/postgres"
	writingapp "ego-server/internal/writing/app"
)

// Deps contains process-level resources and external capabilities needed to
// assemble the writing bounded context.
type Deps struct {
	DB             sqlc.DBTX
	AIClient       *platformai.Client
	EmbeddingDim   int
	EchoRecallTopK int32
}

// NewHandler wires the writing module's adapters and use cases.
func NewHandler(deps Deps) *writinggrpc.Handler {
	queries := sqlc.New(deps.DB)

	traceRepo := writingpostgres.NewTraceRepository(queries)
	momentRepo := writingpostgres.NewMomentRepositoryWithVector(queries, deps.DB, deps.EmbeddingDim)
	echoCandidateReader := writingpostgres.NewEchoCandidateReader(queries, deps.DB, deps.EmbeddingDim)
	echoRepo := writingpostgres.NewEchoRepository(queries)
	insightRepo := writingpostgres.NewInsightRepository(queries)
	reader := writingpostgres.NewReader(queries)
	ids := writingid.NewUUIDGenerator()

	embedder := writingai.NewEmbedder(deps.AIClient)
	echoMatcher := writingapp.NewDefaultEchoMatcher()
	insightGenerator := writingai.NewInsightGenerator(deps.AIClient, momentRepo, echoRepo)

	createMoment := writingapp.NewCreateMomentUseCaseWithCandidates(
		traceRepo, momentRepo, echoCandidateReader,
		echoRepo, embedder, echoMatcher,
		ids, deps.EchoRecallTopK,
	)

	generateInsight := writingapp.NewGenerateInsightUseCase(
		insightRepo, insightGenerator, ids,
	)

	return writinggrpc.NewHandler(createMoment, generateInsight, reader)
}
