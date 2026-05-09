# platform Progress

## Current State

- Auth: JWT primitives and gRPC interceptor implemented, tests passing.
- Postgres: Docker Compose configured (pgvector/pgvector:pg16), volume fixed to managed Docker volume. Connection pool (`Connect()`) tested. Migrations applied. sqlc queries tested — 21 tests passing.
- Logging: slog+zap structured logging with context propagation. `logging.New()` creates `*slog.Logger` backed by zap via `zapslog.NewHandler()`. `WithLogger`/`FromContext` propagate request-scoped logger through context. gRPC interceptor injects request_id/user_id/method into every request logger. Config-driven via `LOG_LEVEL`/`LOG_FORMAT` env vars. Integrated into `bootstrap.Platform` and all cmd entry points. 19 tests passing (12 logging + 7 auth).
- grpc: placeholder (README only, no code).
- ai: placeholder (README only, no code).
- eventbus: placeholder (README only, no code).

## Docker

- `docker compose up -d postgres` from repo root starts the database on port 5432.
- DB: `ego`, user: `ego`, password: `ego`.
- Migration: `server/internal/platform/postgres/migrations/001_users.sql`.

## Next Best Step

- Implement `grpc` server plumbing (error mapping, transport helpers).
- Implement `eventbus` in-memory dispatcher for domain events.
