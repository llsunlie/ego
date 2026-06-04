# P2 Echo 匹配质量优化设计与实施记录

## 文档定位

本文覆盖 P2「Echo 匹配质量优化」。

P2 不改变 P1 的 pgvector/HNSW topK 召回路径，也不引入 Elasticsearch。P2 的目标是在 P1 候选集上增加确定性规则，减少同 Trace 内容和连续重复候选对 Echo 体验的干扰。

## 已确认约束

1. Echo 必须实时返回。
2. 不使用 insight embedding。
3. 不修改 proto。
4. 不修改 proto，前端继续显示 `similarities`，其语义升级为最终 `echo_score`。
5. P2 不使用手写关键词表。
6. Elasticsearch sparse search 进入 P2.5 单独设计。

## 目标链路

```text
当前 Moment
  -> P1 pgvector topK 候选
  -> P2 规则过滤
  -> P2 echo_score 排序
  -> P2 去重与截断
  -> 保存 Echo
```

## 规则设计

### 排除当前 Trace

同一个 Trace 内的 Moment 往往是用户刚刚连续写下的内容。返回同 Trace 内容会让 Echo 像重复当前上下文，而不是“过去的自己回应我”。

规则：

```text
candidate.trace_id == current.trace_id -> 排除
```

### 时间距离轻量加权

P2 使用 `echo_score` 排序，并通过既有 `similarities` 字段返回给前端。

第一版时间加权：

| 时间距离 | score 调整 |
|---|---:|
| 0 - 24 小时 | -0.01 |
| 1 天 - 7 天 | +0.01 |
| 7 天 - 90 天 | +0.005 |
| 90 天以上 | +0.01 |

### 同 Trace 去重

P1 以后 P2.5 可能会从 dense 和 sparse 两路召回同一批候选。即使当前 P2 只处理 dense 候选，也应先建立去重规则。

规则：

```text
同一个 candidate.trace_id 只保留 echo_score 最高的候选
```

如果候选没有 `trace_id`，则按独立候选处理，保持测试和历史兼容。

### 返回数量上限

最多返回 3 条 matched moments。

理由：

- Echo 应该有代表性，不应一次返回大量相似历史。
- 前端展示更稳定。
- 后续 ES hybrid recall 和 rerank 可复用该上限。

## 分数字段语义

P2 内部使用：

```text
echo_score = cosine_similarity + time_adjustment
```

Echo 持久化的 `similarities` 保存：

```text
echo_score
```

raw cosine similarity 不再作为 Echo 返回值保存，只保留在候选 score 调试日志中用于排查。

## 不在 P2 中处理

- 不引入 Elasticsearch。
- 不做 BM25 / sparse search。
- 不做手写关键词表。
- 不做 LLM rerank。
- 不修改 proto。
- 不引入强/弱/无 Echo 前端分档。

## 验收标准

1. 同 Trace 候选不会作为 Echo 返回。
2. 同一历史 Trace 多个候选只保留一个。
3. 最多返回 3 条 Echo matched moments。
4. `similarities` 保存最终 `echo_score`。
5. Writing app 测试通过。
