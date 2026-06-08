package setting

import (
	settinggrpc "ego-server/internal/setting/adapter/grpc"
	settingpostgres "ego-server/internal/setting/adapter/postgres"
	settingapp "ego-server/internal/setting/app"
	"ego-server/internal/platform/postgres/sqlc"
)

// Deps contains process-level resources needed by the setting module.
type Deps struct {
	DB sqlc.DBTX
}

// NewHandler wires the setting module's adapters, use cases, and gRPC handler.
func NewHandler(deps Deps) *settinggrpc.Handler {
	queries := sqlc.New(deps.DB)
	userReader := settingpostgres.NewUserReader(queries)
	getProfileUseCase := settingapp.NewGetProfileUseCase(userReader)
	return settinggrpc.NewHandler(getProfileUseCase)
}
