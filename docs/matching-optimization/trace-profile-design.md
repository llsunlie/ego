# P4 TraceProfile 旁路持久化实施记录

## 文档定位

本文记录 P4 TraceProfile 的当前实施边界。

P4 的目标是为后续星座聚合升级准备 Trace 级算法画像。当前阶段只做旁路持久化，不替换现有星座 topic 聚合，不改变 `StashTrace` 返回，不改变 proto 和前端。

> P7 后现状：TraceProfile 已从旁路材料升级为星座聚合主路径，旧 topic 聚合 fallback 已移除。本文保留 P4 设计记录，当前运行链路以 `constellation-matching-design.md` 为准。

## 当前链路

```text
StashTrace
  -> 同步创建 Star(topic="聚合中")
  -> 同步 MarkStashed
  -> 返回 Star
  -> 后台旁路 TraceProfile 生成
       -> LLM 生成 TraceProfile JSON
       -> 拼接 profile_text
       -> embedding(profile_text)
       -> upsert trace_profiles
       -> upsert trace_profile_vectors（如果 embedding 成功）
```

P4 时 TraceProfile 失败不会阻断现有星座聚合。P7 后异步聚合依赖 TraceProfile，生成失败会在应用层重试，仍失败则记录可恢复错误，后续计划由消息队列或补偿任务保证一致性。

## 字段设计

`trace_profiles`

```text
trace_id primary key
user_id
topic
summary
keywords jsonb
emotions jsonb
scenes jsonb
central_pattern
pattern_tags jsonb
representative_moment_id
profile_text
status
retry_count
last_error
created_at
updated_at
```

`trace_profile_vectors`

```text
trace_id primary key
user_id
model
dim
embedding vector(4096)
created_at
updated_at
```

P4 只负责存储 profile embedding，不建立 ANN 索引。当前 active embedding 是 4096 维，pgvector HNSW 对该维度形态有限制；后续 P5/P6 设计正式匹配流程时再决定降维、halfvec、模型切换或其他索引策略。

## 字段语义

- `topic`: 这段 trace 在讲什么。短、稳定、直接。
- `summary`: trace 整体表达了什么。
- `keywords`: 关键词，可为空。
- `emotions`: 情绪词，可为空，不强行制造情绪。
- `scenes`: 场景词，可为空。
- `central_pattern`: 用户在这段 trace 中呈现的核心模式、关注点或处境结构。不是所有 trace 都有冲突，因此允许为空。
- `pattern_tags`: P7.1 目标字段。用于算法比较的短标签，描述 trace 中的经历方式、处境结构或反复模式。它不等同于 `keywords`，也不要求制造复杂心理解释。
- `representative_moment_id`: 持久化字段。生成时 LLM 只输出 `representative_moment_index`，后端根据输入顺序映射成真实 Moment ID；旧版直接输出 ID 时只作为兼容兜底，并会校验是否属于当前 trace moments。
- `profile_text`: 用于 embedding 的拼接文本，也用于排查画像质量。
- `status`: `ready` / `fallback` / `failed`。

## topic 与 central_pattern 的区别

```text
topic = 这段 trace 在讲什么
central_pattern = 这件事里用户怎么在经历它
```

例子：

```text
moment:
这个月无法入职了呢，那只能好好享受最后的生活了

topic:
入职计划延迟

central_pattern:
计划被推迟后，尝试把被动等待转化成主动安排当下

pattern_tags:
["计划变化", "等待", "主动安排"]
```

`central_pattern` 面向可读解释，保留完整句子；`pattern_tags` 面向匹配，保持短、稳定、可做集合 overlap。

## 生成输入

当前 P4 第一版使用：

```text
必选：
- trace moments 原文
- moment 顺序
- trace motivation

预留增强：
- moment insight
- echo matched moment content
```

Insight 和 Echo 暂不作为强依赖，避免 TraceProfile 被不稳定辅助材料污染。后续可以在生成器输入中增强。

## Prompt 约束

TraceProfile prompt 使用“基于证据的压缩”口径：

- 只整理输入中已经出现或有明确证据支撑的信息。
- 不诊断用户，不给建议，不补全背景。
- `topic` 使用短、稳定、直接的日常短语。
- `keywords` 优先使用用户原话附近的具体词，避免“事情、感觉、生活、问题”这类过泛词。
- `emotions`、`scenes` 没有明确依据时输出空数组。
- `central_pattern` 没有明显模式时输出空字符串，不强行制造冲突。
- `pattern_tags` 输出 1 到 5 个短标签，用来描述经历方式、处境结构或反复模式；不要重复 `keywords`，不要医学化、诊断化或性格定性。
- `representative_moment_index` 必须是输入 moments 的序号，后端映射成 `representative_moment_id` 并校验。

## 重试与 fallback

LLM 结构化生成：

```text
最多 3 次尝试 = 首次尝试 + 2 次重试
```

如果仍失败：

```text
status = fallback
topic = 第一条 moment 的短截断
summary = trace moment content 的截断拼接
keywords/emotions/scenes = []
central_pattern = ""
representative_moment_index = 1
representative_moment_id = 后端映射出的第一条 moment id
```

如果 embedding 失败：

```text
status = failed
last_error = embedding error
只写 trace_profiles，不写 trace_profile_vectors
```

## 当前不做

- 不替换当前星座 topic 生成。
- 不替换 constellation matching。
- 不新增 ConstellationProfile。
- 不做 lonely / forming / active 状态。
- 不改变前端展示。
- 不改变 proto。
- 不提供历史回填命令。

## 后续 P5

P5 再讨论：

- ConstellationProfile 字段设计。
- 现有 constellation 如何生成 profile embedding。
- TraceProfile 与 ConstellationProfile 的匹配评分。
- fallback TraceProfile 是否参与正式聚合。
- topic embedding 与 profile embedding 的过渡兼容。
