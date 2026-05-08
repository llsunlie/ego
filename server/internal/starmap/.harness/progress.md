# starmap Progress

## Current State (2026-05-08)

Starmap module implemented with full DDD architecture. StashTrace, ListConstellations, and GetConstellation RPCs routed through composite handler. All unit tests pass. Build: `go build ./...` succeeds.

### Test summary

| Layer | Tests | Status |
|---|---|---|
| `app/` | 9 (4 StashTrace + 2 ListConstellations + 3 GetConstellation) | All pass |
| `adapter/grpc/` | 5 (StashTrace, StashTrace_Error, ListConstellations, GetConstellation, GetConstellation_Error) | All pass |
| `smoke.sh` | F3 StashTrace → Constellation | Pending smoke run |

### Completed layers

| Layer | Files | Status |
| --- | --- | --- |
| `domain` | `types.go`, `ports.go`, `errors.go` | Complete |
| `app` | `ports.go`, `stash_trace.go`, `list_constellations.go`, `get_constellation.go` | Complete |
| `adapter/postgres` | `star_repo.go`, `constellation_repo.go`, `trace_stasher.go` | Complete |
| `adapter/grpc` | `handler.go`, `mapper.go` | Complete |
| `bootstrap` | `starmap.go` | Complete |
| `platform/migrations` | `006_starmap.sql` | Complete |
| `platform/queries` | `stars.sql`, `constellations.sql` | Complete |

### RPCs owned by Starmap

| RPC | Description |
| --- | --- |
| `StashTrace` | Stash a Trace as a Star, match/create Constellation |
| `ListConstellations` | List all Constellations for current user |
| `GetConstellation` | Constellation detail with Moments + Stars |

### Key Design Decisions

1. **Star per Trace**: Each stashed Trace becomes exactly one Star (`idx_stars_trace` unique index).
2. **Constellation clustering**: `star_ids UUID[]` stored in constellation table (no join table).
3. **Topic/prompts embedding**: `topic_prompts TEXT[]` embedded in constellations table.
4. **Synchronous insourcing**: Constellation matching and asset generation happen synchronously during StashTrace (not async).
5. **AI Stubs**: TopicGenerator (first 20 chars), ConstellationMatcher (always no match = lone-star), AssetGenerator (placeholder names/insights/prompts).
6. **TraceStasher via sqlc**: Marks trace.stashed=true directly using sqlc.Queries (no dependency on writing adapter).
7. **Handler use-case interfaces**: Handler accepts interfaces (`StashTraceUseCase`, etc.) rather than concrete use-case types, enabling clean mock testing.
8. **Mapper duplication**: `momentToProto`, `starToProto`, `constellationToProto` in starmap adapter to avoid cross-module dependency.

### Reads from other modules

- `writing/domain` types: Trace, Moment
- `writing/adapter/postgres`: Reader (via `NewReader`) for TraceReader interface
- `writing/adapter/postgres`: UpdateTrace query reused for TraceStasher

## Next Steps

1. Implement real AI services (TopicGenerator, ConstellationMatcher, AssetGenerator) via `platform/ai`
2. Consider async constellation insourcing for large traces (callback pattern)
3. Implement StartChat + SendMessage RPCs (Chat module)
4. Add postgres adapter integration tests for star_repo and constellation_repo
