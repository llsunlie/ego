# Echo 回声召回实现流程

## 概述

Echo 是用户写下一条 Moment 后，系统从该用户历史 Moment 中召回语义相近内容的机制。当前实现已经接入 OpenAI-compatible Embedding API：创建 Moment 时生成 embedding，随后在应用层用余弦相似度全量扫描历史 Moment，匹配结果持久化为 Echo。

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
        -> matchEcho()
          -> MomentRepository.ListByUserID(userID)
          -> writing/app/echo_matcher.go DefaultEchoMatcher
            -> platform/ai/similarity.go CosineSimilarity
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
4. 读取用户全部历史 Moment，排除刚创建的 Moment，计算 Echo 并持久化
```

关键依赖均通过接口注入：

| 依赖 | 接口 | 当前实现 |
|---|---|---|
| `traces` | `domain.TraceRepository` | `writing/adapter/postgres.TraceRepository` |
| `moments` | `domain.MomentRepository` | `writing/adapter/postgres.MomentRepository` |
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

### Echo 匹配算法

`writing/app/echo_matcher.go` 的实际算法：

```text
1. 如果当前 Moment 没有 embedding，返回 nil
2. 取 current.Embeddings[0].Embedding
3. 遍历历史 Moment
4. 跳过没有 embedding 的历史 Moment
5. 计算 CosineSimilarity(current, history)
6. 保留相似度 >= 0.65 的记录
7. 按相似度降序返回
```

匹配结果被转换为 `domain.Echo`：

```text
Echo{
  MomentID: 当前 Moment ID,
  UserID: 当前用户 ID,
  MatchedMomentIDs: 按相似度降序排列的历史 Moment IDs,
  Similarities: 与 MatchedMomentIDs 一一对应
}
```

复杂度为 `O(n)`，`n` 是当前用户历史 Moment 数量。当前实现是应用层全量比对，没有使用 pgvector 向量索引。

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
| Echo 匹配或 Echo 持久化失败 | `CreateMoment` 返回错误；此时 Moment 已经保存，当前代码不会回滚 Moment |
| Insight 生成失败 | `GenerateInsight` 返回错误；前端此刻页静默忽略 |

proto 中注释过 AI 超时策略，但当前 `CreateMomentUseCase`、`GenerateInsightUseCase` 没有在用例内设置 5s/10s deadline，也没有将 Echo 超时降级为 `echo=null`；实际超时取决于调用方 context 和底层 HTTP client 行为。

## 当前限制

- Echo 匹配为应用层全量扫描，历史数据增长后需要引入向量索引或分页候选召回。
- 当前只使用每个 Moment 的第一个 embedding。
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
