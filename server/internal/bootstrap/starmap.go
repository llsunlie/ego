package bootstrap

import (
	"context"
	"math/rand/v2"

	"ego-server/internal/platform/postgres/sqlc"
	starmapapp "ego-server/internal/starmap/app"
	starmapdomain "ego-server/internal/starmap/domain"
	starmapgrpc "ego-server/internal/starmap/adapter/grpc"
	starmappostgres "ego-server/internal/starmap/adapter/postgres"
	writingdomain "ego-server/internal/writing/domain"
	writingpostgres "ego-server/internal/writing/adapter/postgres"

	"github.com/google/uuid"

	pb "ego-server/proto/ego"
)

// --- AI stubs for MVP -----------------------------------------------------

type stubTopicGenerator struct{}

func (stubTopicGenerator) Generate(_ context.Context, moments []writingdomain.Moment) (string, error) {
	if len(moments) > 0 {
		content := []rune(moments[0].Content)
		if len(content) > 20 {
			content = content[:20]
		}
		return "关于" + string(content) + "…", nil
	}
	return "未命名的星", nil
}

type simConstellationMatcher struct{}

func (simConstellationMatcher) FindMatch(_ context.Context, _ string, existing []starmapdomain.Constellation) (string, error) {
	if len(existing) == 0 {
		return "", nil
	}
	// 65% chance to cluster with an existing constellation (simulates AI matching)
	if rand.IntN(100) < 65 {
		return existing[rand.IntN(len(existing))].ID, nil
	}
	return "", nil
}

type stubConstellationAssetGenerator struct{}

func (stubConstellationAssetGenerator) Generate(_ context.Context, _ []writingdomain.Moment) (string, string, []string, error) {
	return "星座" + uuid.New().String()[:8],
		"这些话语之间似乎有着某种共鸣。随着你写下更多，它们之间的联系会变得越来越清晰。",
		[]string{"关于这个主题，还有什么想说的吗？", "换个角度再看一看？"},
		nil
}

// --- Wiring ---------------------------------------------------------------

func NewStarmapHandler(p *Platform) pb.EgoServer {
	queries := sqlc.New(p.Pool)

	starRepo := starmappostgres.NewStarRepository(queries)
	constellationRepo := starmappostgres.NewConstellationRepository(queries)
	traceStasher := starmappostgres.NewTraceStasher(queries)
	traceReader := writingpostgres.NewReader(queries)

	stashTrace := starmapapp.NewStashTraceUseCase(
		traceReader, traceStasher, starRepo, constellationRepo,
		stubTopicGenerator{}, simConstellationMatcher{}, stubConstellationAssetGenerator{},
		uuidGenerator{},
	)
	listConstellations := starmapapp.NewListConstellationsUseCase(constellationRepo)
	getConstellation := starmapapp.NewGetConstellationUseCase(
		constellationRepo, starRepo, traceReader,
	)

	return starmapgrpc.NewHandler(stashTrace, listConstellations, getConstellation)
}
