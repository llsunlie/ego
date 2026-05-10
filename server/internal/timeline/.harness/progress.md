# timeline Progress

## Current State (2026-05-10)

Timeline module refactored following the two-level assembly pattern. App use cases extracted from handler, module.go introduced, bootstrap slimmed to DB-only injection.

**6 tests pass** (ListTraces, ListTraces_Pagination, GetTraceDetail, GetTraceDetail_NotFound, GetRandomMoments, GetRandomMoments_DefaultCount).

### Test summary

| Layer | Tests | Status |
|---|---|---|
| `adapter/grpc/` | 6 | All pass |

### Completed layers

| Layer | Files | Status |
| --- | --- | --- |
| `domain` | `ports.go` | Complete |
| `app` | `list_traces.go`, `get_trace_detail.go`, `get_random_moments.go` | Complete |
| `adapter/grpc` | `handler.go`, `mapper.go` | Complete |
| `module wiring` | `module.go` | Complete |
| `bootstrap` | `bootstrap/timeline.go` | Complete |

### RPCs owned by Timeline

| RPC | Description |
| --- | --- |
| `ListTraces` | Cursor-paginated trace list for current user |
| `GetTraceDetail` | Trace detail with Moment + Echo[] + Insight items |
| `GetRandomMoments` | Random N historical moments (memory dot blind box) |

### Key Design Decisions

1. **Read-only module**: Timeline owns no tables. All reads go through Writing's postgres adapter via Reader, EchoRepository, InsightRepository.
2. **Domain types from writing**: Timeline imports `writing/domain` types rather than redefining them.
3. **Own domain ports**: Timeline defines its own `MomentReader`, `TraceReader`, `EchoReader`, `InsightReader` interfaces — compatible with Writing domain, but decoupled.
4. **Two-level assembly**: `bootstrap/timeline.go` passes only DB pool; `timeline/module.go` assembles read adapters, app use cases, and gRPC handler.
5. **App use cases**: Default values (pageSize=20, count=3) and TraceItem assembly logic live in `app/`, not in handler.
6. **Mapper in adapter/grpc**: Proto conversion kept in `mapper.go`.

### Reads from other modules

- `writing/domain` types: Trace, Moment, Echo, Insight, TraceItem
- `writing/adapter/postgres`: Reader, EchoRepository, InsightRepository

## Next Steps

- Add dedicated app layer unit tests
- Wire real Reader/InsightRepo/EchoRepo to pass GetTraceDetail end-to-end data
