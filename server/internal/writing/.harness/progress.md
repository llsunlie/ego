# writing Progress

## Current State (2026-05-08)

All 12 features in `feature_list.json` are passing with tests. The writing module has a full DDD skeleton wired into the server binary. Build passes: `go build ./...` succeeds. Tests: 39 passing across app (13), adapter/grpc (9), adapter/postgres (17).

### Completed layers

| Layer | Files | Status |
| --- | --- | --- |
| `domain` | `types.go`, `errors.go`, `ports.go` | Complete |
| `app` | `ports.go`, `create_moment.go`, `generate_insight.go` | Complete |
| `adapter/postgres` | `trace_repo.go`, `moment_repo.go`, `reader.go` | Complete |
| `adapter/grpc` | `handler.go`, `mapper.go` | Complete |
| `bootstrap` | `writing.go`, `composite.go`, `server.go` (updated), `cmd/ego/main.go` (updated) | Complete |
| `platform/migrations` | `002_moments.sql`, `003_traces.sql` | Complete |
| `platform/queries` | `moments.sql`, `traces.sql` | Complete |

### Key Design Decisions

1. **Trace + Moment as separate aggregates**: Trace owns the session lifecycle; Moment is independently queryable for Echo matching and cross-module reads.
2. **Insight not persisted in Writing**: Writing generates "current-session" Insight on-the-fly and returns it. Constellation-level Insight persists in Starmap's `insights` table.
3. **Embedding stored on Moment**: Generated once at creation time and persisted for subsequent Echo matching queries.
4. **Moment.Connected managed externally**: Writing defaults it to `false`; Starmap updates it via domain event after `StashTrace`.
5. **Cursor pagination via timestamp**: ListByUserID uses cursor (last item ID) → converts to created_at for SQL query, avoiding ambiguous subquery.
6. **Composite gRPC handler**: `bootstrap/composite.go` routes each RPC to the owning module handler. Only Login (identity), CreateMoment (writing), and GenerateInsight (writing) are wired; other RPCs return Unimplemented.

## Known Gaps

- **AI stubs**: `EmbeddingGenerator`, `EchoMatcher`, `InsightGenerator` are stubbed in bootstrap/writing.go. They need real implementations via `platform/ai` module.
- **Starmap interaction**: `Moment.Connected` update path from Starmap's `StashTrace` is not yet defined (needs domain event wiring via `platform/eventbus`).

## Next Steps

1. Implement real `EmbeddingGenerator`, `EchoMatcher`, `InsightGenerator` via `platform/ai`
2. Wire domain events for cross-module communication (Moment.Connected update)
3. Proceed with `timeline` or `starmap` module implementation
