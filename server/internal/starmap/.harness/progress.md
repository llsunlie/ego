# starmap Progress

## Current State (2026-06-04)

Starmap module with async clustering and optimistic pending-star UI. `StashTrace` returns immediately (~50ms) with Star(topic="聚合中"), background goroutine handles the AI pipeline (topic → match → assets → constellation). `ListConstellations` surfaces unclustered stars as synthetic single-star constellations so the frontend can render them immediately.

P4 TraceProfile sidecar persistence is implemented. `StashTrace` still returns immediately and the existing async topic clustering path is unchanged. A separate background path generates a TraceProfile from trace moments, retries LLM JSON generation up to two times, falls back to a minimal profile when needed, embeds `profile_text`, and upserts `trace_profiles` plus `trace_profile_vectors`. TraceProfile failures are logged and do not block existing constellation clustering.

P5 TraceProfile quality baseline is established. Fixed review cases now live under `docs/matching-optimization/test-data/trace_profile_cases.json`, and adapter-level regression tests cover prompt construction, JSON parsing, field normalization, fallback behavior, and profile text construction without calling live AI services.

P6 ConstellationProfile target design is documented. The planned model keeps `Trace -> Star` one-to-one, introduces `Star <-> Constellation` many-to-many membership through a future `constellation_stars` relation, keeps `TraceProfile` as the content profile name, and adds `ConstellationProfile` as the long-term theme profile. This is design-only; current runtime still uses topic-based async clustering.

### Test summary

| Layer | Tests | Status |
|---|---|---|
| `app/` | 11 (4 StashTrace + 3 ListConstellations + 4 GetConstellation) | All pass |
| `adapter/grpc/` | 5 (StashTrace, StashTrace_Error, ListConstellations, GetConstellation, GetConstellation_Error) | All pass |
| `internal/starmap/...` | TraceProfile sidecar persistence, TraceProfile quality helper coverage, and existing starmap tests | All pass |

### Completed layers

| Layer | Files | Status |
| --- | --- | --- |
| `domain` | `types.go`, `ports.go`, `errors.go` | Complete |
| `app` | `ports.go`, `stash_trace.go`, `list_constellations.go`, `get_constellation.go`, `topic_generator.go`, `constellation_matcher.go`, `constellation_asset_generator.go` | Complete |
| `adapter/postgres` | `star_repo.go`, `constellation_repo.go`, `trace_stasher.go`, `star_reader.go` | Complete |
| `adapter/postgres` | `trace_profile_repo.go` | Complete |
| `adapter/ai` | `trace_profile_generator.go` | Complete |
| `adapter/grpc` | `handler.go`, `mapper.go` | Complete |
| `adapter/id` | `uuid.go` | Complete |
| `module wiring` | `module.go` | Complete |
| `bootstrap` | `bootstrap/starmap.go` | Complete |
| `platform/migrations` | `006_starmap.sql`, `011_trace_profiles.sql` | Complete |
| `platform/queries` | `stars.sql`, `constellations.sql` | Complete |

### RPCs owned by Starmap

| RPC | Description |
| --- | --- |
| `StashTrace` | Stash a Trace as a Star(topic="聚合中"), return immediately, cluster async |
| `ListConstellations` | List all Constellations + unclustered stars as synthetic single-star Constellations |
| `GetConstellation` | Constellation detail; falls back to friendly synthetic view for unclustered star IDs |

### Key Design Decisions

1. **Star per Trace**: Each stashed Trace becomes exactly one Star (`idx_stars_trace` unique index).
2. **Constellation clustering**: `star_ids UUID[]` stored in constellation table.
3. **Async insourcing**: `StashTrace.Execute()` creates Star with topic "聚合中" and returns immediately (~50ms). A background goroutine (`clusterAsync`) handles topic generation, constellation matching, and asset generation. Each step logs errors via `logging.FromContext`.
4. **Optimistic pending stars (no DB changes)**: No `status` column on `stars`. Unclustered stars are detected at query time (stars not referenced by any constellation's `star_ids`). `ListConstellations` wraps them as synthetic single-star Constellations with user-friendly insight text. `GetConstellation` falls back to the same synthetic response when the ID resolves to a star instead of a constellation. Frontend requires zero changes — `StarFieldPainter._drawLone()` already handles star_count=1.
5. **Business policies in app**: TopicGenerator, ConstellationMatcher, ConstellationAssetGenerator MVP defaults are Starmap business policies, located in `app/` per two-level assembly rules.
6. **Two-level assembly**: `bootstrap/starmap.go` passes only DB pool; `starmap/module.go` assembles repos, business policies, app use cases, and gRPC handler.
7. **Handler use-case interfaces**: Handler accepts interfaces rather than concrete use-case types, enabling clean mock testing.
8. **Mapper in adapter/grpc**: Proto conversion kept in `mapper.go`.
9. **TraceProfile as sidecar**: TraceProfile is generated asynchronously after `StashTrace` and persisted for future aggregation work. It does not replace the current topic-based constellation clustering in P4.
10. **TraceProfile quality before replacement**: P5 validates generation quality before ConstellationProfile or matching replacement work. Quality samples and generator helper tests are the baseline for prompt tuning.
11. **P6 target membership model**: Future matching should use TraceProfile to compare against ConstellationProfile. A Star can join multiple Constellations as primary/secondary memberships, while `constellations` remains the proto-compatible display entity.

### Known Issues

- **Goroutine failure leaves star permanently unclustered**: If the async `clusterAsync` goroutine fails (AI call error, DB error, etc.), the star remains with topic="聚合中" forever. It will continue to appear as a synthetic single-star constellation in `ListConstellations` and `GetConstellation` responses. There is no retry mechanism or dead-letter queue. Recovery requires manual DB intervention or a future background reconciliation job. All failures are logged via `logger.ErrorContext` with `star_id` for diagnosis.

### Reads from other modules

- `writing/domain` types: Trace, Moment
- `writing/adapter/postgres`: Reader (via `NewReader`) for TraceReader implementation
