# P6 ConstellationProfile 设计方案

本文记录 P6 的目标设计。它基于当前讨论结论，不描述当前已实现行为。

P6 的核心目标是把星座从短 `topic` 聚合升级为长期主题画像聚合，同时保持现有 proto / 前端返回结构兼容。

## 关键决策

- 不考虑历史星座数据迁移；正式启用前会清空数据库。
- 保留 `constellations` 作为星座实体主表和 proto 兼容输出来源。
- 不保留旧 topic matching 作为长期新流程的一部分。
- `Trace -> Star` 是一对一：一个 Trace 被收进星图后生成一个 Star。
- `Star <-> Constellation` 是多对多：一个 Star 可以从多个视角进入多个星座。
- 内容画像继续命名为 `TraceProfile`，不改成 `StarProfile`。
- 星座长期画像命名为 `ConstellationProfile`。
- 多对多归属关系由 `constellation_stars` 表表达。

## 概念关系

```text
Trace
  -> Star

Trace
  -> TraceProfile

TraceProfile
  -> match ConstellationProfile
  -> create/update Constellation

Star
  <-> Constellation
```

含义：

- `Trace` 是用户一次连续书写。
- `Star` 是 Trace 被收进星图后的展示节点。
- `TraceProfile` 是 Trace 的内容画像，回答“这次写了什么”。
- `Constellation` 是星座展示实体，继续服务 proto / 前端兼容。
- `ConstellationProfile` 是星座长期主题画像，服务算法匹配和画像演化。
- `constellation_stars` 是 Star 与星座的真实归属关系。

## 为什么保留 TraceProfile 命名

画像描述的是一次书写内容，而不是星星本身。

TraceProfile 的生成材料来自：

- trace moments
- moment 顺序
- trace motivation
- 后续可选 echo / insight 材料

Star 的职责是“这个 Trace 已经进入星图”，更接近展示节点和归档实体。

如果命名为 `StarProfile`，容易混淆为星星视觉属性、星图布局状态或 Star 在某个星座里的关系画像。多对多归属下，这种混淆会更明显。

因此推荐命名保持：

```text
TraceProfile
ConstellationProfile
constellation_stars
```

如果后续需要描述“这个 Star 为什么加入这个星座”，由 membership/link 表字段表达，而不是新增 StarProfile。

## 数据结构方向

### constellations

继续作为星座主表和 proto 兼容输出来源。

```text
id
user_id
topic
name
constellation_insight
topic_prompts
created_at
updated_at
```

字段职责：

- `topic`：兼容旧 proto 字段，后续直接来自 `ConstellationProfile.topic`。
- `name`：展示名，可以比 topic 更有表达感。
- `constellation_insight`：展示文案，后续可由 ConstellationProfile 异步生成。
- `topic_prompts`：展示或生成视觉资产所需 prompt。

旧的 `topic_embedding` 不再作为新匹配流程的核心依据，可在新流程中废弃。

### constellation_profiles

算法画像表。

```text
constellation_id
user_id
topic
summary
keywords jsonb
emotions jsonb
scenes jsonb
central_pattern
pattern_tags jsonb
profile_text
trace_count
moment_count
status
last_error
created_at
updated_at
```

字段职责：

- `topic`：稳定、直接的主题表达。
- `summary`：星座长期在讲什么。
- `keywords`：长期反复出现的具体词。
- `emotions`：长期常见情绪。
- `scenes`：长期场景。
- `central_pattern`：长期反复出现的经历方式，可为空。
- `pattern_tags`：P7.1 目标字段。用于算法比较的长期模式标签，表达反复出现的经历方式、处境结构或心理动作，避免直接比较中文长句导致 overlap 失效。
- `profile_text`：用于 embedding 和排查。
- `trace_count` / `moment_count`：画像规模统计。
- `status`：画像状态，例如 `ready / refreshing / fallback / failed`。

### constellation_profile_vectors

向量表。向量独立存放，避免把 1024 维 embedding 数据混入常规列表查询表。当前默认模型为 `BAAI/bge-m3`。

```text
constellation_id
user_id
model
dim
profile_embedding
centroid_embedding
created_at
updated_at
```

`profile_embedding`：

- 来源于 `ConstellationProfile.profile_text`。
- 表示 LLM 总结后的主题语义。
- 更适合冷启动和可解释匹配。

`centroid_embedding`：

- 来源于该星座吸收过的 TraceProfile embedding 加权平均。
- 表示星座真实吸收过的数据分布。
- 更适合判断新 Trace 是否接近星座历史成员。

第一版可以主用 `profile_embedding`，同步维护 `centroid_embedding`，后续匹配评分再决定权重。

### constellation_stars

Star 与 Constellation 的多对多归属表。

```text
constellation_id
star_id
trace_id
user_id
match_score
match_type
match_dimensions jsonb
match_reason
weight
created_at
```

字段职责：

- `match_type`：`primary / secondary`。
- `match_dimensions`：相似维度，例如 `["scene", "keyword"]`、`["emotion", "pattern_tags"]`。
- `match_reason`：可解释原因，供日志和后续审核使用。
- `weight`：画像更新权重，建议 primary 为 `1.0`，secondary 为 `0.5`。

`constellations.star_ids` 如果继续保留，可以作为 proto 兼容字段从 `constellation_stars` 聚合得到；算法关系以 `constellation_stars` 为准。

