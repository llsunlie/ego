# writing Progress

## Current State (2026-06-04)

P1 Echo recall efficiency optimization implemented. Writing now keeps `moments.embeddings` JSONB for long-term compatibility while dual-writing the active content embedding into `moment_embedding_vectors` for pgvector/HNSW topK recall. `CreateMomentUseCase` uses the `EchoCandidateReader` port as the primary candidate source; pgvector query errors surface directly and do not fall back to full user-history scans.

Historical vector backfill is available through `server/cmd/backfill-moment-vectors`, using the active `AI_EMBEDDING_MODEL` and `AI_EMBEDDING_DIM` configuration. `ECHO_RECALL_TOP_K` is configurable and defaults to 10.

Core P1 paths now include structured debug logs for vector dual-write, pgvector topK candidate recall, and historical vector backfill. Logs avoid raw content and embedding payloads, recording only ids, model, dimension, limits, counts, and elapsed time.

P2 Echo ranking rules implemented. `DefaultEchoMatcher` now filters same-trace candidates, applies light time-distance scoring, deduplicates by historical trace, limits results to 3 matches, and preserves raw cosine similarity in persisted Echo similarities.

P2.5 Echo sparse hybrid recall implemented. Writing now best-effort indexes new Moments into Elasticsearch, can backfill historical Moments through `server/cmd/backfill-moment-search`, queries sparse topK Moment IDs from the `ego_moments` index, loads those Moments from PostgreSQL, and merges dense pgvector rank with sparse ES rank through RRF before applying P2 Echo ranking rules. Dense pgvector and sparse ES recall now run concurrently inside `CreateMoment`; pgvector failure remains fatal, while sparse ES failure logs warn and continues with dense-only candidates.

Echo recall logging now focuses on algorithm inspection: `CreateMoment: echo recall candidates` records dense, ES, and fused candidate summaries, `echo match candidate scores` records per-candidate score calculation details, and `CreateMoment: echo final matches` records final matched Moments with similarity. Composite gRPC request/response body logs and low-value ES success step logs are intentionally suppressed.

### Test summary

| Layer | Tests | Status |
|---|---|---|
| `app/` | CreateMoment candidate-reader tests, EchoMatcher P2 ranking tests, and RRF hybrid recall tests | All pass |
| `adapter/elasticsearch/` | ES index creation, sparse query request, and bulk error handling tests | All pass |
| `cmd/backfill-moment-search` | compile coverage for ES search backfill command | All pass |
| `adapter/postgres/` | vector literal validation | All pass |
| `internal/writing/...` | module package tests | All pass |
| `server ./...` | global package tests | All pass |

## Previous State (2026-05-10)

Writing module fully aligned with updated design docs (`docs/app/entity-relationships.md`, `docs/app/api.proto`). All unit tests and smoke tests pass. Build: `go build ./...` succeeds.

**ListTraces**, **GetTraceDetail**, and **GetRandomMoments** RPCs moved to Timeline module (2026-05-08). Writing now owns CreateMoment, GenerateInsight, and the narrow GetMoments read endpoint used by cross-module aggregation.

**Module-level wiring** introduced (2026-05-10): `internal/writing/module.go` now assembles Writing's own repositories, application use cases, internal business policies, and gRPC handler. `bootstrap/writing.go` only passes process-level resources and external infrastructure capabilities into the module.

### Test summary

| Layer | Tests | Status |
|---|---|---|
| `app/` | 13 | All pass |
| `adapter/grpc/` | 11 (including 3 smoke, 4 handler, 4 mapper) | All pass |
| `adapter/postgres/` | 17 | Require running PostgreSQL |
| `smoke.sh` | 5 (F1, F2, ListTraces, GetTraceDetail, GetRandomMoments) | All pass |

### Completed layers

| Layer | Files | Status |
| --- | --- | --- |
| `domain` | `types.go`, `errors.go`, `ports.go` | Complete |
| `app` | `ports.go`, `create_moment.go`, `generate_insight.go` | Complete |
| `adapter/postgres` | `trace_repo.go`, `moment_repo.go`, `echo_repo.go`, `insight_repo.go`, `reader.go` | Complete |
| `adapter/grpc` | `handler.go`, `mapper.go` | Complete |
| `module wiring` | `module.go`, `app/echo_matcher.go`, `app/insight_generator.go`, `adapter/id/uuid.go` | Complete |
| `bootstrap` | `bootstrap/writing.go`, `bootstrap/composite.go`, `bootstrap/server.go`, `cmd/ego/main.go` | Complete |
| `platform/migrations` | `002_moments.sql`, `003_traces.sql`, `004_echos.sql`, `005_insights.sql` | Complete |
| `platform/queries` | `moments.sql`, `traces.sql`, `echos.sql`, `insights.sql` | Complete |

### Key Design Decisions

1. **Trace.Motivation**: 替代旧 Topic 字段。来源：`direct`（直接创建）、`trace:<id>`（延续）、`constellation:<id>`（从星座进入）。
2. **Moment.Embeddings JSONB**: 多模型向量组 `[{model, embedding}]` 替代旧 pgvector 单向量，避免向量扩展依赖。
3. **Echo 持久化**: Echo 从临时值对象变为持久化实体，支持 GetTraceDetail 回溯。初始冷启动时 echo 为 nil。
4. **Insight 持久化**: Writing 生成的会话级 Insight 持久化到 `insights` 表（与 Starmap 星座级 Insight 共享）。
5. **只读 reader 契约**: Timeline 等模块通过 Writing 暴露的 reader/contract 读取 Trace、Moment、Echo、Insight；分页查询能力保留在 reader 实现中，但对外 RPC 归 Timeline。
6. **Composite gRPC handler**: `bootstrap/composite.go` 路由 RPC 到各模块 handler。Writing 负责 CreateMoment、GenerateInsight。
7. **两级装配**: `bootstrap` 是进程级 composition root，只创建和注入 DB、外部 AI 基础能力等进程级资源；`writing/module.go` 是模块级 composition function，只装配 Writing 自己的 adapter、app use case、handler。
8. **业务策略归属 app**: Echo 匹配默认策略和 Insight 默认生成策略属于 Writing 应用业务逻辑，放在 `app/echo_matcher.go`、`app/insight_generator.go`，不放在 bootstrap，也不伪装成 platform adapter。
9. **基础设施能力外部注入**: `EmbeddingGenerator` 是基础 AI 向量化能力，由 bootstrap 注入。Writing 只依赖 `domain.EmbeddingGenerator` port，不直接创建 `platform/ai` 对象、不读取配置、不初始化外部 SDK。

## Known Gaps

- **AI stubs**: `EmbeddingGenerator` 仍由 `bootstrap/writing.go` 提供 stub；EchoMatcher 和 InsightGenerator 目前是 Writing app 内的 MVP 默认业务策略。EchoMatcher 仍做全量匹配（相似度固定 0.85），未使用实际 embedding 向量。
- **pgvector 加速**: 当前 JSONB embedding 在应用层做匹配；数据量大后需切换到 pgvector 做 `<=>` 余弦距离查询。

## Next Steps

1. 在 `platform/ai` 实现真实 `EmbeddingGenerator`，由 `bootstrap` 注入给 `writing.NewHandler`
2. 将 Echo 匹配从 MVP 全量匹配升级为基于 embedding/vector search 的 Writing app 策略
3. 将 Insight 默认生成策略升级为通过 app port 调用 platform AI 基础能力，但业务提示词、输出约束和持久化规则仍归 Writing app
4. 将两级装配方案推广到 Identity、Starmap、Timeline、Conversation
