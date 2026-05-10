# starmap Contract

## Owned writes

- `stars`
- `constellations`
- `insights`

## RPCs

- `StashTrace` — Stash a Trace as a Star, match/create Constellation
- `ListConstellations` — List all Constellations for current user with total star count
- `GetConstellation` — Constellation detail with aggregated Moments and Stars

## Read contracts for other modules

- PastSelfCard via ConstellationInsight for Conversation.

## Reads from other modules

- `writing/domain` types: Trace, Moment
- `writing/adapter/postgres`: Reader (for TraceReader implementation)

Starmap must not mutate Traces or Moments except via its own TraceStasher (marks `stashed=true`).
