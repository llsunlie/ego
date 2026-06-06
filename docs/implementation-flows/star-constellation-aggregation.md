# Star 聚合星座实现流程

## 概述

Star 是用户将一个 Trace 收进星图后形成的实体。Constellation 是若干主题接近的 Star 聚合后的星座资产。当前实现中，`StashTrace` 同步创建 Star 并返回，真正的主题生成、星座匹配和星座资产刷新在后台 goroutine 中异步执行。

本文按当前 worktree 代码核查，覆盖前端触发、后端同步路径、异步聚类路径、待聚合 Star 的展示降级，以及当前实现与产品/协议预期的差异。

## 前端调用链

```text
NowPage EchoSection
  -> beginStash()
    -> NowPageStatus.stashing
    -> NowPage 显示 StashAnimation overlay
  -> StashAnimation.onComplete
    -> NowPageNotifier.completeStash()
      -> EgoClient.stashTrace(ref, traceId)
        -> grpc.EgoClient.stashTrace(StashTraceReq)
      -> 无论成功失败都 reset 到 idle
```

相关文件：

| 层次 | 文件 |
|---|---|
| 此刻页状态 | `client/lib/features/now/providers/now_page_provider.dart` |
| 此刻页 overlay | `client/lib/features/now/now_page.dart` |
| 飞星动画 | `client/lib/features/now/widgets/stash_animation.dart` |
| gRPC 客户端封装 | `client/lib/data/services/ego_client.dart` |

当前前端不会在 `StashTrace` 后主动轮询聚合结果。星图页后续加载时才会调用 `ListConstellations`：

```text
StarmapPage / StarmapNotifier.loadConstellations()
  -> EgoClient.listConstellations(ref)
    -> grpc.EgoClient.listConstellations(ListConstellationsReq)
```

如果后台聚类尚未完成，后端会把未聚类 Star 合成为单星星座返回，前端仍可渲染。

## 后端整体调用链

```text
gRPC: StashTrace
  -> bootstrap/composite.go
    -> starmap/adapter/grpc/handler.go
      -> starmap/app/stash_trace.go StashTraceUseCase.Execute()
        -> 校验 Trace 存在、归属、未 stashed
        -> 读取 Trace 下所有 Moment
        -> stars.Create(Star{topic:"聚合中"})
        -> traceStasher.MarkStashed(traceID)
        -> go clusterWithProfileAsync(trace, star, moments)
        -> 同步返回 Star

后台 clusterWithProfileAsync:
  -> TraceProfileGenerator.Generate(trace, moments)
  -> TraceProfileRepository.Upsert(trace_profile, trace_vector)
  -> stars.UpdateTopic(trace_profile.topic)
  -> ConstellationProfileRepository.FindCandidates(trace_vector.embedding)
  -> 综合评分并选择 primary / secondary 星座
  -> 若没有 strong primary：创建新 Constellation / ConstellationProfile
  -> 写入 constellation_stars membership
  -> 更新 ConstellationProfile 统计和 centroid embedding
```

模块装配在 `server/internal/starmap/module.go`：

```text
sqlc.Queries
  -> StarRepository / ConstellationRepository / TraceStasher
  -> writing/adapter/postgres.Reader 作为 TraceReader
  -> adapter/ai.ConstellationAssetGenerator
  -> adapter/ai.TraceProfileGenerator
  -> adapter/postgres.TraceProfileRepository
  -> adapter/postgres.ConstellationProfileRepository
  -> StashTraceUseCase / ListConstellationsUseCase / GetConstellationUseCase
  -> starmap gRPC Handler
```

## StashTrace 同步部分

`server/internal/starmap/app/stash_trace.go` 的同步路径如下：

```text
1. 从 context["user_id"] 读取当前用户
2. traceReader.GetTraceByID(traceID)
3. 校验 trace.UserID == 当前用户
4. 校验 trace.Stashed == false
5. traceReader.ListMomentsByTraceID(traceID)
6. 创建 Star:
     ID: 新 UUID
     UserID: 当前用户
     TraceID: req.trace_id
     Topic: "聚合中"
7. traceStasher.MarkStashed(traceID)
8. 启动 go clusterWithProfileAsync(...)
9. 返回 StashTraceRes{star}
```

