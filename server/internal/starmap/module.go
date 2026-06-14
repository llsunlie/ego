package starmap

import (
	platformai "ego-server/internal/platform/ai"
	platformes "ego-server/internal/platform/elasticsearch"
	"ego-server/internal/platform/postgres/sqlc"
	starmapai "ego-server/internal/starmap/adapter/ai"
	starmapes "ego-server/internal/starmap/adapter/elasticsearch"
	starmapgrpc "ego-server/internal/starmap/adapter/grpc"
	starmapid "ego-server/internal/starmap/adapter/id"
	starmappostgres "ego-server/internal/starmap/adapter/postgres"
	starmapapp "ego-server/internal/starmap/app"
	writingpostgres "ego-server/internal/writing/adapter/postgres"
)

// Deps contains process-level resources and external capabilities needed to
// assemble the starmap bounded context.
type Deps struct {
	DB                      sqlc.DBTX
	AIClient                *platformai.Client
	ESClient                *platformes.Client
	AIEmbeddingDim          int
	ConstellationSparseOn   bool
	ConstellationSparseTopK int
	ConstellationHybridRRFK int
}

// NewHandler wires the starmap module's adapters, application use cases, and
// gRPC handler.
func NewHandler(deps Deps) *starmapgrpc.Handler {
	queries := sqlc.New(deps.DB)

	starRepo := starmappostgres.NewStarRepository(queries)
	constellationRepo := starmappostgres.NewConstellationRepository(queries)
	traceStasher := starmappostgres.NewTraceStasher(queries)
	traceReader := writingpostgres.NewReader(queries)

	assetGen := starmapai.NewConstellationAssetGenerator(deps.AIClient)
	profileGen := starmapai.NewTraceProfileGenerator(deps.AIClient)
	borderlineJudge := starmapai.NewConstellationBorderlineJudge(deps.AIClient)
	profileRefiner := starmapai.NewConstellationProfileRefiner(deps.AIClient)
	profileRepo := starmappostgres.NewTraceProfileRepository(deps.DB, deps.AIEmbeddingDim)
	constellationProfileRepo := starmappostgres.NewConstellationProfileRepository(deps.DB, deps.AIEmbeddingDim)
	ids := starmapid.NewUUIDGenerator()

	stashTrace := starmapapp.NewStashTraceUseCaseWithTraceProfile(
		traceReader, traceStasher, starRepo, constellationRepo,
		assetGen,
		profileGen, borderlineJudge, profileRepo, constellationProfileRepo,
		ids,
	)
	stashTrace.UseConstellationProfileRefiner(profileRefiner)
	if deps.ConstellationSparseOn && deps.ESClient != nil {
		profileSearch := starmapes.NewConstellationProfileSearch(deps.ESClient, starmapes.DefaultConstellationProfileIndex)
		stashTrace.UseConstellationSparseSearch(profileSearch, profileSearch, deps.ConstellationSparseTopK, deps.ConstellationHybridRRFK)
	}
	listConstellations := starmapapp.NewListConstellationsUseCase(constellationRepo, starRepo)
	getConstellation := starmapapp.NewGetConstellationUseCase(
		constellationRepo, starRepo, traceReader,
	)

	return starmapgrpc.NewHandler(stashTrace, listConstellations, getConstellation)
}
