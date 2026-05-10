package timeline

import (
	"ego-server/internal/platform/postgres/sqlc"
	timelinegrpc "ego-server/internal/timeline/adapter/grpc"
	timelineapp "ego-server/internal/timeline/app"
	writingpostgres "ego-server/internal/writing/adapter/postgres"
)

// Deps contains process-level resources and external capabilities needed to
// assemble the timeline bounded context.
type Deps struct {
	DB sqlc.DBTX
}

// NewHandler wires the timeline module's read adapters, application use cases,
// and gRPC handler.
func NewHandler(deps Deps) *timelinegrpc.Handler {
	queries := sqlc.New(deps.DB)

	reader := writingpostgres.NewReader(queries)
	echoRepo := writingpostgres.NewEchoRepository(queries)
	insightRepo := writingpostgres.NewInsightRepository(queries)

	listTraces := timelineapp.NewListTracesUseCase(reader)
	getTraceDetail := timelineapp.NewGetTraceDetailUseCase(reader, echoRepo, insightRepo)
	getRandomMoments := timelineapp.NewGetRandomMomentsUseCase(reader)

	return timelinegrpc.NewHandler(listTraces, getTraceDetail, getRandomMoments)
}
