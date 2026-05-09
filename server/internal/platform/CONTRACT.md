# platform Contract

Platform exposes infrastructure capabilities to application wiring.

Current concrete capabilities:

- `logging.New(cfg Config) (*slog.Logger, error)` — creates a structured logger backed by zap
- `logging.NewDefault() *slog.Logger` — dev-friendly defaults (text, debug level, caller info)
- `logging.NewNop() *slog.Logger` — no-op logger for tests
- `logging.WithLogger(ctx, logger)` — inject logger into context
- `logging.FromContext(ctx) *slog.Logger` — extract logger from context, falls back to slog.Default()
- `auth.UnaryServerInterceptor(jwtSecret, baseLogger)` — gRPC interceptor with auth + request-scoped logger injection (request_id, user_id, method)
- `auth.GenerateJWT`
- `auth.ParseJWT`
- `postgres.Connect`
- `postgres/sqlc` generated queries

Future capabilities should be exposed through narrow interfaces requested by business modules, for example:

- `EmbeddingProvider`
- `MomentInsightGenerator`
- `PastSelfResponder`
- `EventPublisher`

Business modules should not call external AI SDKs, pgx pool setup, or JWT internals directly from domain code.

