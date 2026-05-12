# conversation Architecture

Bounded context: Conversation.

Responsibilities:

- Start or restore ChatSession.
- Append user ChatMessage.
- Generate past-self reply through a port.
- Validate citations and scope.
- Store AI replies and references.

## Layer design

| Layer | Role |
| --- | --- |
| `domain/` | Domain types (ChatSession, ChatMessage) + ports (ChatSessionRepository, ChatMessageRepository, StarReader, MomentReader, ChatGenerator) |
| `app/` | Use cases: StartChat, SendMessage + business policy: DefaultChatGenerator |
| `adapter/grpc/` | gRPC handler + mapper, delegates to app use cases |
| `adapter/postgres/` | session_repo, message_repo |
| `adapter/id/` | UUID generator |
| `module.go` | Module-level composition: creates repos from DB, wires use cases + policies + handler |

## Dependency flow

```
bootstrap/chat.go → conversation.Deps{DB} → conversation/module.go
  → conversation postgres repos (SessionRepo, MessageRepo)
  → starmap postgres StarReader
  → writing postgres ChatMomentReader
  → conversation app policy (DefaultChatGenerator)
  → conversation app use cases (StartChat, SendMessage)
  → conversation adapter/grpc handler
```