同步部分的失败会直接导致 `StashTrace` RPC 返回错误，Star 不一定创建成功。异步部分失败不会影响已返回的 RPC。

### TraceProfile 与 ConstellationProfile 聚类

P7 后当前模块装配默认走 `clusterWithProfileAsync(trace, star, moments)`。TraceProfile 不再只是旁路材料，而是星座归属的主要匹配依据。

```text
clusterWithProfileAsync
  -> adapter/ai.TraceProfileGenerator.Generate(trace, moments)
       -> LLM 生成结构化 TraceProfile
       -> 生成失败时应用层最多尝试 3 次
       -> 生成器内部仍保留 minimal profile 兜底
       -> embedding(profile_text)
  -> adapter/postgres.TraceProfileRepository.Upsert(profile, vector)
  -> adapter/postgres.ConstellationProfileRepository.FindCandidates(...)
  -> rankConstellationCandidates(...)
  -> createConstellationFromTraceProfile(...) 或 attachStarToConstellation(...)
```

当前 TraceProfile 字段：

```text
topic
summary
keywords
emotions
scenes
central_pattern
representative_moment_index
profile_text
status
retry_count
last_error
```

`representative_moment_index` 是 LLM 输出的代表性 Moment 序号，后端会映射并校验成持久化字段 `representative_moment_id`；旧版 `representative_moment_id` 输出只作为兼容兜底。`central_pattern` 表示 trace 中的核心模式、关注点或处境结构，允许为空；它不是强制的“冲突”。如果 embedding 失败，会持久化 `status=failed` 的 profile，但不会写入 `trace_profile_vectors`。

P7.1 后新增 `pattern_tags`：

```text
pattern_tags
```

`pattern_tags` 是算法匹配用的短标签集合，用于表达经历方式、处境结构或反复模式。它替代 `central_pattern_overlap` 成为模式维度的主要 overlap 信号；`central_pattern` 继续保留为可读描述和后续画像重写材料。

### P7 当前归属模型

当前运行路径：

```text
Trace -> Star
Trace -> TraceProfile
TraceProfile -> match ConstellationProfile
Star <-> Constellation
```

关键变化：

- `Trace -> Star` 是一对一。Star 是 Trace 被收进星图后的展示节点。
- `Star <-> Constellation` 是多对多。同一个 Star 可以从不同视角加入多个星座。
- 多对多关系由 `constellation_stars` 表表达，同时为了 proto 兼容继续同步维护 `constellations.star_ids`。
- `TraceProfile` 保持内容画像命名，不改成 `StarProfile`。
- `ConstellationProfile` 表示星座长期主题画像，匹配依据从短 topic 升级为画像匹配。
- `constellations` 继续作为 proto / 前端兼容的星座主表。

更完整的设计和实现记录见 `docs/matching-optimization/constellation-profile-design.md` 与 `docs/matching-optimization/constellation-matching-design.md`。

### P7.1 / P7.2 下一步设计

基于真实测试数据，P7 第一版存在偏向创建新星座的问题。后续拆为两步优化：

```text
P7.1:
  -> 已实现 TraceProfile / ConstellationProfile pattern_tags
  -> 已用 pattern_tags_overlap 替代 central_pattern_overlap
  -> 已调整评分权重与阈值
  -> 已让单 Trace 星座避免 profile_similarity 与 centroid_similarity 重复计权
  -> 已增加解释性 middle 规则和候选级决策日志

P7.2:
  -> borderline top3 LLM 判断
  -> 完善 secondary 多视角归属规则
  -> 引入 ConstellationProfile Elasticsearch sparse 召回
  -> dense + sparse RRF 融合后进入综合评分
```

P7.1 / P7.2 设计详见 `docs/matching-optimization/constellation-matching-design.md`。

### Trace.stashed 写入说明

项目设计上 Writing 拥有 `traces` 表，但当前实现中 `starmap/adapter/postgres.TraceStasher` 会直接更新 `traces.stashed=true`。这是当前代码里的跨模块写入例外，文档和后续重构应将它视为需要特别留意的边界点。

