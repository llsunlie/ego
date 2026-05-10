# starmap Progress

## Current State (2026-05-10)

Starmap module refactored following the two-level assembly pattern. Business policies (TopicGenerator, ConstellationMatcher, ConstellationAssetGenerator) moved from bootstrap to app layer. module.go introduced. All 14 tests pass.

### Test summary

| Layer | Tests | Status |
|---|---|---|
| `app/` | 9 (4 StashTrace + 2 ListConstellations + 3 GetConstellation) | All pass |
| `adapter/grpc/` | 5 (StashTrace, StashTrace_Error, ListConstellations, GetConstellation, GetConstellation_Error) | All pass |

### Completed layers

| Layer | Files | Status |
| --- | --- | --- |
| `domain` | `types.go`, `ports.go`, `errors.go` | Complete |
| `app` | `ports.go`, `stash_trace.go`, `list_constellations.go`, `get_constellation.go`, `topic_generator.go`, `constellation_matcher.go`, `constellation_asset_generator.go` | Complete |
| `adapter/postgres` | `star_repo.go`, `constellation_repo.go`, `trace_stasher.go`, `star_reader.go` | Complete |
| `adapter/grpc` | `handler.go`, `mapper.go` | Complete |
| `adapter/id` | `uuid.go` | Complete |
| `module wiring` | `module.go` | Complete |
| `bootstrap` | `bootstrap/starmap.go` | Complete |
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
2. **Constellation clustering**: `star_ids UUID[]` stored in constellation table.
3. **Synchronous insourcing**: Constellation matching and asset generation happen synchronously during StashTrace.
4. **Business policies in app**: TopicGenerator, ConstellationMatcher, ConstellationAssetGenerator MVP defaults are Starmap business policies, located in `app/` per two-level assembly rules.
5. **Two-level assembly**: `bootstrap/starmap.go` passes only DB pool; `starmap/module.go` assembles repos, business policies, app use cases, and gRPC handler.
6. **Handler use-case interfaces**: Handler accepts interfaces rather than concrete use-case types, enabling clean mock testing.
7. **Mapper in adapter/grpc**: Proto conversion kept in `mapper.go`.

### Reads from other modules

- `writing/domain` types: Trace, Moment
- `writing/adapter/postgres`: Reader (via `NewReader`) for TraceReader implementation

## Next Steps

1. Implement real AI services (TopicGenerator, ConstellationMatcher, AssetGenerator) via `platform/ai`
2. Consider async constellation insourcing for large traces
3. Add postgres adapter integration tests for star_repo and constellation_repo
