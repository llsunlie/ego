# Echo 回声召回实现流程

## 概述

Echo 是用户写下一条 Moment 后，系统从该用户历史 Moment 中召回与当前处境产生呼应的内容的机制。当前实现已经接入 OpenAI-compatible Embedding API：创建 Moment 时生成 embedding，写入 PostgreSQL JSONB 与 pgvector 向量表，并同步 best-effort 写入 Elasticsearch。Echo 查询时同时使用 pgvector/HNSW dense topK 与 Elasticsearch sparse topK，再通过 RRF 融合候选，最后进入应用层 Echo 规则过滤、排序与持久化。

本文按当前 worktree 代码核查，覆盖后端 RPC 链路、前端触发链路和当前实现与 proto 约定的差异。

## 前端调用链

```text
NowPage / WritingInput
  -> NowPageNotifier.submitMoment(content)
    -> EgoClient.createMoment(ref, content, traceId: currentTraceId)
      -> grpc.EgoClient.createMoment(CreateMomentReq)
    -> 成功后更新 currentTraceId/currentMomentId/echo
    -> 异步触发 _fetchInsight(momentId, echoId)
      -> EgoClient.generateInsight(...)
    -> 若 echo.matched_moment_ids 非空，异步触发 _fetchMatchedMoments(ids)
      -> EgoClient.getMoments(...)
```

相关文件：

| 层次 | 文件 |
|---|---|
| gRPC 客户端封装 | `client/lib/data/services/ego_client.dart` |
| 此刻页状态 | `client/lib/features/now/providers/now_page_provider.dart` |
| 输入组件 | `client/lib/features/now/widgets/writing_input.dart` |
| Echo 展示 | `client/lib/features/now/widgets/echo_card.dart` |

注意：

- 前端在 `CreateMoment` 返回后不会等待 `GenerateInsight` 才展示 Echo；Insight 是后续异步补充。
- `GenerateInsight` 和 `GetMoments` 在前端是可选增强，失败会静默忽略。
- `traceId` 来自 `NowPageState.currentTraceId`，用于“顺着再想想”继续同一个 Trace。

## 后端整体调用链

```text
gRPC: CreateMoment
  -> bootstrap/composite.go
    -> writing/adapter/grpc/handler.go
      -> writing/app/create_moment.go (CreateMomentUseCase.Execute)
        -> createTrace 或校验既有 Trace
        -> writing/adapter/ai/embedder.go
          -> platform/ai/client.go CreateEmbedding()
        -> writing/adapter/postgres/moment_repo.go 保存 Moment
        -> writing/adapter/elasticsearch/moment_search.go best-effort 写入 ES
        -> matchEcho()
          -> 并发启动 dense/sparse 两路候选召回
            -> writing/adapter/postgres/echo_candidate_reader.go pgvector topK 候选召回
            -> writing/adapter/elasticsearch/moment_search.go ES sparse topK 候选召回
          -> writing/app/echo_hybrid.go RRF 融合 dense/sparse 候选
          -> writing/app/echo_matcher.go DefaultEchoMatcher
            -> platform/ai/similarity.go CosineSimilarity
            -> 当前 Trace 排除、同 Trace 去重、echo_score 排序
          -> writing/adapter/postgres/echo_repo.go 保存 Echo
```

模块装配在 `server/internal/writing/module.go`：

```text
sqlc.Queries
  -> TraceRepository / MomentRepository / EchoRepository / InsightRepository / Reader
  -> adapter/ai.Embedder(platform AI client)
  -> app.DefaultEchoMatcher
  -> adapter/ai.InsightGenerator(platform AI client)
  -> CreateMomentUseCase / GenerateInsightUseCase
  -> writing gRPC Handler
```

## CreateMoment 用例流程

`server/internal/writing/app/create_moment.go` 中的 `CreateMomentUseCase.Execute()` 编排四步：

```text
1. 校验 content 非空，并从 context["user_id"] 获取当前用户
2. 如果 req.trace_id 为空：
     创建 Trace，motivation 默认为 "direct"
   否则：
     读取既有 Trace，并校验 trace.UserID == 当前用户
3. 生成 embedding，保存 Moment
4. 并发使用 pgvector dense topK 与 ES sparse topK 读取历史候选
5. 使用 RRF 融合候选，经过 EchoMatcher 规则排序后持久化 Echo
```

关键依赖均通过接口注入：