## 旧 topic 异步聚类

P7 后旧 `topic -> ConstellationMatcher` 异步聚类路径已移除，不再作为 fallback。Star topic 会从 TraceProfile.topic 更新；星座归属由 TraceProfile 与 ConstellationProfile 匹配决定。

当前异步聚合关键失败会记录 error 日志并带上 `recovery=pending_message_queue`。后续 P8 会设计消息队列或补偿任务，保证 TraceProfile、membership 与 ConstellationProfile 更新最终一致。

## 星座资产生成

`server/internal/starmap/adapter/ai/constellation_asset_generator.go` 一次生成：

| 产物 | 用途 |
|---|---|
| `topic` | 后续匹配使用的核心主题 |
| `topicEmbedding` | topic embedding 缓存 |
| `name` | 前端展示的星座名 |
| `insight` | 星座级观察 |
| `prompts` | 话题引子 |

AI Chat 返回应为 JSON：

```json
{"topic":"疲惫拖延","name":"总是拖延","insight":"你反复写到的是一种提不起劲的状态。","prompts":["什么时候最明显？","最想被谁理解？"]}
```

当前实现会：

- 去掉可能包裹的 markdown code fence。
- 裁剪 `topic` 到 20 rune、`name` 到 8 rune、`insight` 到 80 rune。
- `prompts` 最多保留 3 条并过滤空字符串。
- 对 `topic` 再调用 Embedding API，成功则缓存到 `topicEmbedding`。

Chat 失败、JSON 解析失败都会使用 fallback 默认资产且返回 `nil` error。topic embedding 失败只会记录 warn，星座仍会创建或更新。

`starmap/app/constellation_asset_generator.go` 仍存在 MVP 默认实现，但当前模块装配使用的是 `starmap/adapter/ai.ConstellationAssetGenerator`。

## ListConstellations 展示降级

`server/internal/starmap/app/list_constellations.go` 不只是返回真实 Constellation，还会把未被任何 Constellation 引用的 Star 合成为“单星星座”：

```text
1. constellations.FindAllByUserID(userID)
2. 汇总所有真实 constellation.star_ids
3. stars.FindAllByUserID(userID)
4. 对未被引用的 Star 追加一个合成 Constellation:
     ID = star.ID
     Name = star.Topic
     Topic = star.Topic
     StarIDs = [star.ID]
     ConstellationInsight = "正在分析这些想法，稍后就会汇聚成星座…"
5. total_star_count = 真实星座内 star 数 + 未聚类 star 数
```

这使得后台聚类尚未完成时，前端星图仍能立即看到用户刚收进去的星。

## GetConstellation 待聚合兼容

`server/internal/starmap/app/get_constellation.go` 先按 constellation ID 查找。如果找不到，会再按相同 ID 查 Star：

```text
constellations.FindByID(id)
  -> not found
  -> stars.FindByIDs([id])
  -> 如果 Star 属于当前用户，返回合成单星详情
```

合成详情包含：

- `Constellation.ID = Star.ID`
- `Name/Topic = Star.Topic`
- `StarIDs = [Star.ID]`
- `Stars = [Star]`
- `Moments = nil`

如果查到真实 Constellation，则会批量读取星座内 Star，再按每个 Star 的 TraceID 读取对应 Moment。

## 前端星图详情与对话入口

星图详情页调用：

```text
ConstellationDetailPage._load()
  -> client.stub.getConstellation(GetConstellationReq)
```

返回后前端使用：

- `res.constellation.constellationInsight` 渲染“我发现”
- `res.constellation.topicPrompts` 渲染“我想听你说更多”
- `res.stars` 渲染“和那时的自己说说话”

话题引子点击后，前端只做本地状态跳转：

```text
TopicPromptCard.onTap
  -> pendingTopicPromptProvider = prompt
  -> tabProvider.setIndex(0)
  -> context.go('/now')
```

Now 页会把 prompt 放进输入框 hint，并不会把 constellation ID 或 motivation 传给 `CreateMoment`。因此当前代码没有形成 `motivation = "constellation:<id>"` 的后端记录。

