# timeline Architecture

Bounded context: Timeline.

Responsibilities:

- List Traces for the Past page.
- Return Trace detail with aggregated Moment, Echo, and Insight items.
- Return random Moments for memory dots.
- Preserve a read/query-only boundary.

## Layer design

| Layer | Role |
| --- | --- |
| `domain/` | Read-only ports (MomentReader, TraceReader, EchoReader, InsightReader) |
| `app/` | Use cases: ListTraces, GetTraceDetail, GetRandomMoments (default values + assembly) |
| `adapter/grpc/` | gRPC handler + mapper, delegates to app use cases |
| `module.go` | Module-level composition: creates read adapters from DB, wires use cases + handler |

## Dependency flow

```
bootstrap/timeline.go → timeline.Deps{DB} → timeline/module.go
  → writingpostgres repos (Reader, EchoRepo, InsightRepo)
  → timeline/app use cases
  → timeline/adapter/grpc handler
```
