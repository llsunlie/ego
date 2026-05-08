package bootstrap

import (
	"context"
	"fmt"

	conversationapp "ego-server/internal/conversation/app"
	conversationdomain "ego-server/internal/conversation/domain"
	conversationgrpc "ego-server/internal/conversation/adapter/grpc"
	conversationpostgres "ego-server/internal/conversation/adapter/postgres"
	"ego-server/internal/platform/postgres/sqlc"
	starmappostgres "ego-server/internal/starmap/adapter/postgres"
	writingdomain "ego-server/internal/writing/domain"
	writingpostgres "ego-server/internal/writing/adapter/postgres"

	pb "ego-server/proto/ego"
)

// --- AI stub for MVP --------------------------------------------------------

type stubChatGenerator struct{}

func (stubChatGenerator) GenerateOpening(_ context.Context, topic string, moments []writingdomain.Moment) (string, []conversationdomain.MomentReference, error) {
	refs := buildChatRefs(moments)
	return fmt.Sprintf("嗨，我是那时的你。关于「%s」，那时候我写下了这些想法。你想聊些什么？", topic), refs, nil
}

func (stubChatGenerator) GenerateReply(_ context.Context, input conversationdomain.GenerateReplyInput) (*conversationdomain.GenerateReplyOutput, error) {
	refs := buildChatRefs(input.ContextMoments)
	return &conversationdomain.GenerateReplyOutput{
		Content:           "嗯，我明白你的感受。那时候的我也是这样的，有些事说出来就好多了。",
		ReferencedMoments: refs,
	}, nil
}

func buildChatRefs(moments []writingdomain.Moment) []conversationdomain.MomentReference {
	if len(moments) == 0 {
		return nil
	}
	refs := make([]conversationdomain.MomentReference, 0, len(moments))
	for _, m := range moments {
		date := m.CreatedAt.Format("1月2日")
		snippet := []rune(m.Content)
		if len(snippet) > 30 {
			snippet = snippet[:30]
		}
		refs = append(refs, conversationdomain.MomentReference{
			Date:    date,
			Snippet: string(snippet),
		})
	}
	return refs
}

// --- Wiring ---------------------------------------------------------------

func NewChatHandler(p *Platform) pb.EgoServer {
	queries := sqlc.New(p.Pool)

	sessionRepo := conversationpostgres.NewSessionRepository(queries)
	messageRepo := conversationpostgres.NewMessageRepository(queries)
	starReader := starmappostgres.NewStarReader(queries)
	momentReader := writingpostgres.NewChatMomentReader(queries)

	startChat := conversationapp.NewStartChatUseCase(
		sessionRepo, messageRepo, starReader, momentReader,
		stubChatGenerator{}, uuidGenerator{},
	)
	sendMessage := conversationapp.NewSendMessageUseCase(
		sessionRepo, messageRepo, starReader, momentReader,
		stubChatGenerator{}, uuidGenerator{},
	)

	return conversationgrpc.NewHandler(startChat, sendMessage)
}
