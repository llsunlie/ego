# platform Architecture

`platform` is the infrastructure layer for the Go server.

It contains concrete technology adapters used by DDD modules through interfaces.

## Current Subpackages

- `auth`: JWT and password-related primitives.
- `postgres`: pgx pool, migrations, sqlc queries.
- `grpc`: server plumbing and gRPC-specific helpers.
- `ai`: AI SDK clients, prompt templates, output validation.
- `eventbus`: in-process event bus and future outbox support.

## Dependency Rule

Business modules may depend on interfaces. `platform` may implement those interfaces. Domain packages must not import concrete platform adapters.

