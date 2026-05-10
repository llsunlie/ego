package starmap

import (
	"ego-server/internal/platform/postgres/sqlc"
	starmapgrpc "ego-server/internal/starmap/adapter/grpc"
	starmapid "ego-server/internal/starmap/adapter/id"
	starmappostgres "ego-server/internal/starmap/adapter/postgres"
	starmapapp "ego-server/internal/starmap/app"
	writingpostgres "ego-server/internal/writing/adapter/postgres"
)

// Deps contains process-level resources and external capabilities needed to
// assemble the starmap bounded context.
type Deps struct {
	DB sqlc.DBTX
}

// NewHandler wires the starmap module's adapters, application use cases, and
// gRPC handler.
func NewHandler(deps Deps) *starmapgrpc.Handler {
	queries := sqlc.New(deps.DB)

	starRepo := starmappostgres.NewStarRepository(queries)
	constellationRepo := starmappostgres.NewConstellationRepository(queries)
	traceStasher := starmappostgres.NewTraceStasher(queries)
	traceReader := writingpostgres.NewReader(queries)

	topicGen := starmapapp.NewDefaultTopicGenerator()
	constellationMat := starmapapp.NewDefaultConstellationMatcher()
	assetGen := starmapapp.NewDefaultConstellationAssetGenerator()
	ids := starmapid.NewUUIDGenerator()

	stashTrace := starmapapp.NewStashTraceUseCase(
		traceReader, traceStasher, starRepo, constellationRepo,
		topicGen, constellationMat, assetGen,
		ids,
	)
	listConstellations := starmapapp.NewListConstellationsUseCase(constellationRepo)
	getConstellation := starmapapp.NewGetConstellationUseCase(
		constellationRepo, starRepo, traceReader,
	)

	return starmapgrpc.NewHandler(stashTrace, listConstellations, getConstellation)
}
