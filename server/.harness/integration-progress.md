# Backend Integration Progress

This file records backend cross-module integration status only.

Module-local progress belongs in each module's `.harness/progress.md`.

## Current State

- P1 Echo recall integration is implemented across Platform and Writing.
- Platform owns pgvector schema/config: `010_moment_embedding_vectors.sql`, `AI_EMBEDDING_DIM`, and `ECHO_RECALL_TOP_K`.
- Writing owns the app-level recall port and Postgres candidate reader. CreateMoment now uses pgvector topK candidates as the primary Echo recall source.
- Historical vector backfill is available via `server/cmd/backfill-moment-vectors`.
- P2.5 Echo sparse hybrid recall is implemented across Platform and Writing.
- Platform owns the ES client/config and local Docker service entry.
- Writing owns the ES Moment search index adapter, search backfill command, sparse topK port, concurrent dense+sparse recall orchestration, and RRF candidate fusion before Echo ranking.
- Echo recall logging now records dense / ES / fused candidate summaries, EchoMatcher score calculations, and final matches, while composite gRPC full req/res logs are suppressed.
- Existing proto/API surface is unchanged.

## Last Verified

- 2026-06-03: `go test ./internal/writing/...` passes.
- 2026-06-03: `go test ./...` still fails in unrelated `internal/conversation/adapter/ai` truncation test.
- 2026-06-04: `go test ./internal/writing/app ./internal/writing/adapter/elasticsearch ./cmd/backfill-moment-search ./internal/bootstrap` passes.
- 2026-06-04: `go test ./...` passes.
- 2026-06-04: `docker compose config` and `docker compose build elasticsearch` pass; IK plugin installs successfully.

## Next Best Step

- Apply pending database migration and run `server/cmd/backfill-moment-vectors` before enabling dense recall against existing data.
- Start Elasticsearch and run `server/cmd/backfill-moment-search` before relying on sparse recall against existing data.
- Continue with P3 Echo score/front-end compatibility design.
