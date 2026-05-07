# platform Contract

Platform exposes infrastructure capabilities to application wiring.

Current concrete capabilities:

- `auth.GenerateJWT`
- `auth.ParseJWT`
- `auth.UnaryServerInterceptor`
- `postgres.Connect`
- `postgres/sqlc` generated queries

Future capabilities should be exposed through narrow interfaces requested by business modules, for example:

- `EmbeddingProvider`
- `MomentInsightGenerator`
- `PastSelfResponder`
- `EventPublisher`

Business modules should not call external AI SDKs, pgx pool setup, or JWT internals directly from domain code.

