# ego Server Agent Guide

This directory is a Go modular monolith organized by DDD bounded contexts.

## Read First

1. `../docs/architecture/backend-structure-decision.md`
2. `../docs/architecture/backend-ddd.md`
3. The target module's `AGENTS.md`, `ARCHITECTURE.md`, and `CONTRACT.md`

## Architecture Rules

- `proto/` is the client-server API contract only.
- Do not use generated proto types as backend domain models.
- `domain/` must not import proto, pgx, sqlc, gRPC status, config, or concrete AI clients.
- `app/` orchestrates use cases and depends on domain interfaces.
- `adapter/grpc/` maps proto requests and responses.
- `adapter/postgres/` implements repositories and read models.
- `platform/` owns infrastructure capabilities such as PostgreSQL, auth, AI clients, gRPC plumbing, and event bus.
- A table has one writing owner. Other modules may only read through explicit contracts.
- Non-owners should not modify another module's implementation. Read its `CONTRACT.md` first.

## Verification

Run from `server/`:

```text
go test ./...
go build ./...
```

