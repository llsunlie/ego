# P7 星座匹配流程第一版实施记录

本文记录 P7 第一版实现边界。P7 将 `StashTrace` 后台聚类从短 topic matching 切换为 TraceProfile -> ConstellationProfile matching，并引入 Star 与 Constellation 的多对多归属关系。

## 当前链路

当前模块装配已注入 TraceProfile 与 ConstellationProfile 仓储，因此 `StashTrace` 后台只走新链路：

```text
StashTrace
  -> create Star(topic="聚合中")
  -> mark Trace stashed
  -> async clusterWithProfileAsync
       -> TraceProfileGenerator.Generate(trace, moments), LLM JSON and embedding retry inside generator
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

## P7.2 迭代设计：Codebook 与边界 LLM 判断

状态：已实现。

P7.2 的目标是处理“同一长期主题，但确定性 overlap 分数过低”的候选。

P7.1 后，确定性评分已经可以处理明显相似样本。但真实测试里仍会出现这类情况：

```text
A: 入职申请被驳回，资料填写不正确，审核慢，很烦
B: 可能不能入职，计划被打乱，心烦
```

P7.1 只能看到：

```text
scene_overlap = 工作
profile_similarity = 0.408
keyword_overlap = 0
emotion_overlap = 0
pattern_tags_overlap = 0
score = 0.345
```

确定性规则不合并是合理的，否则会把大量“工作”内容误合并。但从产品语义看，两者可能都属于同一个长期主题：

```text
入职流程受阻 / 工作开始受阻 / 计划受阻
```

P7.2 因此引入受约束的 LLM 边界判断。LLM 不是主流程，也不是每次调用；它只在确定性算法无法处理的边界候选上判断是否存在稳定的长期主题。

LLM prompt 不使用某一组测试数据里的“行动受阻 / 计划被打断”作为核心模板。边界判断应回到产品语义：当前 Trace 和候选星座是否共享反复出现的处境、自我反应模式、关系位置、身份状态，或同一种内在需要与顾虑。

### 不采用自由 normalized 字段

P7.2 不新增自由文本式的 `normalized_topic` / `normalized_pattern_tags`。

原因：

```text
normalized_topic: 入职流程受阻
normalized_topic: 工作开始受阻
normalized_topic: 报到受阻
normalized_topic: 入职计划卡住
```

如果这些字段仍由 LLM 自由生成，本质上只是“更上位的 pattern_tags”，仍然存在同义不同词问题。

P7.2 改用 codebook 思路：让模型优先从已有稳定 code 中选择，而不是每次自由发明标签。

### Theme Codebook

Codebook 是一组可复用的主题代码。每个 code 至少包含：

```text
theme_code
theme_label
theme_description
theme_examples
```

示例：

```json
{
  "theme_code": "work_onboarding_blocked",
  "theme_label": "入职流程受阻",
  "theme_description": "入职、报到、申请、审核、资料、offer 等流程发生阻碍、延迟或不确定。",
  "theme_examples": [
    "入职申请被驳回了，说资料填写不正确，审核又慢",
    "可能不能入职，计划被打乱",
    "报到前资料又出问题"
  ]
}
```

字段定位：

- `theme_code`：稳定英文 snake_case 标识，算法和日志使用。
- `theme_label`：中文可读标签，便于调试和后续后台查看。
- `theme_description`：边界定义，帮助 LLM 判断哪些 Trace 可以复用这个 code。
- `theme_examples`：代表性样本，后续 P8 可持续更新。

### 最小数据结构

P7.2 先不引入独立 codebook 表，避免一次改动过大。

最小落点：

```text
constellation_profiles.theme_code
constellation_profiles.theme_label
constellation_profiles.theme_description
constellation_profiles.theme_examples jsonb
```

新建星座时：

```text
theme_code / theme_label / theme_description
默认由 TraceProfile.topic / summary / central_pattern 和代表 moments 生成 fallback codebook
如果边界 LLM 返回有效 suggest_new，则使用 suggested_theme_code / label / description 覆盖 fallback
```

加入已有星座时：

```text
复用已有 ConstellationProfile.theme_code
membership.match_reason 记录 LLM 判断理由
```

TraceProfile 暂不强制新增 theme 字段。若后续需要离线分析，可再讨论：

```text
trace_profiles.theme_code
trace_profiles.theme_confidence
```

### 触发条件

P7.2 不每次调用 LLM。调用前先走 P7.1 确定性评分：

```text
score >= strong_threshold:
  直接加入已有星座，不调用 LLM

candidate_count == 0:
  直接新建星座，不调用 LLM

score < weak_threshold 且没有边界证据:
  直接新建星座，不调用 LLM

weak_threshold <= score < strong_threshold:
  进入边界判断
