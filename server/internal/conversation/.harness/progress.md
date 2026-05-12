# conversation Progress

## Current State (2026-05-10)

Conversation module refactored following the two-level assembly pattern. DefaultChatGenerator moved from bootstrap to app layer. module.go introduced. All 14 tests pass.

### Test summary

| Layer | Tests | Status |
|---|---|---|
| `app/` | 10 (4 StartChat + 6 SendMessage) | All pass |
| `adapter/grpc/` | 4 (StartChat, StartChat_Error, SendMessage, SendMessage_Error) | All pass |

### Completed layers

| Layer | Files | Status |
| --- | --- | --- |
| `domain` | `types.go`, `ports.go`, `errors.go` | Complete |
| `app` | `ports.go`, `start_chat.go`, `send_message.go`, `chat_generator.go` | Complete |
| `adapter/postgres` | `session_repo.go`, `message_repo.go` | Complete |
| `adapter/grpc` | `handler.go`, `mapper.go` | Complete |
| `adapter/id` | `uuid.go` | Complete |
| `module wiring` | `module.go` | Complete |
| `bootstrap` | `bootstrap/chat.go` | Complete |
| `platform/migrations` | `007_chat.sql` | Complete |
| `platform/queries` | `chat_sessions.sql`, `chat_messages.sql` | Complete |

### RPCs owned by Conversation

| RPC | Description |
| --- | --- |
| `StartChat` | Create or resume ChatSession, return opening + history |
| `SendMessage` | Save user message, generate past-self reply with citations |

### Key Design Decisions

1. **Two-level assembly**: `bootstrap/chat.go` passes only DB pool; `conversation/module.go` assembles repos, business policies, app use cases, and gRPC handler.
2. **ChatGenerator in app**: DefaultChatGenerator (opening + reply generation policies) is a Conversation business strategy, located in `app/`.
3. **Handler use-case interfaces**: Handler accepts interfaces rather than concrete use-case types, enabling clean mock testing.
4. **Mapper in adapter/grpc**: Proto conversion kept in `mapper.go`.
5. **Cross-module reads**: Star data via `starmap/adapter/postgres`.StarReader, Moment data via `writing/adapter/postgres`.ChatMomentReader.

### Reads from other modules

- `starmap/adapter/postgres`: StarReader
- `writing/adapter/postgres`: ChatMomentReader

## Next Steps

1. Implement real AI ChatGenerator via `platform/ai`
2. Add composite handler: StartChat, SendMessage
3. Add postgres adapter integration tests for session_repo and message_repo
