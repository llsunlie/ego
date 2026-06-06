# P7 星座匹配流程第一版实施记录

本文记录 P7 第一版实现边界。P7 将 `StashTrace` 后台聚类从短 topic matching 切换为 TraceProfile -> ConstellationProfile matching，并引入 Star 与 Constellation 的多对多归属关系。

## 当前链路

当前模块装配已注入 TraceProfile 与 ConstellationProfile 仓储，因此 `StashTrace` 后台只走新链路：

```text
StashTrace
  -> create Star(topic="聚合中")
  -> mark Trace stashed
  -> async clusterWithProfileAsync
       -> TraceProfileGenerator.Generate(trace, moments), failed attempts retry up to 3 total attempts
       -> upsert trace_profiles / trace_profile_vectors
       -> stars.UpdateTopic(trace_profile.topic)
       -> recall ConstellationProfile candidates
       -> score candidates
       -> attach primary Constellation or create new Constellation
       -> optionally attach secondary Constellations
       -> upsert constellation_profiles / constellation_profile_vectors
       -> upsert constellation_stars
```

旧 `topic -> ConstellationMatcher` 聚类路径已移除。P7 后不再通过短 topic 做星座匹配 fallback。

## 新增持久化

迁移：

`server/internal/platform/postgres/migrations/012_constellation_profiles.sql`

新增表：

- `constellation_profiles`
- `constellation_profile_vectors`
- `constellation_stars`

`constellations` 仍作为 proto / 前端兼容主表，当前仍同步维护 `star_ids`，因此前端无需改动。

## 匹配策略

候选召回：

```text
TraceProfile.profile_embedding
  vs
ConstellationProfile.profile_embedding
```

当前第一版使用数据库 topK 召回，默认 topK 为 10。

综合评分：

```text
0.55 * profile_similarity
+ 0.25 * centroid_similarity
+ 0.08 * keyword_overlap
+ 0.05 * scene_overlap
+ 0.04 * emotion_overlap
+ 0.03 * central_pattern_overlap
```

阈值：

```text
strong >= 0.78
middle >= 0.68
```

决策：

- 若最佳候选达到 strong，作为 primary 加入已有星座。
- 若没有 strong 候选，创建新星座作为 primary。
- middle 以上且相似维度与 primary 不同的候选，可作为 secondary。
- 每个 Star 最多 1 个 primary，最多 2 个 secondary。

## P7.1 迭代：减少过度新建星座

状态：已实现。

P7 第一版在真实测试中暴露出一个明显问题：算法偏保守，容易把语义接近的 Trace 拆成多个新星座。例如“明天就入职，有点紧张和担心呢”与已有“入职前的期待与担心”在关键词、场景、情绪上都有明显重合，但由于最终分数低于 `middle_threshold=0.68`，仍被创建为新星座。

P7.1 的目标不是引入复杂 LLM rerank，而是先用结构化信号和更合理的阈值修正过度拆分。

### 新增 pattern_tags

`central_pattern` 保持为一句人可读的模式描述，用于日志、画像重写和人工复核。它不再直接承担主要算法 overlap。

新增 `pattern_tags` 作为算法比较字段：

```json
{
  "central_pattern": "面对即将开始的新工作感到紧张和担心",
  "pattern_tags": ["新开始", "工作变化", "不确定性", "紧张", "担心"]
}
```

`pattern_tags` 的定位：

- 描述 Trace 或 Constellation 中反复出现的经历方式、情绪模式、处境结构。
- 使用短中文标签，避免自然语言句子分词失败。
- 不等同于 `keywords`。`keywords` 偏内容词，`pattern_tags` 偏模式词。
- 不要求每条 Trace 都有复杂模式，但至少可输出 1 到 5 个克制标签。

字段落点：

- `trace_profiles.pattern_tags jsonb`
- `constellation_profiles.pattern_tags jsonb`
- `profile_text` 纳入 `pattern_tags`，但保持 `central_pattern` 作为独立可读字段。

兼容策略：

- 旧 profile 没有 `pattern_tags` 时按空数组处理。
- 正式启用前数据库可清空；同时为了兼容当前本地库，已增加 `013_profile_pattern_tags.sql` 为已有表补列。

### pattern_tags 生成规范

TraceProfile 生成 prompt 需要增加：

```text
pattern_tags: 1 到 5 个短标签，用来描述这次 trace 的经历方式、处境结构或反复模式。
不要重复 keywords。
不要做医学化、诊断化、性格定性。
如果内容只是日常记录，也可以输出轻量标签，例如 ["生活安顿"]、["新开始"]。
```