```

建议初始参数：

```text
strong_threshold = 0.72
weak_threshold = 0.30
borderline_profile_similarity = 0.38
llm_top_k = 3
```

进入 LLM 的候选还需要至少满足一个边界证据：

```text
profile_similarity >= 0.38
or keyword_overlap > 0
or emotion_overlap > 0
or pattern_tags_overlap > 0
or scene_overlap >= 1 且该候选在 dense top3 内
```

这样可以避免只因为同属“工作”就调用 LLM。

每次 `StashTrace` 最多调用一次边界 LLM，且只传 top3 候选。

当前实现中，进入 LLM 的候选来自已召回并排序后的 top3 边界候选；每个候选都必须低于 middle 阈值且不低于 weak 阈值，并具备至少一类边界证据。

### LLM 输入

输入包含当前 TraceProfile 和 top3 候选星座的 codebook 信息：

```json
{
  "trace_profile": {
    "topic": "入职计划延迟",
    "summary": "用户因可能无法入职而计划被打乱，感到心烦。",
    "keywords": ["入职", "计划", "不确定"],
    "emotions": ["心烦"],
    "scenes": ["工作"],
    "central_pattern": "当计划遭遇不确定性时的情绪反应",
    "pattern_tags": ["计划受挫", "不确定性", "情绪波动"],
    "representative_moment": "最后跟我说可能不能入职，打乱计划，心烦"
  },
  "candidates": [
    {
      "constellation_id": "...",
      "topic": "入职申请被驳回",
      "summary": "用户因入职申请资料填写不正确多次被驳回，审核过程缓慢，感到烦躁。",
      "theme_code": "work_onboarding_blocked",
      "theme_label": "入职流程受阻",
      "theme_description": "入职、报到、申请、审核、资料、offer 等流程发生阻碍、延迟或不确定。",
      "keywords": ["入职申请", "驳回", "资料填写", "审核慢"],
      "emotions": ["烦躁"],
      "scenes": ["工作", "入职"],
      "pattern_tags": ["反复受阻", "流程拖延"],
      "score": 0.345
    }
  ]
}
```

### LLM 输出

结构化 JSON：

```json
{
  "decision": "use_existing",
  "constellation_id": "...",
  "theme_code": "work_onboarding_blocked",
  "confidence": 0.74,
  "shared_situation": "两者都围绕入职流程受阻，以及由流程不确定带来的计划打乱和烦躁。",
  "match_dimensions": ["situation", "self_pattern"],
  "reason": "虽然一个是资料被驳回，一个是可能无法入职，但共同处境都是入职流程被阻断。"
}
```

如果没有合适候选：

```json
{
  "decision": "suggest_new",
  "suggested_theme_code": "plan_disrupted_by_work_uncertainty",
  "suggested_theme_label": "工作不确定打乱计划",
  "suggested_theme_description": "工作相关安排发生不确定变化，导致原计划被迫调整并引发烦躁或不安。",
  "confidence": 0.68,
  "reason": "这更关注计划被打乱，而不是已有候选的申请审核流程本身。"
}
```

允许的 `decision`：

```text
use_existing
suggest_new
```

P7.2 不让 LLM 直接决定多个 secondary。Secondary 规则先保持 P7.1 的确定性策略；更完整多视角归属可后续单独设计。

### 接受门控

LLM 判断不能无条件覆盖确定性算法。

`use_existing` 必须满足：

```text
confidence >= 0.65
constellation_id 属于输入 candidates
theme_code 属于该 candidate 的 theme_code
shared_situation 非空
match_dimensions 至少包含 situation / self_pattern / relationship / identity / need_conflict 中的一类
```

`wording` 只能作为辅助证据，不作为独立接受门控。`process_blocked` / `plan` 不再作为 prompt 的核心维度，避免模型把一次性任务受阻或计划变化误判为长期主题。

否则拒绝 LLM 结果，回退到 P7.1 确定性决策。

`suggest_new` 必须满足：

```text
suggested_theme_code 非空
suggested_theme_label 非空
suggested_theme_description 非空
confidence >= 0.55
```

如果 LLM 失败、JSON 解析失败、返回未知 `constellation_id`、低 confidence，均回退 P7.1。

### 硬约束

Prompt 必须强调：

```text
不能只因为同属“工作”就合并。
不能只因为都有负面情绪就合并。
不能只因为关键词相同就合并。
必须能说出共同的长期主题。
如果共同长期主题说不清楚，返回 suggest_new。
优先避免污染已有星座。
只能选择输入 candidates 里的 constellation_id。
不要只因为一次性任务受阻、计划变化或具体事件相似就合并。
```

### 评分与落库

P7.2 不改变 P7.1 的确定性 score 公式。

如果 LLM `use_existing` 被接受：

```text
membership.match_score = deterministic_score
membership.match_reason = LLM reason
membership.match_dimensions = LLM match_dimensions
```

日志中同时记录：

```text
deterministic_score
llm_confidence
shared_situation
accepted
```

是否新增数据库字段记录 `llm_confidence` 暂缓；P7.2 先通过日志观察。

当前实现落点：

```text
domain:
  ConstellationProfile 增加 theme codebook 字段
  ConstellationBorderlineJudge 端口与结构化输入/输出

