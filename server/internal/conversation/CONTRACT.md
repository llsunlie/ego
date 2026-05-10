# conversation Contract

## Owned writes

- `chat_sessions`
- `chat_messages`

## RPCs

- `StartChat` — Create or resume a ChatSession, return opening message
- `SendMessage` — Save user message, generate and save past-self reply with citations

## Reads from other modules

- Starmap: Star (via `starmap/adapter/postgres`.StarReader)
- Writing: Moment (via `writing/adapter/postgres`.ChatMomentReader)

Conversation must not generate PastSelfCards or mutate Constellations.
