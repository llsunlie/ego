# writing Progress

## Current State (2026-05-08)

Writing module fully aligned with updated design docs (`docs/app/entity-relationships.md`, `docs/app/api.proto`). All unit tests and smoke tests pass. Build: `go build ./...` succeeds.

### Test summary

| Layer | Tests | Status |
|---|---|---|
| `app/` | 13 | All pass |
| `adapter/grpc/` | 18 (including 3 smoke, 5 handler, 5 mapper) | All pass |
| `adapter/postgres/` | 17 | Require running PostgreSQL |
| `smoke.sh` | 5 (F1, F2, F9, ListTraces, GetTraceDetail) | All pass |

### Completed layers

| Layer | Files | Status |
| --- | --- | --- |
| `domain` | `types.go`, `errors.go`, `ports.go` | Complete |
| `app` | `ports.go`, `create_moment.go`, `generate_insight.go` | Complete |
| `adapter/postgres` | `trace_repo.go`, `moment_repo.go`, `echo_repo.go`, `insight_repo.go`, `reader.go` | Complete |
| `adapter/grpc` | `handler.go`, `mapper.go` | Complete |
| `bootstrap` | `writing.go`, `composite.go`, `server.go`, `cmd/ego/main.go` | Complete |
| `platform/migrations` | `002_moments.sql`, `003_traces.sql`, `004_echos.sql`, `005_insights.sql` | Complete |
| `platform/queries` | `moments.sql`, `traces.sql`, `echos.sql`, `insights.sql` | Complete |

### Key Design Decisions

1. **Trace.Motivation**: 替代旧 Topic 字段。来源：`direct`（直接创建）、`trace:<id>`（延续）、`constellation:<id>`（从星座进入）。
2. **Moment.Embeddings JSONB**: 多模型向量组 `[{model, embedding}]` 替代旧 pgvector 单向量，避免向量扩展依赖。
3. **Echo 持久化**: Echo 从临时值对象变为持久化实体，支持 GetTraceDetail 回溯。初始冷启动时 echo 为 nil。
4. **Insight 持久化**: Writing 生成的会话级 Insight 持久化到 `insights` 表（与 Starmap 星座级 Insight 共享）。
5. **游标分页**: `ListTraces` 和 `ListByUserID` 使用 cursor（最后一条 ID）→ created_at 转换做 SQL 查询。
6. **Composite gRPC handler**: `bootstrap/composite.go` 路由 RPC 到各模块 handler。Writing 负责 CreateMoment、GenerateInsight、ListTraces、GetTraceDetail。

## Known Gaps

- **AI stubs**: `EmbeddingGenerator`、`EchoMatcher`、`InsightGenerator` 在 bootstrap/writing.go 中为 stub 实现。EchoMatcher stub 做全量匹配（相似度固定 0.85），未使用实际 embedding 向量。
- **pgvector 加速**: 当前 JSONB embedding 在应用层做匹配；数据量大后需切换到 pgvector 做 `<=>` 余弦距离查询。

## Next Steps

1. 实现真实 `EmbeddingGenerator`、`EchoMatcher`、`InsightGenerator` 通过 `platform/ai`
2. 实现 `StashTrace`、`GetRandomMoments` RPC（当前返回 Unimplemented）
3. 继续 `timeline` 或 `starmap` 模块实现