ConstellationProfile 初始化时直接继承 TraceProfile.pattern_tags。

加入已有星座时，P7.1 仍使用轻量合并：

```text
pattern_tags = merge_top_unique(existing.pattern_tags, trace.pattern_tags, max=8)
```

这只是过渡方案。标签频次、权重、衰减和 LLM 画像重写放到 P8。

### pattern_tags_overlap

用 `pattern_tags_overlap` 替换当前 `central_pattern_overlap`。

当前计算：

```text
intersection = trace.pattern_tags ∩ constellation.pattern_tags
denominator = min(len(trace.pattern_tags), len(constellation.pattern_tags))
pattern_tags_overlap = len(intersection) / denominator
```

边界：

- 任一侧为空时返回 `0`。
- 标签比较先做 trim、去空、去重。
- 暂不做同义词归一，例如“新开始”和“入职变化”先视为不同标签。

后续如果观察到标签表述仍然分散，再进入 P8/P10 讨论标签归一化或 LLM 辅助重写。

### P7.1 评分公式

当前权重：

```text
score =
0.45 * profile_similarity
+ 0.20 * centroid_similarity
+ 0.12 * keyword_overlap
+ 0.08 * scene_overlap
+ 0.07 * emotion_overlap
+ 0.08 * pattern_tags_overlap
```

单 Trace 星座的特殊处理：

```text
if constellation.trace_count <= 1:
  score =
  0.60 * profile_similarity
  + 0.14 * keyword_overlap
  + 0.10 * scene_overlap
  + 0.08 * emotion_overlap
  + 0.08 * pattern_tags_overlap
```

原因：只有一个 Trace 的星座中，`profile_similarity` 与 `centroid_similarity` 通常来自同一条 TraceProfile，重复计权会让 embedding 信号被放大到 80%，结构化证据反而太弱。

### P7.1 阈值与组合规则

当前阈值：

```text
strong_threshold = 0.72
middle_threshold = 0.60
```

新增解释性通过规则：

```text
if score >= 0.58
and score < 0.60
and count_positive(keyword_overlap, scene_overlap, emotion_overlap, pattern_tags_overlap) >= 3:
  allow_middle_match
```

含义：

- 如果 embedding 分数略低，但关键词、场景、情绪、模式标签中至少 3 类给出正证据，不直接新建星座。
- 该规则只能把候选提升到 middle，不直接提升到 strong。
- primary 仍选择最终排序最高的候选。

P7.1 的主要验收样本：

```text
输入：明天就入职，有点紧张和担心呢
已有：入职前的期待与担心
期望：进入已有星座，或至少作为边界候选进入后续判断，不应直接新建“入职前焦虑”。
```

### P7.1 日志要求

每个候选的 debug 日志应包含：

```text
constellation_id
topic
score
profile_similarity
centroid_similarity
keyword_overlap
scene_overlap
emotion_overlap
pattern_tags_overlap
matched_keywords
matched_scenes
matched_emotions
matched_pattern_tags
trace_count
score_components
strong_threshold_gap
middle_threshold_gap
decision_hint
```

最终决策日志应包含：

```text
decision
primary_constellation_id
primary_score
created_constellation_id
secondary_count
thresholds
explainable_middle_rule_applied
```

## P7.2 迭代设计：边界判断与混合召回

P7.2 的目标是处理“有点像，但不是完全同一个主题”的候选，并进一步减少候选召回遗漏。

### Borderline LLM 判断

只在边界区间触发：

```text
0.55 <= top_score < strong_threshold
```

输入 top3 候选星座：

- 当前 TraceProfile
- 每个候选 ConstellationProfile
- 每个候选的关键词、场景、情绪、pattern_tags 命中情况
- 当前确定性 score

输出：

```json
{
  "decision": "primary_match | secondary_match | create_new",
  "primary_constellation_id": "optional",
  "secondary_constellation_ids": ["optional"],
  "reason": "简短说明"
}
```

约束：

- LLM 不参与全量召回，只判断 top3 边界候选。
- LLM 结果不能绕过强负例：如果候选无关键词、无场景、无情绪、无 pattern_tags 命中，不能被选为 primary。
- LLM 失败时回到 P7.1 确定性决策。

### Secondary 多视角归属

P7.2 完善 secondary 选择：

```text
primary: 最主要归属，weight=1.0
secondary: 其他合理视角，weight=0.5
```

secondary 条件：

- 分数达到 `middle_threshold`，或被 borderline LLM 判定为 secondary。
- 与 primary 的 `match_dimensions` 不能完全相同。
- 最多 2 个 secondary。

