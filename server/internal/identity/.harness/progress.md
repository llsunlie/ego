# identity Progress

## Current State

Identity module fully refactored to two-level assembly pattern:

- `domain/` — User entity, UserRepository interface, domain errors
- `app/` — LoginUseCase (login/register), PasswordHasher port, TokenIssuer port, IDGenerator port
- `adapter/grpc/` — Thin handler: proto ↔ app mapping, error → gRPC status mapping
- `adapter/postgres/` — UserRepository impl wrapping sqlc.Queries
- `adapter/id/` — UUID generator satisfying app.IDGenerator
- `module.go` — Module composition: adapter/id → app use case → gRPC handler
- `bootstrap/identity.go` — Slim injection: DB + Hasher + Tokens

All 9 tests (4 integration + 5 unit) pass.

## Next Steps

- Add unit tests for LoginUseCase directly (with mocked ports)

## Blockers

None.
