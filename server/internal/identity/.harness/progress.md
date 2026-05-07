# identity Progress

## Current State

Identity module has been fully refactored into proper DDD layers:

- `domain/` — User entity, UserRepository interface, domain errors (ErrUserNotFound, ErrInvalidPassword)
- `app/` — LoginUseCase orchestrating login/register, PasswordHasher port, TokenIssuer port
- `adapter/grpc/` — Thin handler: proto ↔ app mapping, error → gRPC status mapping
- `adapter/postgres/` — UserRepository impl wrapping sqlc.Queries, sqlc ↔ domain mapping
- `platform/auth/` — BcryptHasher + JWTIssuer satisfying app ports via Go structural typing
- `cmd/ego/main.go` — Bootstrap wiring updated

All 5 handler test scenarios pass.

## Next Steps

- Add unit tests for LoginUseCase directly (with mocked ports)
- Add unit tests for UserRepository against a test database

## Blockers

None.
