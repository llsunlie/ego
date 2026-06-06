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
- P4 TraceProfile sidecar persistence is implemented across Platform and Starmap.
- Platform owns `011_trace_profiles.sql` for structured TraceProfile storage and pgvector profile embeddings.
- Starmap owns TraceProfile generation, retry/fallback behavior, async sidecar orchestration after `StashTrace`, and Postgres upsert persistence.
- Current constellation topic clustering remains unchanged; TraceProfile does not replace matching in P4.
- P5 TraceProfile quality baseline is established with fixed review samples and generator helper regression tests; it still does not change proto/API or constellation matching behavior.
- P6 ConstellationProfile target design is documented. The model preserves proto-compatible `constellations`, uses `TraceProfile -> ConstellationProfile` matching, and supports many-to-many Star memberships through `constellation_stars`.
- P7 TraceProfile -> ConstellationProfile matching is implemented across Platform and Starmap. Platform owns `012_constellation_profiles.sql`; Starmap owns ConstellationProfile persistence, profile candidate recall/scoring, primary/secondary membership writes, and proto-compatible `constellations.star_ids` synchronization.
- P7 cleanup removes the old topic-based clustering fallback. TraceProfile generation now has app-level retry; critical async failures log `recovery=pending_message_queue` until P8 introduces a durable queue or compensation mechanism.
- Existing proto/API surface is unchanged.

## Last Verified

- 2026-06-03: `go test ./internal/writing/...` passes.
- 2026-06-03: `go test ./...` still fails in unrelated `internal/conversation/adapter/ai` truncation test.
- 2026-06-04: `go test ./internal/writing/app ./internal/writing/adapter/elasticsearch ./cmd/backfill-moment-search ./internal/bootstrap` passes.
- 2026-06-04: `go test ./...` passes.
- 2026-06-04: `docker compose config` and `docker compose build elasticsearch` pass; IK plugin installs successfully.
- 2026-06-04: `go test ./internal/starmap/...` passes with TraceProfile sidecar persistence.
- 2026-06-04: `go test ./internal/starmap/...` passes with TraceProfile quality helper coverage.
- 2026-06-05: P6 ConstellationProfile and multi-membership design documented; no code path changed.
- 2026-06-05: `go test ./internal/starmap/...` passes with P7 profile-based constellation matching.

## Next Best Step

- Apply pending database migration and run `server/cmd/backfill-moment-vectors` before enabling dense recall against existing data.
- Start Elasticsearch and run `server/cmd/backfill-moment-search` before relying on sparse recall against existing data.
- Apply pending `011_trace_profiles.sql` migration before relying on TraceProfile persistence.
- Apply pending `012_constellation_profiles.sql` migration before relying on profile-based constellation matching.
- Continue with P8 ConstellationProfile merge quality and async consistency design.
