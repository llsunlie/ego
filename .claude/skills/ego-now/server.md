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
│   ├── echo_hybrid.go              # Dense/Sparse 候选 RRF 融合
│   ├── echo_matcher.go             # Echo 匹配用例（向量相似度阈值过滤）
│   ├── generate_insight.go         # AI 洞察生成用例
│   └── ports.go                    # IDGenerator 等接口
└── adapter/
    ├── grpc/handler.go             # gRPC Handler
    ├── grpc/mapper.go              # proto ↔ domain 映射
    ├── id/uuid.go                  # UUID 生成
    ├── postgres/                   # PostgreSQL 实现
    │   ├── trace_repo.go           # TraceRepository
    │   ├── moment_repo.go          # MomentRepository + moment_embedding_vectors 写入
    │   ├── echo_candidate_reader.go # pgvector dense Echo 候选读取
    │   ├── moment_sparse_search.go # pg_trgm sparse Echo 候选读取
    │   ├── vector.go               # pgvector literal 校验/格式化
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
Moment  { ID, TraceID, UserID, Content, Embeddings, CreatedAt }
Echo    { ID, TraceID, MomentID, Content, MatchedMomentIDs, Similarities, CreatedAt }
Insight { ID, MomentID, EchoID, TraceID, Text, CreatedAt }
```

## CreateMoment 用例流程 (`app/create_moment.go`)

1. traceId 为空 → 创建新 Trace（`Motivation: "direct"`）
2. traceId 非空 → 追加到已有 Trace（"顺着再想想"）
3. 调用 Embedding 生成向量，创建 Moment；Postgres adapter 同步写入 `moments.embeddings` JSONB 与 `moment_embedding_vectors`
4. Echo 候选召回：
   - dense: `moment_embedding_vectors.embedding <=> current`，同用户同模型最近邻，排除当前 moment
   - sparse: `pg_trgm similarity(moments.content, current.Content)`，排除当前 moment 和同 trace moments
   - `echo_hybrid.go` 使用 Reciprocal Rank Fusion (RRF) 合并候选
5. `DefaultEchoMatcher` 对融合候选重新计算余弦相似度并按阈值过滤，持久化 Echo
6. GenerateInsight 可读取 Echo 匹配到的历史 Moment 原文作为 LLM 上下文

## Echo 匹配 (`app/echo_matcher.go`)

- 候选来源由 `CreateMomentUseCase` 注入：`EchoCandidateReader`（dense）+ `EchoSparseCandidateReader`（sparse）
- 默认阈值 `echoSimilarityThreshold = 0.65`
- 返回匹配 moment IDs + 相似度分数；候选日志仅输出短 preview

## GenerateInsight (`app/generate_insight.go`)

- 基于当前 moment + echo → LLM 生成个性化洞察
- 若 Echo 有匹配历史 moment，会最多读取 3 条原文放入 prompt；取不到原文时明确要求模型不要编造历史内容
- 使用 `platform/ai.ChatWithRetry` 做轻量重试

## 跨模块接口 (`domain/ports.go`)

```go
MomentReader interface { ... }  // 供 timeline/starmap/conversation 使用
TraceReader interface { ... }   // 供 starmap 使用
EchoCandidateReader interface { FindNearestMoments(...) }      // dense 召回
EchoSparseCandidateReader interface { SearchMomentIDs(...) }   // sparse 召回
MomentSearchIndexer interface { IndexMoment(...) }             // 兼容外部索引抽象，pg_trgm 实现为 no-op
```

这些接口由 `adapter/postgres/reader.go` 实现，被 timeline、starmap、conversation 模块通过 `module.go` 依赖注入。

## 模块组装 (`module.go`)

```go
type Deps struct {
    DB             sqlc.DBTX
    AIClient       *platformai.Client
    EmbeddingDim   int
    EchoRecallTopK int32
    EchoSparseOn   bool
    EchoSparseTopK int32
    EchoHybridRRFK int
}
func NewHandler(deps Deps) *writinggrpc.Handler
```

依赖 `platform/ai` 的 `Client`（提供 Embedding + Chat API），以及 `platform/postgres/sqlc` + raw pgx 查询完成 pgvector/pg_trgm 召回。

## 相关文件

| 文件 | 说明 |
|------|------|
| `server/internal/platform/ai/client.go` | AI API 客户端（Embedding + Chat） |
| `server/internal/platform/ai/retry.go` | AI Chat 重试策略 |
| `server/internal/platform/ai/similarity.go` | 余弦相似度计算 |
| `server/internal/platform/postgres/migrations/012_constellation_profiles.sql` | pgvector 扩展 + moment/profile vector 表 |
| `server/internal/platform/postgres/migrations/015_pgtrgm_search.sql` | pg_trgm 扩展 + GIN trigram 索引 |
| `server/internal/platform/postgres/sqlc/moments.sql.go` | sqlc 生成的 moment 查询 |
| `server/internal/platform/postgres/sqlc/echos.sql.go` | sqlc 生成的 echo 查询 |
| `server/internal/platform/postgres/sqlc/insights.sql.go` | sqlc 生成的 insight 查询 |
| `server/internal/platform/postgres/sqlc/traces.sql.go` | sqlc 生成的 trace 查询 |
| `server/internal/bootstrap/writing.go` | 顶层 wiring |
