# platform Progress

## Current State

- Auth: JWT primitives and gRPC interceptor implemented, tests passing.
- Postgres: Docker Compose configured (pgvector/pgvector:pg16), volume fixed to managed Docker volume. Connection pool (`Connect()`) tested. Migrations (`001_users.sql`, `002_moments.sql`, `003_traces.sql`) applied. sqlc queries tested — 21 tests passing (3 users + 7 moments + 5 traces + 2 connection).
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
