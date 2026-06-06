# starmap Progress

## Current State (2026-06-04)

Starmap module with async clustering and optimistic pending-star UI. `StashTrace` returns immediately (~50ms) with Star(topic="聚合中"), background goroutine handles the AI pipeline (topic → match → assets → constellation). `ListConstellations` surfaces unclustered stars as synthetic single-star constellations so the frontend can render them immediately.

P4 TraceProfile sidecar persistence is implemented. `StashTrace` still returns immediately and the existing async topic clustering path is unchanged. A separate background path generates a TraceProfile from trace moments, retries LLM JSON generation up to two times, falls back to a minimal profile when needed, embeds `profile_text`, and upserts `trace_profiles` plus `trace_profile_vectors`. TraceProfile failures are logged and do not block existing constellation clustering.

P5 TraceProfile quality baseline is established. Fixed review cases now live under `docs/matching-optimization/test-data/trace_profile_cases.json`, and adapter-level regression tests cover prompt construction, JSON parsing, field normalization, fallback behavior, and profile text construction without calling live AI services.

P6 ConstellationProfile target design is documented. The planned model keeps `Trace -> Star` one-to-one, introduces `Star <-> Constellation` many-to-many membership through `constellation_stars`, keeps `TraceProfile` as the content profile name, and adds `ConstellationProfile` as the long-term theme profile.

P7 profile-based constellation matching is implemented as the only `StashTrace` async clustering path. Runtime now generates/persists TraceProfile, recalls ConstellationProfile candidates, scores them, writes primary/secondary memberships through `constellation_stars`, syncs `constellations.star_ids` for proto compatibility, and updates ConstellationProfile stats plus centroid embedding. The old topic-based `clusterAsync` path and TopicGenerator/ConstellationMatcher fallback have been removed.

P7 cleanup adds app-level TraceProfile generation retry around the async clustering path. Critical async failures log `recovery=pending_message_queue`, marking the future consistency mechanism; P8 is now reserved for ConstellationProfile merge quality and eventual-consistency design.

P7.1 over-splitting reduction is implemented. TraceProfile/ConstellationProfile now carry `pattern_tags`, profile text includes them, repositories persist them through JSONB, and migration `013_profile_pattern_tags.sql` backfills the column for existing local tables. Matching now uses `pattern_tags_overlap` instead of `central_pattern_overlap`, lowers deterministic thresholds to strong=0.72 / middle=0.60, avoids duplicate centroid weighting for single-trace constellations, and allows explainable middle matches when score is in [0.58, 0.60) with at least three structured evidence dimensions. Candidate logs now include matched keywords, scenes, emotions, pattern tags, score components, and threshold gaps.

P7.2 follow-up design is documented and remains pending. It is reserved for borderline top3 LLM judgement, more complete secondary memberships, and ConstellationProfile Elasticsearch sparse recall with dense/sparse RRF fusion.

### Test summary

| Layer | Tests | Status |
|---|---|---|
| `app/` | 11 (4 StashTrace + 3 ListConstellations + 4 GetConstellation) | All pass |
| `adapter/grpc/` | 5 (StashTrace, StashTrace_Error, ListConstellations, GetConstellation, GetConstellation_Error) | All pass |
| `internal/starmap/...` | TraceProfile persistence, P7 profile-based constellation matching, P7.1 pattern_tags scoring, multi-membership unique list count, and existing starmap tests | All pass |

### Completed layers

| Layer | Files | Status |
| --- | --- | --- |
| `domain` | `types.go`, `ports.go`, `errors.go` | Complete |
| `app` | `ports.go`, `stash_trace.go`, `list_constellations.go`, `get_constellation.go`, `constellation_asset_generator.go` | Complete |
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
3. **Async insourcing**: `StashTrace.Execute()` creates Star with topic "聚合中" and returns immediately (~50ms). A background goroutine (`clusterWithProfileAsync`) handles TraceProfile generation, ConstellationProfile matching, membership persistence, and asset generation. Each critical failure logs errors via `logging.FromContext`.
4. **Optimistic pending stars (no DB changes)**: No `status` column on `stars`. Unclustered stars are detected at query time (stars not referenced by any constellation's `star_ids`). `ListConstellations` wraps them as synthetic single-star Constellations with user-friendly insight text. `GetConstellation` falls back to the same synthetic response when the ID resolves to a star instead of a constellation. Frontend requires zero changes — `StarFieldPainter._drawLone()` already handles star_count=1.
5. **Business policies in app**: Profile-based clustering orchestration and ConstellationAssetGenerator remain Starmap business logic, with AI and persistence injected through domain ports.
6. **Two-level assembly**: `bootstrap/starmap.go` passes only DB pool; `starmap/module.go` assembles repos, business policies, app use cases, and gRPC handler.
7. **Handler use-case interfaces**: Handler accepts interfaces rather than concrete use-case types, enabling clean mock testing.
8. **Mapper in adapter/grpc**: Proto conversion kept in `mapper.go`.
9. **TraceProfile profile clustering**: TraceProfile is generated asynchronously after `StashTrace`, persisted, and used as the primary material for ConstellationProfile matching.
10. **TraceProfile quality before replacement**: P5 validates generation quality before ConstellationProfile or matching replacement work. Quality samples and generator helper tests are the baseline for prompt tuning.
11. **P6 target membership model**: Future matching should use TraceProfile to compare against ConstellationProfile. A Star can join multiple Constellations as primary/secondary memberships, while `constellations` remains the proto-compatible display entity.
12. **P7 profile matching path**: Current module wiring uses TraceProfile -> ConstellationProfile matching only. `constellation_stars` is the algorithm membership table; `constellations.star_ids` is still synchronized for compatibility.
13. **P7.1 matching refinement**: Pattern tags and revised deterministic scoring are implemented to reduce over-splitting before broader P8 profile merge quality work.
14. **P7.2 marker**: Borderline LLM judgement and constellation sparse recall remain designed follow-up matching iterations.
15. **P8 marker**: ConstellationProfile merge quality and message-queue backed async consistency are reserved for P8 design.

### Known Issues

- **Goroutine failure can leave star partially clustered**: If the async `clusterWithProfileAsync` goroutine exhausts retries or fails during profile/membership/profile-vector persistence, the star can remain with topic="聚合中" or with incomplete profile membership. Critical failures are logged with `star_id`, `trace_id`, and `recovery=pending_message_queue`; durable retry/dead-letter handling is planned for P8.

### Reads from other modules

- `writing/domain` types: Trace, Moment
- `writing/adapter/postgres`: Reader (via `NewReader`) for TraceReader implementation
