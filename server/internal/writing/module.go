package writing

import (
	platformai "ego-server/internal/platform/ai"
	platformes "ego-server/internal/platform/elasticsearch"
	"ego-server/internal/platform/postgres/sqlc"
	writingai "ego-server/internal/writing/adapter/ai"
	writinges "ego-server/internal/writing/adapter/elasticsearch"
	writinggrpc "ego-server/internal/writing/adapter/grpc"
	writingid "ego-server/internal/writing/adapter/id"
	writingpostgres "ego-server/internal/writing/adapter/postgres"
	writingapp "ego-server/internal/writing/app"
	"ego-server/internal/writing/domain"
)

// Deps contains process-level resources and external capabilities needed to
// assemble the writing bounded context.
type Deps struct {
	DB             sqlc.DBTX
	AIClient       *platformai.Client
	ESClient       *platformes.Client
	EmbeddingDim   int
	EchoRecallTopK int32
	EchoSparseOn   bool
	EchoSparseTopK int32
	EchoHybridRRFK int
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
	var searchIndexer domain.MomentSearchIndexer
	var sparseReader domain.EchoSparseCandidateReader
	if deps.EchoSparseOn && deps.ESClient != nil {
		momentSearch := writinges.NewMomentSearch(deps.ESClient, writinges.DefaultMomentIndex)
		searchIndexer = momentSearch
		sparseReader = momentSearch
	}

	createMoment := writingapp.NewCreateMomentUseCaseWithHybridCandidates(
		traceRepo, momentRepo, echoCandidateReader,
		searchIndexer, sparseReader,
		echoRepo, embedder, echoMatcher,
		ids, deps.EchoRecallTopK, deps.EchoSparseTopK, deps.EchoHybridRRFK,
	)

	generateInsight := writingapp.NewGenerateInsightUseCase(
		insightRepo, insightGenerator, ids,
	)

	return writinggrpc.NewHandler(createMoment, generateInsight, reader)
}