例如同一条入职 Trace 可以：

```text
primary: 入职前的期待与担心
secondary: 工作身份转变
```

### ConstellationProfile ES sparse 召回

P7.2 可以引入星座级 sparse recall，类似 Echo 的 dense + sparse hybrid recall。

索引对象是 `ConstellationProfile`，不是原始 Star 内容：

```text
constellation_id
user_id
topic
name
summary
keywords
emotions
scenes
pattern_tags
central_pattern
profile_text
trace_count
updated_at
```

查询文本来自当前 TraceProfile：

```text
topic + summary + keywords + emotions + scenes + pattern_tags + central_pattern
```

召回流程：

```text
dense: pgvector profile_embedding topK
sparse: ES ConstellationProfile topK
fused: RRF(dense, sparse)
score: P7.1 综合评分
borderline: optional LLM
```

ES 的定位是扩大候选召回，不直接决定最终归属。最终归属仍由综合评分与边界判断决定。

## P8 衔接

P7.1 会引入 `pattern_tags`，但仍只做轻量合并。P8 需要进一步解决：

- `pattern_tags`、`keywords`、`emotions`、`scenes` 的频次与权重统计。
- 低辨识度标签过滤。
- 标签同义归一。
- 加入新 Trace 后异步 LLM 重写 ConstellationProfile。
- 使用消息队列或补偿任务保证 profile、membership、向量更新最终一致。

## 多归属写入

每个归属写入 `constellation_stars`：

```text
constellation_id
star_id
trace_id
user_id
match_score
match_type        primary / secondary
match_dimensions
match_reason
weight
created_at
```

画像更新权重：

```text
primary   weight = 1.0
secondary weight = 0.5
```

同时为了兼容现有 proto 和 read model，会把 `star_id` 同步追加到对应 `constellations.star_ids`。

## 新建星座

没有 strong 候选时：

```text
TraceProfile
  -> Constellation
  -> ConstellationProfile
  -> ConstellationProfileVector
  -> constellation_stars(primary)
```

新星座的算法画像直接来自 TraceProfile：

- `topic`
- `summary`
- `keywords`
- `emotions`
- `scenes`
- `central_pattern`

展示资产仍通过现有 `ConstellationAssetGenerator` 生成 name / insight / prompts，但 `constellations.topic` 以 TraceProfile.topic 为准。

## 加入已有星座

加入已有星座时：

- `constellation_stars` upsert membership。
- `constellations.star_ids` 追加当前 star。
- `constellation_profiles` 合并 keywords / emotions / scenes。
- `constellation_profile_vectors.centroid_embedding` 做加权平均。
- `profile_embedding` 暂时保留已有画像 embedding，不在 P7 同步刷新 LLM 画像。

## 查询兼容

`ListConstellations` 继续返回 `Constellation` 列表。

由于一个 Star 可能属于多个星座，`TotalStarCount` 改为按唯一 Star 计数，避免多归属导致总数重复。

`GetConstellation` 仍通过 `constellations.star_ids` 找到星座内 Stars 和 Moments。

## 当前暂缓

P7 第一版暂不做：

- 不扩展 proto。
- 不展示 secondary membership。
- 不做异步 LLM 刷新 ConstellationProfile。
- 不做 `centroid_embedding` 的 pgvector topK 第二路召回。
- 不做 `constellations.star_ids` 查询时聚合替换；当前仍同步维护数组。
- 不引入 ANN index。当前 4096 维 pgvector 仍按普通排序召回。
- 不引入消息队列保证异步一致性。当前异步聚类关键失败会记录 `recovery=pending_message_queue`，后续通过队列或补偿任务补齐。

## P8 优化标记

当前 `mergeConstellationProfile` 仍是第一版轻量合并：keywords / emotions / scenes 去重并保留 topN，summary / topic / central_pattern 只在为空时补齐。这能避免字段无限增长，但还不足以让星座画像长期保持辨识度。

P8 需要专门设计：

- 画像标签的频次/权重统计，而不是简单去重。
- topN 保留和过泛词过滤。
- 加入新 Trace 后的异步 LLM 画像刷新。
- 代表性 Moment 选择与更新。
- 使用消息队列或补偿任务保证 TraceProfile、membership、ConstellationProfile 更新的一致性。

## 验证

新增测试覆盖：

- profile clustering 创建 primary constellation。
- profile clustering 不走旧 topic generator / constellation matcher。
- 多归属时 `ListConstellations.TotalStarCount` 按唯一 Star 计数。

验证命令：

```text
go test ./internal/starmap/...
```
