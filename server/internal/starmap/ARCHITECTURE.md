# starmap Architecture

Bounded context: Starmap.

Responsibilities:

- Stash Trace as Star.
- Cluster Stars into Constellations.
- Generate constellation-level Insight and assets.
- Serve starmap overview and constellation detail.

## Layer design

| Layer | Role |
| --- | --- |
| `domain/` | Domain types (Star, Constellation) + ports (StarRepository, ConstellationRepository, TraceStasher, TopicGenerator, ConstellationMatcher, ConstellationAssetGenerator) |
| `app/` | Use cases: StashTrace, ListConstellations, GetConstellation + business policies: TopicGenerator, ConstellationMatcher, ConstellationAssetGenerator |
| `adapter/grpc/` | gRPC handler + mapper, delegates to app use cases |
| `adapter/postgres/` | star_repo, constellation_repo, trace_stasher, star_reader |
| `adapter/id/` | UUID generator |
| `module.go` | Module-level composition: creates repos from DB, wires use cases + policies + handler |

## Dependency flow

```
bootstrap/starmap.go → starmap.Deps{DB} → starmap/module.go
  → starmap postgres repos (StarRepo, ConstellationRepo, TraceStasher)
  → writing postgres Reader (for Trace/Moment reads)
  → starmap app policies (TopicGenerator, ConstellationMatcher, ConstellationAssetGenerator)
  → starmap app use cases (StashTrace, ListConstellations, GetConstellation)
  → starmap adapter/grpc handler
```