| 依赖 | 接口 | 当前实现 |
|---|---|---|
| `traces` | `domain.TraceRepository` | `writing/adapter/postgres.TraceRepository` |
| `moments` | `domain.MomentRepository` | `writing/adapter/postgres.MomentRepository` |
| `echoCandidates` | `domain.EchoCandidateReader` | `writing/adapter/postgres.EchoCandidateReader` |
| `searchIndexer` | `domain.MomentSearchIndexer` | `writing/adapter/elasticsearch.MomentSearch` |
| `sparseCandidates` | `domain.EchoSparseCandidateReader` | `writing/adapter/elasticsearch.MomentSearch` |
| `echos` | `domain.EchoRepository` | `writing/adapter/postgres.EchoRepository` |
| `embedding` | `domain.EmbeddingGenerator` | `writing/adapter/ai.Embedder` |
| `echo` | `domain.EchoMatcher` | `writing/app.DefaultEchoMatcher` |
| `ids` | `app.IDGenerator` | `writing/adapter/id.UUIDGenerator` |

### Trace 处理

- 新建 Trace 时 `Motivation` 使用 `CreateMomentInput.Motivation`，为空则降级为 `"direct"`。
- 当前 gRPC `CreateMomentReq` 只有 `content` 和 `trace_id`，handler 未从 proto 接收 motivation，因此普通前端创建的新 Trace 都是 `"direct"`。
- 如果指定 `trace_id`，后端会读取该 Trace 并校验归属；归属不匹配时返回错误。
- 若新建 Trace 后 Moment 创建失败，会调用 `traces.Delete(traceID)` 回滚刚创建的 Trace。

### Moment 和 Embedding

`writing/adapter/ai/embedder.go` 调用 `platform/ai.Client.CreateEmbedding()`，将返回结果映射为：

```json
[{"model":"<embedding model>","embedding":[0.123, "..."]}]
```

该数组保存在 Moment 的 `embeddings` 字段中，支持未来同一 Moment 保存多个模型版本的 embedding。当前 Echo 匹配只取第一个 embedding。

同时，active content embedding 会双写到 `moment_embedding_vectors`：

```text
moment_id
user_id
trace_id
model
dim
embedding VECTOR(1024)
created_at
```

该表用于 pgvector/HNSW topK 候选召回，索引为 `idx_moment_embedding_vectors_embedding_hnsw`。当前默认 embedding 模型为 `BAAI/bge-m3`，输出维度通过 `AI_EMBEDDING_DIM` 配置，默认 `1024`。`moments.embeddings` JSONB 长期保留，用于兼容现有领域模型和跨模块读取。

### ES sparse recall

Moment 保存成功后，Writing 会 best-effort 将 Moment 写入 Elasticsearch `ego_moments` index。PostgreSQL 仍是事实来源，ES 只作为 sparse search 索引。

当前 ES document：

```json
{
  "moment_id": "xxx",
  "user_id": "xxx",
  "trace_id": "xxx",
  "content": "每次都是我先开口，我真的有点累。",
  "created_at": "2026-06-04T10:00:00Z"
}
```

查询时：

- `user_id` 必须匹配当前用户。
- 排除当前 `moment_id`。
- 排除当前 `trace_id`，避免当前 Trace 内内容像重复。
- `content` 使用 IK analyzer 做主召回。
- `content.ngram` 作为中文分词失败的兜底召回字段。

ES 写入失败或 sparse recall 失败不会阻断 `CreateMoment`；失败会记录 warn，并继续使用 pgvector dense 候选。

### 并发候选召回

`matchEcho()` 中 dense 与 sparse 两路候选召回没有数据依赖，当前会并发执行：

```text
current Moment + embedding
  -> goroutine A: pgvector/HNSW dense topK
  -> goroutine B: Elasticsearch sparse topK + PostgreSQL GetByIDs
  -> 等待两路返回
  -> RRF 融合
```

错误语义保持分层：

- pgvector dense 召回是主链路，查询失败会让 `CreateMoment` 返回错误，不回退全量扫描。
- ES sparse 召回是增强链路，查询失败或回读失败只记录 warn，并以空 sparse 候选继续。

### RRF 候选融合

pgvector dense 候选与 ES sparse 候选不直接相加分数。当前使用 RRF 按排名融合：

```text
rrf_score = 1 / (K + dense_rank) + 1 / (K + sparse_rank)
```

默认 `ECHO_HYBRID_RRF_K=60`。如果同一个 Moment 同时出现在 dense 和 sparse 两路，它会因为两路排名贡献而更靠前。

