# timeline Progress

## Current State (2026-05-08)

Timeline module implemented. ListTraces and GetTraceDetail moved from Writing, GetRandomMoments newly implemented. All 3 RPCs routed through composite handler.

### Test summary

| Layer | Tests | Status |
|---|---|---|
| `adapter/grpc/` | 6 (ListTraces, ListTraces_Pagination, GetTraceDetail, GetTraceDetail_NotFound, GetRandomMoments, GetRandomMoments_DefaultCount) | All pass |

### Completed layers

| Layer | Files | Status |
| --- | --- | --- |
| `domain` | `ports.go` | Complete |
| `adapter/grpc` | `handler.go`, `mapper.go` | Complete |
| `bootstrap` | `timeline.go` | Complete |

### RPCs owned by Timeline

| RPC | Description |
| --- | --- |
| `ListTraces` | Cursor-paginated trace list for current user |
| `GetTraceDetail` | Trace detail with Moment + Echo[] + Insight items |
| `GetRandomMoments` | Random N historical moments (memory dot blind box) |

### Key Design Decisions

1. **Read-only module**: Timeline owns no tables. All reads go through Writing's `MomentReader` / `TraceReader` interfaces.
2. **Domain types from writing**: Timeline imports `writing/domain` types (Trace, Moment, Echo, Insight, TraceItem) rather than redefining them.
3. **Independent domain ports**: Timeline defines its own `MomentReader`, `TraceReader`, `EchoReader`, `InsightReader` interfaces — signatures compatible with Writing domain, but timeline doesn't import writing/domain ports.
4. **Mapper duplication**: `momentToProto`, `echoToProto`, `insightToProto`, `traceToProto`, `traceItemToProto` are duplicated in timeline to avoid cross-module adapter dependency.
5. **No postgres adapter**: Timeline reuses `writingpostgres.NewReader` (which implements both MomentReader and TraceReader) from `bootstrap/timeline.go`.

### Reads from other modules

- `writing/domain` types: Trace, Moment, Echo, Insight, TraceItem, EmbeddingEntry
- `writing/adapter/postgres`: Reader (via `NewReader`), EchoRepository (via `NewEchoRepository`), InsightRepository (via `NewInsightRepository`)

## Next Steps

- Implement timeline-specific queries if read patterns diverge from Writing's reader
- Add mapper tests for timeline