## Star 如何归入星座

归属依据不是 Star 自身，而是 Star 背后的 TraceProfile。

```text
Trace
  -> Star
Trace
  -> TraceProfile
  -> compare ConstellationProfile
  -> write constellation_stars
```

### 候选召回

使用 TraceProfile 向量召回候选星座：

```text
TraceProfile.profile_embedding
  vs
ConstellationProfile.profile_embedding
```

取 topK，例如 5 或 10。

如果已维护 `centroid_embedding`，可以额外召回：

```text
TraceProfile.profile_embedding
  vs
ConstellationProfile.centroid_embedding
```

两路候选合并后进入综合评分。

### 综合评分

第一版可以不引入 LLM judgement，先使用可解释信号：

```text
match_score =
profile_similarity
+ keyword_overlap
+ scene_overlap
+ emotion_overlap
+ pattern_tags_overlap
```

这些信号不在 P6 文档中固定具体权重，后续 P7 实现前再结合样本和日志确定。

P7 第一版使用了 `central_pattern_overlap`，真实中文样本中容易因为长句无法有效分词而长期为 0。P7.1 设计改为引入 `pattern_tags`，用短标签集合 overlap 承担这部分算法信号，`central_pattern` 保留为可读描述。

### 多星座归属

一个 Star 可以进入多个星座，但需要避免“什么都连上”。

建议初始规则：

```text
score >= strong_threshold
  可加入，最高分作为 primary

middle_threshold <= score < strong_threshold
  只有当相似维度与 primary 明显不同，才可作为 secondary

score < middle_threshold
  不加入
```

初始数量限制：

```text
每个 Star 最多 1 个 primary
每个 Star 最多 2 个 secondary
每个 Star 最多加入 3 个星座
```

### 多视角去重

如果多个候选星座的相似原因相同，只保留分数最高的。

例如都只是：

```text
["小红书", "焦虑"]
```

只加入一个。

如果相似维度不同，可以都加入：

```text
社交媒体焦虑：["scene", "keyword"]
工作过渡焦虑：["scene", "central_pattern"]
比较带来的不安：["emotion", "central_pattern"]
```

这类情况才体现多对多的产品价值。

## 新建星座

当没有合格候选时：

```text
TraceProfile
  -> create Constellation
  -> create ConstellationProfile
  -> create ConstellationProfileVector
  -> insert constellation_stars(primary)
```

新 ConstellationProfile 可以直接从 TraceProfile 初始化：

```text
topic = TraceProfile.topic
summary = TraceProfile.summary
keywords = TraceProfile.keywords
emotions = TraceProfile.emotions
scenes = TraceProfile.scenes
central_pattern = TraceProfile.central_pattern
trace_count = 1
moment_count = 当前 trace moment 数
```

展示字段：

```text
constellations.topic = ConstellationProfile.topic
constellations.name = topic 或后续 LLM 展示名
constellations.constellation_insight = summary 或后续 LLM 展示文案
```

## 加入已有星座

当 Star 加入已有星座：

```text
1. insert constellation_stars
2. update constellation_profiles 统计字段
3. update constellation_profile_vectors.centroid_embedding
4. 异步刷新 ConstellationProfile 文本画像
5. 同步或异步更新 constellations 展示字段
```

画像更新权重：

```text
primary membership: weight = 1.0
secondary membership: weight = 0.5
```

这样一个 Star 可以从多个视角进入多个星座，但不会把 secondary 星座过度拉偏。

## ConstellationProfile 更新

建议分两层。

同步轻量更新：

```text
trace_count += weight
moment_count += trace_moment_count * weight
keywords/emotions/scenes 合并计数
centroid_embedding 重新加权平均
updated_at 更新
```

异步 LLM 更新：

```text
旧 ConstellationProfile
+ 新加入的 TraceProfile
+ 代表性 TraceProfiles
-> 新 topic / summary / central_pattern / keywords / emotions / scenes / profile_text
-> 新 profile_embedding
-> 同步 constellations.topic/name/insight
```

异步失败不应破坏已写入的 membership。失败记录到 profile status / last_error，后续可通过补偿任务重试。

## 与 proto 的兼容

proto 兼容通过 `constellations` 主表和 adapter mapper 保持。

新流程不要求 proto 立即感知：

- ConstellationProfile 字段
- match_score
- match_type
- match_dimensions
- weight

这些先作为内部算法数据。

如果前端未来需要展示“这个 Star 也属于哪些星座”或“为什么属于这个星座”，再讨论 proto 扩展。

## P6 边界

P6 只完成设计，不直接替换当前 topic-based matching。

P6 建议产出：

- `ConstellationProfile` 字段设计。
- `constellation_profiles` / `constellation_profile_vectors` / `constellation_stars` 表设计。
- TraceProfile 到 ConstellationProfile 的多对多归属规则。
- 与 proto 兼容关系说明。

P7 再进入：

- 实现候选召回。
- 实现综合评分。
- 实现 primary / secondary 归属。
- 替换当前 topic matching。
- 更新 StashTrace / ListConstellations / GetConstellation 测试。

## 待 P7 具体化的问题

- strong / middle 阈值取值。
- 各评分信号权重。
- `central_pattern_similarity` 如何计算。
- secondary 星座是否参与视觉布局。
- `constellations.star_ids` 是落库同步字段，还是查询时聚合字段。
- 异步画像刷新失败后的重试机制。