### Echo 匹配算法

`writing/app/echo_matcher.go` 的实际算法：

```text
1. 如果当前 Moment 没有 embedding，返回 nil
2. 取 current.Embeddings[0].Embedding
3. 对 RRF 融合后的候选逐条处理
4. 排除当前 Trace 内候选
5. 跳过没有 embedding 的候选
6. 计算 CosineSimilarity(current, candidate)
7. 计算内部 echo_score = cosine_similarity + time_adjustment
8. 保留 echo_score >= 0.65 的记录
9. 同一历史 Trace 只保留 echo_score 最高的一条
10. 按 echo_score 降序返回最多 3 条
```

时间距离调整：

| 时间距离 | score 调整 |
|---|---:|
| 0 - 24 小时 | -0.01 |
| 1 天 - 7 天 | +0.01 |
| 7 天 - 90 天 | +0.005 |
| 90 天以上 | +0.01 |

匹配结果被转换为 `domain.Echo`：

```text
Echo{
  MomentID: 当前 Moment ID,
  UserID: 当前用户 ID,
  MatchedMomentIDs: 按内部 echo_score 排序的历史 Moment IDs,
  Similarities: 与 MatchedMomentIDs 一一对应的最终 echo_score
}
```

候选召回不再全量扫描用户历史，而是通过 `moment_embedding_vectors` 做 pgvector topK 查询，并通过 ES 做 sparse topK 查询。dense 与 sparse 召回并发执行，因此召回等待时间接近两路中较慢的一路，而不是两路串行相加。应用层复杂度约为 `O(k)`，`k` 由 `ECHO_RECALL_TOP_K`、`ECHO_SPARSE_RECALL_TOP_K` 与融合后候选数量决定。

## 召回日志

Echo 召回日志聚焦算法效果，而不是记录每个基础设施动作。

核心日志：

| 日志 | 目的 | 关键字段 |
|---|---|---|
| `CreateMoment: echo recall candidates` | 查看 dense、ES、RRF 三阶段分别召回了什么 | `current_preview`, `dense_candidates`, `es_candidates`, `fused_candidates`, `*_candidate_count`, `dense_top_k`, `sparse_top_k`, `rrf_k` |
| `CreateMoment: echo final matches` | 查看 EchoMatcher 最终选中了什么 | `matches[].moment_id`, `matches[].trace_id`, `matches[].content_preview`, `matches[].similarity`；这里的 `similarity` 是最终 `echo_score` |
| `echo match candidate scores` | 查看 EchoMatcher 对每个融合候选的计算明细 | `candidates[].similarity`, `candidates[].time_adjustment`, `candidates[].echo_score`, `candidates[].passed_threshold`, `candidates[].skip_reason`；这里的候选 `similarity` 是 raw cosine |
| `echo match done` | 查看规则过滤后的统计结果 | `history_size`, `skipped_no_embedding`, `filtered_same_trace`, `matched`, `top_score`, `top_raw_similarity` |

候选内容只记录 `content_preview`，当前最多 48 个 rune。gRPC composite 层不再记录完整 req/res，避免重复输出用户原文；ES HTTP 成功请求、ES 写入成功、sparse ids loaded、sparse moments loaded、hybrid merged 等中间碎片日志也不记录。

## GetMoments 后续召回

`CreateMomentRes.Echo` 只返回匹配到的历史 Moment ID 和相似度，不包含历史 Moment 内容。前端如果需要展示原文，会调用：

```text
EgoClient.getMoments(ids)
  -> writing/adapter/grpc.Handler.GetMoments
    -> writing/adapter/postgres.Reader.GetByIDs
```

`NowPageNotifier._fetchMatchedMoments()` 会按请求 ID 顺序重新排序，因为 `GetMoments` 返回顺序不保证与输入一致。

`TraceDetailPage` 也会在 `GetTraceDetail` 后收集所有 Echo 的 `matched_moment_ids`，再调用 `GetMoments` 批量补齐历史原文。

## GenerateInsight 关联流程

Echo 返回后，前端异步调用 `GenerateInsight(momentId, echoId)`。当前后端链路：

```text
GenerateInsight
  -> writing/adapter/grpc.Handler.GenerateInsight
    -> writing/app.GenerateInsightUseCase
      -> writing/adapter/ai.InsightGenerator
        -> momentRepo.GetByID(momentID)
        -> echoRepo.FindByMomentID(momentID)
        -> platform/ai.Client.Chat()
      -> insightRepo.Create()
```

