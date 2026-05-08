package bootstrap

import (
	"ego-server/internal/platform/postgres/sqlc"
	timelinegrpc "ego-server/internal/timeline/adapter/grpc"
	writingpostgres "ego-server/internal/writing/adapter/postgres"

	pb "ego-server/proto/ego"
)

func NewTimelineHandler(p *Platform) pb.EgoServer {
	queries := sqlc.New(p.Pool)

	reader := writingpostgres.NewReader(queries)
	echoRepo := writingpostgres.NewEchoRepository(queries)
	insightRepo := writingpostgres.NewInsightRepository(queries)

	return timelinegrpc.NewHandler(reader, reader, echoRepo, insightRepo)
}
