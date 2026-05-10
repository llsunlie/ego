# timeline Contract

Timeline exposes read models for frontend browsing.

## Owned writes

- None

## Read-only RPCs

- `ListTraces` — cursor-paginated Trace list
- `GetTraceDetail` — Trace with aggregated Moment + Echo + Insight items
- `GetRandomMoments` — random N historical Moments

## Reads from other modules

- `writing/domain` types: Trace, Moment, Echo, Insight, TraceItem
- `writing/adapter/postgres`: Reader, EchoRepository, InsightRepository

Timeline must not update `stashed` or mutate Moments.
