---
name: ego-svr-writing
description: 服务端 writing 领域 context — 创建 Moment、Echo 匹配、AI Insight 生成、Trace 聚合。对应前端 client page: now。
---

# ego-cli-server-writing

writing 有界上下文 — ego 的核心写作引擎。DDD 结构：`server/internal/writing/`

## 所属 gRPC 方法

- `CreateMoment` — 创建 moment + 自动创建/关联 trace + 触发 echo 匹配
- `GenerateInsight` — 基于 moment + echo 生成 AI 洞察
- `GetMoments` — 批量获取 moments（用于匹配展示）

## 模块结构 (`server/internal/writing/`)

```
writing/
├── module.go                       # 依赖注入
├── domain/
│   ├── types.go                    # Trace, Moment, Echo, Insight 领域类型
│   ├── ports.go                    # Repository + Reader 接口
│   └── errors.go                   # 领域错误
├── app/
│   ├── create_moment.go            # CreateMoment 用例
│   ├── echo_matcher.go             # Echo 匹配用例（向量相似度）
│   ├── generate_insight.go         # AI 洞察生成用例
│   └── ports.go                    # IDGenerator 等接口
└── adapter/
    ├── grpc/handler.go             # gRPC Handler
    ├── grpc/mapper.go              # proto ↔ domain 映射
    ├── id/uuid.go                  # UUID 生成
    ├── postgres/                   # PostgreSQL 实现
    │   ├── trace_repo.go           # TraceRepository
    │   ├── moment_repo.go          # MomentRepository
    │   ├── moment_reader.go        # MomentReader（跨模块读接口）
    │   ├── echo_repo.go            # EchoRepository
    │   ├── insight_repo.go         # InsightRepository
    │   └── reader.go               # 组合 Reader（给 timeline/starmap 用）
    └── ai/
        ├── embedder.go             # 文本向量嵌入
        └── insight_generator.go    # LLM 洞察生成
```

## 核心领域模型 (`domain/types.go`)

```go
Trace   { ID, UserID, Motivation, Stashed, CreatedAt, FirstMomentContent }
Moment  { ID, TraceID, UserID, Content, Embedding, CreatedAt }
Echo    { ID, TraceID, MomentID, Content, MatchedMomentIDs, Similarities, CreatedAt }
Insight { ID, MomentID, EchoID, TraceID, Text, CreatedAt }
```

## CreateMoment 用例流程 (`app/create_moment.go`)

1. traceId 为空 → 创建新 Trace（`Motivation: "direct"`）
2. traceId 非空 → 追加到已有 Trace（"顺着再想想"）
3. 创建 Moment（含向量嵌入）
4. 触发 Echo 匹配 → 生成 AI 回声 + 匹配历史相似 moments
5. 高相似度 Echo 触发 Insight 生成

## Echo 召回与匹配

当前 Echo 不再全量扫描历史 Moment。创建 Moment 时会生成 active content embedding，并双写：

- `moments.embeddings` JSONB：领域模型兼容和跨模块读取。
- `moment_embedding_vectors.embedding VECTOR(1024)`：pgvector/HNSW dense topK 召回。
- Elasticsearch `ego_moments`：best-effort sparse 召回索引。

`CreateMoment` 中的候选召回流程：

```text
current moment embedding
  -> pgvector/HNSW dense topK
  -> Elasticsearch sparse topK
  -> RRF 融合
  -> EchoMatcher 规则过滤、同 Trace 去重、排序
```

相关配置：

| 环境变量 | 默认 | 说明 |
|---|---:|---|
| `AI_EMBEDDING_MODEL` | `BAAI/bge-m3` | 当前 embedding 模型 |
| `AI_EMBEDDING_DIM` | `1024` | pgvector 写入维度 |
| `ECHO_RECALL_TOP_K` | `10` | dense topK |
| `ECHO_SPARSE_RECALL_ENABLED` | `true` | 是否启用 ES sparse |
| `ECHO_SPARSE_RECALL_TOP_K` | `10` | sparse topK |
| `ECHO_HYBRID_RRF_K` | `60` | RRF 融合常数 |

`app/echo_matcher.go` 仍负责最终规则排序：计算 raw cosine、加时间距离调整、按 `echo_score >= 0.65` 过滤，并对同一历史 Trace 只保留最高分候选。

## GenerateInsight (`app/generate_insight.go`)

- 基于当前 moment + echo → LLM 生成个性化洞察
- 异步执行，失败静默处理

## 跨模块接口 (`domain/ports.go`)

```go
MomentReader interface { ... }  // 供 timeline/starmap/conversation 使用
TraceReader interface { ... }   // 供 starmap 使用
```

这些接口由 `adapter/postgres/reader.go` 实现，被 timeline、starmap、conversation 模块通过 `module.go` 依赖注入。

## 模块组装 (`module.go`)

```go
type Deps struct {
    DB       sqlc.DBTX
    AIClient *platformai.Client
}
func NewHandler(deps Deps) *writinggrpc.Handler
```

依赖 `platform/ai` 的 `Client`（提供 Embedding + Chat API），以及 `platform/postgres/sqlc` 的数据访问层。

## 相关文件

| 文件 | 说明 |
|------|------|
| `server/internal/platform/ai/client.go` | AI API 客户端（Embedding + Chat） |
| `server/internal/platform/ai/similarity.go` | 余弦相似度计算 |
| `server/internal/platform/postgres/migrations/010_moment_embedding_vectors.sql` | Moment 向量表，`VECTOR(1024)` + HNSW |
| `server/internal/writing/adapter/postgres/echo_candidate_reader.go` | pgvector/HNSW dense topK 候选召回 |
| `server/internal/writing/adapter/elasticsearch/moment_search.go` | ES sparse 索引与候选召回 |
| `server/internal/writing/app/echo_hybrid.go` | dense/sparse RRF 融合 |
| `server/internal/platform/postgres/sqlc/moments.sql.go` | sqlc 生成的 moment 查询 |
| `server/internal/platform/postgres/sqlc/echos.sql.go` | sqlc 生成的 echo 查询 |
| `server/internal/platform/postgres/sqlc/insights.sql.go` | sqlc 生成的 insight 查询 |
| `server/internal/platform/postgres/sqlc/traces.sql.go` | sqlc 生成的 trace 查询 |
| `server/internal/bootstrap/writing.go` | 顶层 wiring |
