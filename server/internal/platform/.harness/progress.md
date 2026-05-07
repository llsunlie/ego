# platform Progress

## Current State

- Auth code has been moved under `internal/platform/auth`.
- PostgreSQL connection, migrations, queries, and sqlc output have been moved under `internal/platform/postgres`.

## Next Best Step

- Keep `go test ./...` passing after import path updates.