## 错误处理与降级

| 阶段 | 当前行为 |
|---|---|
| Trace 不存在、归属不符、已 stashed | `StashTrace` 同步返回错误 |
| Star 创建失败 | `StashTrace` 同步返回错误 |
| MarkStashed 失败 | `StashTrace` 同步返回错误；已创建 Star 不会自动回滚 |
| Topic Chat 失败 | 降级为 `"未命名的星"`，继续聚类 |
| 新 topic embedding 失败 | 视为无匹配，继续创建新星座 |
| 单个已有星座 embedding 失败 | 跳过该星座 |
| 资产 Chat/JSON 失败 | 使用 fallback 默认资产 |
| 资产 topic embedding 失败 | 无缓存 embedding，仍创建/更新星座 |
| TraceProfile LLM 失败 | 最多重试 2 次，仍失败则生成 fallback profile |
| TraceProfile embedding 失败 | 写入 `status=failed` profile，不写 vector，不影响当前星座聚类 |
| TraceProfile 持久化失败 | 只记日志，不影响当前星座聚类 |
| 后台任一步骤返回 hard error | 只记日志，Star 可能长期停留在未聚类/合成单星状态 |

当前没有任务队列、重试机制、dead-letter 或后台 reconciliation。服务重启会丢失正在执行的 goroutine；失败的 Star 后续会持续以合成单星星座出现，直到人工修复或未来补偿任务处理。

## 当前限制

- `StashTrace` 中 Star 创建成功但 `MarkStashed` 失败时没有补偿删除 Star。
- 异步 goroutine 使用 `context.Background()`，日志上下文不会继承原请求的 request_id/user_id。
- 前端没有主动轮询聚合完成状态，只有进入/刷新星图页时才重新加载。
- 当前存在 app 层 MVP generator/matcher 和 adapter/ai 真实实现两套策略；模块装配使用 AI 版本。
- `TraceStasher` 对 `traces.stashed` 的写入是当前代码中的跨模块写例外，需要在后续架构演进中重点关注。

## 涉及文件

| 范围 | 文件 |
|---|---|
| Proto 契约 | `proto/ego/api.proto` |
| 前端 Now 状态 | `client/lib/features/now/providers/now_page_provider.dart` |
| 前端飞星动画 | `client/lib/features/now/widgets/stash_animation.dart` |
| 前端星图状态 | `client/lib/features/starmap/providers/starmap_provider.dart` |
| 前端星座详情 | `client/lib/features/starmap/constellation_detail_page.dart` |
| 前端话题引子 | `client/lib/features/starmap/widgets/topic_prompt_card.dart` |
| gRPC 聚合路由 | `server/internal/bootstrap/composite.go` |
| Starmap Handler | `server/internal/starmap/adapter/grpc/handler.go` |
| StashTrace 用例 | `server/internal/starmap/app/stash_trace.go` |
| ListConstellations 用例 | `server/internal/starmap/app/list_constellations.go` |
| GetConstellation 用例 | `server/internal/starmap/app/get_constellation.go` |
| AI Topic 生成 | `server/internal/starmap/adapter/ai/topic_generator.go` |
| AI 星座匹配 | `server/internal/starmap/adapter/ai/constellation_matcher.go` |
| AI 资产生成 | `server/internal/starmap/adapter/ai/constellation_asset_generator.go` |
| AI TraceProfile 生成 | `server/internal/starmap/adapter/ai/trace_profile_generator.go` |
| TraceProfile 持久化 | `server/internal/starmap/adapter/postgres/trace_profile_repo.go` |
| TraceProfile 迁移 | `server/internal/platform/postgres/migrations/011_trace_profiles.sql` |
| MVP 策略 | `server/internal/starmap/app/topic_generator.go`, `server/internal/starmap/app/constellation_matcher.go`, `server/internal/starmap/app/constellation_asset_generator.go` |
| Postgres 适配 | `server/internal/starmap/adapter/postgres/` |
| Writing 只读适配 | `server/internal/writing/adapter/postgres/reader.go` |
| 模块装配 | `server/internal/starmap/module.go` |
