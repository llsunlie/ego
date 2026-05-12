package conversation

import (
	conversationgrpc "ego-server/internal/conversation/adapter/grpc"
	conversationid "ego-server/internal/conversation/adapter/id"
	conversationpostgres "ego-server/internal/conversation/adapter/postgres"
	conversationapp "ego-server/internal/conversation/app"
	"ego-server/internal/platform/postgres/sqlc"
	starmappostgres "ego-server/internal/starmap/adapter/postgres"
	writingpostgres "ego-server/internal/writing/adapter/postgres"
)

// Deps contains process-level resources and external capabilities needed to
// assemble the conversation bounded context.
type Deps struct {
	DB sqlc.DBTX
}

// NewHandler wires the conversation module's adapters, application use cases,
// and gRPC handler.
func NewHandler(deps Deps) *conversationgrpc.Handler {
	queries := sqlc.New(deps.DB)

	sessionRepo := conversationpostgres.NewSessionRepository(queries)
	messageRepo := conversationpostgres.NewMessageRepository(queries)
	starReader := starmappostgres.NewStarReader(queries)
	momentReader := writingpostgres.NewChatMomentReader(queries)

	ids := conversationid.NewUUIDGenerator()
	chatGen := conversationapp.NewDefaultChatGenerator()

	startChat := conversationapp.NewStartChatUseCase(
		sessionRepo, messageRepo, starReader, momentReader,
		chatGen, ids,
	)
	sendMessage := conversationapp.NewSendMessageUseCase(
		sessionRepo, messageRepo, starReader, momentReader,
		chatGen, ids,
	)

	return conversationgrpc.NewHandler(startChat, sendMessage)
}