adapter/ai:
  ConstellationBorderlineJudge 调用 Chat，并解析严格 JSON

adapter/postgres:
  constellation_profiles 读写 theme_code / theme_label / theme_description / theme_examples
  migration 014_constellation_theme_codebook.sql

app:
  P7.1 deterministic score 先执行
  middle / explainable middle 直接加入
  borderline top3 调用 LLM
  use_existing 通过 gate 后写 primary membership
  低置信度、未知候选、theme_code 不匹配、缺少 shared_situation 时创建新星座
```

### 日志要求

新增日志：

```text
starmap borderline judgement started
starmap borderline judgement completed
starmap borderline judgement rejected
```

记录：

```text
trace_id
star_id
top_score
candidate_count
candidate_ids
candidate_theme_codes
llm_decision
llm_confidence
shared_situation
match_dimensions
accepted
fallback_reason
```

### 测试样本

应覆盖：

1. 长期主题相同，确定性分数低，但 LLM 选择已有 code：

```text
已有：入职申请被驳回，审核慢，很烦
新增：可能不能入职，计划被打乱，心烦
期望：use_existing -> 入职流程受阻
```

2. 只有泛场景相同，不能合并：

```text
已有：入职申请被驳回
新增：今天工作会议很多
期望：suggest_new / create_new
```

3. 只有情绪相似，不能合并：

```text
已有：入职申请被驳回，很烦
新增：朋友没回消息，很烦
期望：suggest_new / create_new
```

4. LLM 返回未知 constellation_id，拒绝。

5. LLM confidence 过低，拒绝。

6. 强匹配候选不调用 LLM，直接加入。

7. 无候选不调用 LLM，直接新建。

### P7.3 迭代：统一归属裁判

状态：已实现。

P7.3-a/b 将 P7.2 的边界 LLM 从“只兜底 primary”升级为“统一归属裁判”：

```text
确定性评分
  -> score >= strong_threshold 直接 primary
  -> 0.30 <= score < strong_threshold 进入 LLM
  -> LLM 一次判断 primary + secondary
```

当前实现参数：

```text
strong_threshold = 0.68
borderline_weak_threshold = 0.30
llm_top_k = 3
primary confidence >= 0.65
secondary confidence >= 0.60
```

LLM 输出从单个 `constellation_id` 扩展为：

```json
{
  "decision": "use_existing",
  "primary": {
    "constellation_id": "...",
    "theme_code": "...",
    "confidence": 0.82,
    "shared_theme": "入职资料、审核和反馈反复卡住",
    "match_dimensions": ["situation"],
    "reason": "这是用户回看时最自然的主星座"
  },
  "secondary": [
    {
      "constellation_id": "...",
      "theme_code": "...",
      "confidence": 0.72,
      "shared_theme": "也可以从被审核的位置感理解",
      "match_dimensions": ["identity"],
      "reason": "这是另一个合理视角"
    }
  ]
}
```

落库规则：

- primary 通过 LLM gate 后写入 `match_type=primary`、`weight=1.0`。
- secondary 最多 2 个，写入 `match_type=secondary`、`weight=0.5`。
- secondary 不再要求 `score >= middle_threshold`，但必须通过 confidence、candidate id、theme_code、shared_theme、match_dimensions gate。
- 新建 primary 时，只接受 LLM 明确给出的 secondary；不会再用确定性分数自动挂旧候选。

确定性分数现在主要承担：

- 强匹配直通。
- 候选排序。
- 触发 LLM 的边界区判断。
- 诊断日志和后续调参。

### P7.4 预留：ConstellationProfile ES sparse 召回

星座级 ES sparse 召回暂不放入 P7.2。

原因：当前暴露的问题不是候选没召回，而是候选召回后确定性分数过低。ES sparse 更适合后续解决候选遗漏。

预留 P7.4：

```text
dense: pgvector profile_embedding topK
sparse: ES ConstellationProfile topK
fused: RRF(dense, sparse)
score: P7.1 综合评分
borderline: P7.2 codebook LLM
```

## P8 衔接

P7.1 引入 `pattern_tags`，P7.2 引入 theme codebook，但画像合并仍是轻量版本。P8 需要进一步解决：

- `pattern_tags`、`keywords`、`emotions`、`scenes` 的频次与权重统计。
- `theme_examples` 的代表性更新。
- `theme_description` 的异步重写与收敛。
- 低辨识度标签过滤。
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
