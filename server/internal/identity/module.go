package identity

import (
	identitygrpc "ego-server/internal/identity/adapter/grpc"
	identityid "ego-server/internal/identity/adapter/id"
	identitypostgres "ego-server/internal/identity/adapter/postgres"
	identityapp "ego-server/internal/identity/app"
	"ego-server/internal/platform/postgres/sqlc"
)

// Deps contains process-level resources and external capabilities needed to
// assemble the identity bounded context.
type Deps struct {
	DB        sqlc.DBTX
	Hasher    identityapp.PasswordHasher
	Tokens    identityapp.TokenIssuer
	SmsSender identityapp.SmsService
}

// NewHandler wires the identity module's adapters, application use cases, and
// gRPC handler.
func NewHandler(deps Deps) *identitygrpc.Handler {
	queries := sqlc.New(deps.DB)

	userRepo := identitypostgres.NewUserRepository(queries)
	ids := identityid.NewUUIDGenerator()

	loginUseCase := identityapp.NewLoginUseCase(userRepo, deps.Hasher, deps.Tokens)
	registerUseCase := identityapp.NewRegisterUseCase(userRepo, deps.Hasher, deps.Tokens, ids, deps.SmsSender)
	sendCodeUseCase := identityapp.NewSendCodeUseCase(deps.SmsSender)
	checkPhoneUseCase := identityapp.NewCheckPhoneUseCase(userRepo)
	resetPasswordUseCase := identityapp.NewResetPasswordUseCase(userRepo, deps.Hasher, deps.Tokens, deps.SmsSender)

	return identitygrpc.NewHandler(loginUseCase, registerUseCase, sendCodeUseCase, checkPhoneUseCase, resetPasswordUseCase)
}
