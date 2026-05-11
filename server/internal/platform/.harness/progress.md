# platform Progress

## Current State

- Auth: JWT primitives and gRPC interceptor implemented, tests passing.
- Postgres: Docker Compose configured (pgvector/pgvector:pg16), volume fixed to managed Docker volume. Connection pool (`Connect()`) tested. Migrations applied. sqlc queries tested — 21 tests passing.
- Logging: slog+zap structured logging with context propagation. `logging.New()` creates `*slog.Logger` backed by zap via `zapslog.NewHandler()`. `WithLogger`/`FromContext` propagate request-scoped logger through context. gRPC interceptor injects request_id/user_id/method into every request logger. Config-driven via `LOG_LEVEL`/`LOG_FORMAT` env vars. Integrated into `bootstrap.Platform` and all cmd entry points. 19 tests passing (12 logging + 7 auth).
- AI: OpenAI-compatible HTTP client implemented. `Client` wraps the embeddings and chat completions endpoints. Config-driven via `AI_API_KEY`/`AI_BASE_URL`/`AI_EMBEDDING_MODEL`/`AI_CHAT_MODEL` env vars (with `.env` file support). `similarity.go` provides cosine similarity. Wired into `bootstrap.Platform` via `AIClient` field. 8 tests passing (2 client integration + 6 embedding similarity). Embedding quality verified with Qwen/Qwen3-VL-Embedding-8B on SiliconFlow:
  - 相似组 1 (冒名顶替综合征): within-group sim 0.56-0.67 ✓
  - 相似组 2 (摄影与自我映射): within-group sim 0.59-0.72 ✓
  - 关键词碰撞 (技术 vs 人际关系): sim 0.56, 模型未过拟合到表面词汇 ✓
  - 琐事 vs 深度独白: sim 0.57, 模型能区分日常流水账和内心探索 ✓
  - 跨组分离: within-group avg 0.63, cross-group avg 0.49, gap 0.14 ✓
- grpc: placeholder (README only, no code).
- eventbus: placeholder (README only, no code).

## Docker

- `docker compose up -d postgres` from repo root starts the database on port 5432.
- DB: `ego`, user: `ego`, password: `ego`.
- Migration: `server/internal/platform/postgres/migrations/001_users.sql`.

## Next Best Step

- Implement `grpc` server plumbing (error mapping, transport helpers).
- Implement `eventbus` in-memory dispatcher for domain events.
- Implement `adapter/ai/` in writing/conversation/starmap modules to consume `platform/ai.Client`.
