package bootstrap

import (
	identityapp "ego-server/internal/identity/app"
	identitygrpc "ego-server/internal/identity/adapter/grpc"
	identitypostgres "ego-server/internal/identity/adapter/postgres"
	"ego-server/internal/platform/postgres/sqlc"

	pb "ego-server/proto/ego"
)

func NewIdentityHandler(p *Platform) pb.EgoServer {
	queries := sqlc.New(p.Pool)
	userRepo := identitypostgres.NewUserRepository(queries)
	loginUseCase := identityapp.NewLoginUseCase(userRepo, p.Hasher, p.Tokens)
	return identitygrpc.NewHandler(loginUseCase)
}