需要注意一个实现细节：`InsightGenerator.Generate(ctx, momentID, echoID)` 当前会按 `momentID` 查 Echo，并不直接使用传入的 `echoID` 查库；`echoID` 最终会被写回 `Insight.EchoID`。

## 持久化

| 实体 | 仓库 | 关键字段 |
|---|---|---|
| Trace | `writing/adapter/postgres/trace_repo.go` | `id`, `user_id`, `motivation`, `stashed`, `created_at` |
| Moment | `writing/adapter/postgres/moment_repo.go` | `trace_id`, `user_id`, `content`, `embeddings` |
| Moment vector | `writing/adapter/postgres/moment_repo.go` | `moment_id`, `user_id`, `trace_id`, `model`, `dim`, `embedding` |
| Echo | `writing/adapter/postgres/echo_repo.go` | `moment_id`, `matched_moment_ids`, `similarities` |
| Insight | `writing/adapter/postgres/insight_repo.go` | `moment_id`, `echo_id`, `text`, `related_moment_ids` |

## 错误处理与降级

当前代码行为：

| 场景 | 当前行为 |
|---|---|
| content 为空 | 后端返回错误 |
| context 中没有 user_id | 后端返回错误 |
| embedding 生成失败 | `CreateMoment` 失败；如果刚创建了 Trace，会回滚 Trace |
| 首条 Moment 无历史 | 正常返回，`echo` 为空 |
| 历史中无相似内容 | 正常返回，`echo` 为空 |
| pgvector 候选查询失败 | `CreateMoment` 返回错误，不回退全量扫描 |
| ES 写入失败 | 记录 warn，不阻断 `CreateMoment` |
| ES sparse recall 失败 | 记录 warn，继续使用 pgvector dense 候选 |
| Echo 匹配或 Echo 持久化失败 | `CreateMoment` 返回错误；此时 Moment 已经保存，当前代码不会回滚 Moment |
| Insight 生成失败 | `GenerateInsight` 返回错误；前端此刻页静默忽略 |

proto 中注释过 AI 超时策略，但当前 `CreateMomentUseCase`、`GenerateInsightUseCase` 没有在用例内设置 5s/10s deadline，也没有将 Echo 超时降级为 `echo=null`；实际超时取决于调用方 context 和底层 HTTP client 行为。

## 当前限制

- 当前只使用每个 Moment 的第一个 embedding。
- 当前向量表和 HNSW 索引按 1024 维 `BAAI/bge-m3` 配置；更换为不同维度模型时需要同步数据库迁移和历史向量回填。
- 当前已引入 Elasticsearch sparse search；Echo 候选由 pgvector dense topK 与 ES BM25/IK + ngram sparse topK 通过 RRF 融合后进入 EchoMatcher 规则排序。
- `CreateMoment` 不是完整数据库事务：新建 Trace 后 Moment 失败会回滚 Trace，但 Moment 保存后 Echo 阶段失败不会回滚 Moment。
- `CreateMomentReq` 没有 motivation 字段，因此从星座话题引子回到 Now 页时，前端只能把 prompt 当输入提示，无法把 Trace motivation 标记为 `constellation:<id>`。
- `GenerateInsight` 的生成依据会优先按 `momentID` 查 Echo；传入 `echoID` 不参与查找校验。

## 涉及文件

| 范围 | 文件 |
|---|---|
| Proto 契约 | `proto/ego/api.proto` |
| 前端 gRPC 封装 | `client/lib/data/services/ego_client.dart` |
| 前端 Now 状态 | `client/lib/features/now/providers/now_page_provider.dart` |
| gRPC 聚合路由 | `server/internal/bootstrap/composite.go` |
| Writing Handler | `server/internal/writing/adapter/grpc/handler.go` |
| CreateMoment 用例 | `server/internal/writing/app/create_moment.go` |
| Echo 匹配 | `server/internal/writing/app/echo_matcher.go` |
| Embedding 适配器 | `server/internal/writing/adapter/ai/embedder.go` |
| Insight 适配器 | `server/internal/writing/adapter/ai/insight_generator.go` |
| AI Client | `server/internal/platform/ai/client.go` |
| 相似度 | `server/internal/platform/ai/similarity.go` |
| Postgres 适配 | `server/internal/writing/adapter/postgres/` |
| 模块装配 | `server/internal/writing/module.go` |
