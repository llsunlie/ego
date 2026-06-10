# Echo 与星座匹配聚合优化总体任务表

## 文档定位

本文是后续优化工作的总体任务表，用于拆分 Echo 匹配与星座聚合优化的阶段性工作。

本文不包含具体算法设计、字段设计、prompt 设计、迁移 SQL 设计或前端交互细节。每个任务进入实施前，需要另行补充对应的设计文档或模块级 `.harness` 任务说明。

## 已确认决策

1. Echo 匹配不引入 insight embedding。
   - 原因：Insight 不一定准确，不适合作为 Echo 召回的核心依据。
   - Echo 优化仍围绕 Moment 原文、历史距离、Trace 关系、候选质量与重排判断展开。

2. Echo 候选召回使用 pgvector / HNSW。
   - 目标：把 topK 召回交给数据库，提高检索效率。
   - 应避免继续依赖应用层全量扫描历史 Moment 作为长期方案。

3. Echo sparse search 使用 Elasticsearch。
   - 目标：补足 dense embedding 对具体短语、反复措辞和词面呼应不敏感的问题。
   - 方案方向：pgvector dense recall + Elasticsearch sparse recall + RRF 融合。
   - 中文分词优先评估 IK analyzer，并以 char ngram 字段作为兜底。

4. TraceProfile 需要持久化。
   - TraceProfile 是星座聚合的算法中间层，不等同于用户可见的 MomentInsight。

5. 星座聚合允许较大改造。
   - 当前是新版本阶段，可以对 Constellation 数据结构与匹配流程做替换兼容。

## 总体阶段

| 阶段 | 任务 | 目标 | 产出 | 状态 |
|---|---|---|---|---|
| P0 | 建立评估样本与验收口径 | 让后续优化有可比较的基准 | Echo/星座正负样本、人工验收标准 | 已完成 |
| P1 | Echo 召回效率优化 | 用数据库 topK 替代应用层全量扫描 | pgvector/HNSW 召回能力、候选 topK 接口 | 已完成 |
| P2 | Echo 匹配质量优化 | 先用规则降低重复和同 Trace 干扰 | 候选过滤、去重、时间/Trace 规则、内部 echo_score | 已完成 |
| P2.5 | Elasticsearch sparse search | 引入词面/短语召回，并与 dense 召回融合 | ES index、中文 analyzer、回填、RRF hybrid recall | 已完成 |
| P3 | Echo 结果语义与前端兼容 | 保持 proto 基本兼容，同时让 similarity/score 含义清晰 | Echo score 口径文档、前端展示兼容说明 | 已完成 |
| P4 | TraceProfile 旁路持久化 | 为星座聚合提供稳定的 Trace 级画像 | TraceProfile 异步生成、存储、profile embedding | 已完成 |
| P5 | TraceProfile 质量验证与调优 | 确认 TraceProfile 稳定、具体、不过度推断 | 质量样本、复核流程、生成器回归测试 | 已建立基线 |
| P6 | ConstellationProfile 改造 | 让星座从短 topic 聚合升级为长期主题画像聚合 | 星座画像结构、兼容旧展示字段、Star 多对多归属模型 | 已完成设计 |
| P7 | 星座匹配流程升级 | 用 TraceProfile 匹配 ConstellationProfile | 候选召回、评分、primary/secondary 多归属流程 | 已完成第一版 |
| P7.1 | 星座匹配过度拆分修正 | 用 pattern_tags 与更合理的评分/阈值减少误新建 | pattern_tags 字段、评分公式、解释性 middle 规则、核心日志 | 已完成 |
| P7.2 | 星座 Codebook 与边界判断 | 处理同一长期主题但确定性分数过低的候选 | theme codebook、边界 LLM、门控规则、诊断日志 | 已完成 |
| P7.3 | 星座统一归属裁判 | 让边界 LLM 一次判断 primary + secondary，并降低确定性分数过低导致的兜底比例 | LLM primary/secondary 输出、阈值调整、secondary gate、回归测试 | 已完成 |
| P7.4 | 星座 sparse 召回 | 降低 ConstellationProfile 候选遗漏 | ES sparse index、dense/sparse RRF 融合 | 待设计 |
| P8 | 星座画像合并与一致性优化 | 避免画像简单合并后变长、变泛，并补齐异步一致性 | 标签权重/topN、异步画像刷新、消息队列/补偿任务设计 | 已标记待设计 |
| P9 | lonely / forming / active 状态落地 | 表达星座从孤星到稳定主题的形成过程 | 状态规则、列表/详情返回兼容 | 待讨论 |
| P10 | 观测、回归与调参 | 可观察优化效果并持续校准 | 日志、指标、离线评估脚本、回归用例 | 待设计 |

## 任务分解

### P0. 建立评估样本与验收口径

目标：
- 在优化前确定“什么是更好的 Echo”和“什么是更好的星座聚合”。
- 避免后续只凭感觉调阈值或改 prompt。

任务：
- 收集 Echo 正例与负例。
- 收集星座聚合正例与负例。
- 定义 Echo 误召回、漏召回、重复召回的验收口径。
- 定义星座误合并、过度拆分、孤星保留的验收口径。
- 建立一份可反复运行或人工复核的评估清单。

### P1. Echo 召回效率优化

目标：
- 将历史 Moment 候选召回从应用层全量扫描迁移到数据库 topK。
- 为后续质量优化提供稳定候选集。

任务：
- 确认现有 Moment embedding 的存储形态与 pgvector 目标形态。
- 设计 pgvector/HNSW 迁移路径。
- 增加 topK 候选查询能力。
- 将 EchoMatcher 的候选来源切换为数据库召回结果。
- 保留必要的兼容或回退路径。

### P2. Echo 匹配质量优化

目标：
- 在候选 topK 基础上，先用确定性规则提高 Echo 的“回声感”，减少同 Trace 内容和连续重复候选。

任务：
- 排除当前 Trace 内的候选。
- 增加时间距离轻量加权。
- 同 Trace 候选只保留内部分最高的一条。
- 最多返回 3 条 Echo matched moments。
- 不改 proto，`similarities` 字段直接承载最终 `echo_score`。
- 更新 Echo 相关测试与评估样本。

### P2.5. Elasticsearch sparse search

目标：
- 引入 ES sparse recall，补足 pgvector dense recall 对具体短语、反复措辞和词面呼应不敏感的问题。

任务：
- 增加 Elasticsearch 本地服务与中文 analyzer 方案。
- 设计 Moment search index mapping。
- 将 Moment 写入同步到 ES search index。
- 增加历史 Moment 回填到 ES 的独立命令。
- 增加 ES sparse topK recall。
- 使用 RRF 融合 dense rank 与 sparse rank。
- 将融合候选接入 P2 EchoRanker。

当前结果：
- `CreateMoment` 同步 best-effort 写入 ES，写入失败只记录日志，不阻断创建。
- Echo 候选由 pgvector dense topK 与 ES sparse topK 并发召回，通过 RRF 融合后进入 P2 EchoRanker。
- pgvector dense 查询失败会直接返回错误；ES sparse 查询或回读失败只记录 warn，并继续使用 dense 候选。
- 历史 Moment 可通过 `server/cmd/backfill-moment-search` 回填到 ES。
- 本地 `docker-compose.yml` 已加入 Elasticsearch + IK analyzer 自定义镜像。
- 日志保留 `echo recall candidates`、`echo match candidate scores` 与 `echo final matches` 核心诊断信息，用于查看 dense、ES、RRF 融合、EchoMatcher 分数计算和最终 Echo 命中的具体候选；中间碎片日志与 gRPC 完整 req/res 日志已降噪。

### P3. Echo 结果语义与前端兼容

目标：
- 在不改或少改 proto 的前提下，让 Echo 返回结果对前端保持稳定。

任务：
- 已确认不修改 proto，`similarities` 字段升级为最终 `echo_score`，前端沿用原字段展示。
- 更新契约说明与前端展示说明。
- 讨论是否需要内部强/弱/无 Echo 分档。
- 若分档不进入 proto，则明确其只作为后端过滤策略。

### P4. TraceProfile 持久化

目标：
- 为星座聚合提供 Trace 级算法画像。

任务：
- 在 `StashTrace` 后台旁路异步生成 TraceProfile，不阻塞当前返回。
- 当前星座 topic 聚合照旧，不在 P4 替换。
- 设计并创建 `trace_profiles` 与 `trace_profile_vectors`。
- 生成 `topic`、`summary`、`keywords`、`emotions`、`scenes`、`central_pattern`、`representative_moment_index` 与 `profile_text`，后端将 index 映射为持久化的 `representative_moment_id`。
- 对 `profile_text` 生成 embedding 并写入独立 pgvector 表。
- LLM JSON 生成在 generator 内最多 3 次尝试，仍失败则 fallback；embedding 在 generator 内最多 3 次尝试，仍失败则写入 `failed` profile 但不写 vector。

### P5. TraceProfile 质量验证与调优

目标：
- 在替换星座聚合前，确认 TraceProfile 输出可用。
- 避免后续 ConstellationProfile 匹配被过泛 topic、过度推断或错误代表 moment 污染。

任务：
- 建立 TraceProfile 固定质量样本。
- 明确 topic、summary、keywords、emotions、scenes、central_pattern、representative_moment_index / representative_moment_id 的人工复核口径。
- 增加生成器纯函数回归测试。
- 通过日志观察真实 TraceProfile 输出，继续调整 prompt 和规范化规则。
- 保持当前 topic-based constellation clustering 不变。

### P6. ConstellationProfile 改造

目标：
- 将星座从短 topic 驱动改为长期主题画像驱动。
- 保持 proto / 前端返回兼容，同时为后续多视角归属提供内部数据结构。

任务：
- 保留 `constellations` 作为星座主表和 proto 兼容输出来源。
- 设计 `constellation_profiles` 作为星座长期主题画像。
- 设计 `constellation_profile_vectors` 保存 profile embedding 与 centroid embedding。
- 设计 `constellation_stars` 表达 Star 与 Constellation 的多对多归属。
- 明确 `Trace -> Star` 是一对一，`Star <-> Constellation` 是多对多。
- 明确画像命名保留 `TraceProfile`，不改成 `StarProfile`。
- 不做历史星座迁移；正式启用前会清空数据库。
- 不在 P6 替换当前 topic-based matching。

### P7. 星座匹配流程升级

目标：
- 用 TraceProfile 匹配 ConstellationProfile，替代当前 topic embedding 直接匹配。
- 支持一个 Star 从多个视角加入多个星座。

任务：
- 新增 ConstellationProfile 持久化与向量持久化。
- 新增 `constellation_stars` 表达 Star 与 Constellation 的多对多归属。
- `StashTrace` 后台聚类切换为 TraceProfile -> ConstellationProfile。
- 移除旧 `topic -> ConstellationMatcher` 聚合路径，避免 profile 聚合失败时静默回到旧逻辑。
- TraceProfile 的 LLM JSON 与 embedding 重试收敛在 generator 内；关键异步错误记录 `recovery=pending_message_queue`，后续用消息队列或补偿任务保证一致性。
- 召回 ConstellationProfile topK 候选并综合评分。
- 支持 primary / secondary 归属。
- 新建星座时以 TraceProfile 初始化 ConstellationProfile。
- 加入已有星座时更新 membership、profile 统计和 centroid embedding。
- `ListConstellations.TotalStarCount` 按唯一 Star 计数，兼容多归属。
- 暂不扩展 proto，不展示 secondary membership。

### P7.1. 星座匹配过度拆分修正

目标：
- 修正 P7 第一版在短中文 Trace、单 Trace 星座和强结构化重合样本上偏向新建星座的问题。
- 在不引入 LLM rerank 的情况下，先提升确定性评分的稳定性和可解释性。

任务：
- `TraceProfile` 与 `ConstellationProfile` 增加 `pattern_tags`，用于表达经历方式、处境结构或反复模式。
- `central_pattern` 保持人可读描述，不再直接作为主要 overlap 信号。
- 用 `pattern_tags_overlap` 替换当前 `central_pattern_overlap`。
- 调整评分公式，提高 keywords / scenes / emotions / pattern_tags 等结构化证据权重。
- 对单 Trace 星座避免 `profile_similarity` 与 `centroid_similarity` 重复计权。
- 将目标阈值调整为 `strong=0.72`、`middle=0.60`。
- 增加解释性 middle 规则：当 score 接近 middle 且结构化证据至少 3 类命中时，不直接新建星座。
- 补充候选级和最终决策级 debug 日志，输出各维度分数、命中项、阈值差距和决策原因。

### P7.2. 星座 Codebook 与边界判断

目标：
- 处理“同一长期主题，但确定性 overlap 分数过低”的候选。
- 避免自由生成 normalized 字段导致新的同义不同词问题。
- 通过 theme codebook 让 LLM 在已有稳定主题 code 中选择，而不是每次自由发明标签。
- 让 LLM 只作为边界裁判，不进入每次聚合主流程。

任务：
- 已在 `ConstellationProfile` 增加 `theme_code`、`theme_label`、`theme_description`、`theme_examples`，并通过 `014_constellation_theme_codebook.sql` 迁移落库。
- 新建星座时会从 TraceProfile/moments 生成 fallback theme codebook；当 LLM 返回有效 `suggest_new` 时可覆盖新主题 codebook。
- P7.1 确定性评分先执行；达到 middle/解释性 middle 的候选直接加入，无候选或无边界证据直接新建。
- 仅在边界区间触发 LLM：`weak_threshold=0.30`、`strong_threshold=0.72`、`llm_top_k=3`。
- 边界候选需要满足至少一个证据：`profile_similarity >= 0.38`，或关键词/情绪/场景/pattern_tags 任一命中。
- LLM 输入只包含当前 TraceProfile 与 top3 边界候选的 theme codebook 和确定性分数明细。
- LLM 输出 `use_existing` 或 `suggest_new`，并返回 confidence、shared_situation、match_dimensions、reason。
- `use_existing` 必须通过门控：confidence 足够、constellation_id 属于候选、theme_code 匹配候选、shared_situation 非空、match_dimensions 包含允许的上层维度。
- LLM 失败、低置信度、未知 constellation_id、JSON 解析失败时回退 P7.1 确定性结果。
- 日志输出候选 theme code、确定性分数明细、LLM confidence、shared_situation、accepted / rejected 与 fallback_reason。
- P7.2 不扩展 proto，不强制新增独立 codebook 表，不把 ES sparse 召回纳入同一阶段。

### P7.3. 星座统一归属裁判

目标：
- 将边界 LLM 从“primary 兜底”升级为一次判断 primary + secondary。
- 让确定性分数负责强匹配直通、候选排序和 LLM 触发，而不是独自决定大多数自然主题归属。
- 降低 secondary 对 `middle_threshold` 的依赖，避免主归属靠 LLM、secondary 却全部因分数低被跳过。

已完成：
- `ConstellationBorderlineJudgement` 增加 `Primary` 与 `Secondary` 选择结构，同时兼容旧的单候选 JSON。
- `ConstellationBorderlineJudge` prompt 改为“用户回看时最自然的星座归属”口径，输出 primary + secondary。
- strong threshold 调整为 `0.68`；borderline 候选区间改为 `0.30 <= score < 0.68`。
- score 权重轻调：提高 scene/keyword 对短中文自然主题的辅助作用，降低 emotion 权重。
- LLM primary gate 维持 `confidence >= 0.65`；secondary gate 使用 `confidence >= 0.60`，不再要求 `score >= 0.60`。
- 新建 primary 后仅接受 LLM 明确给出的 secondary，不再用确定性分数自动挂旧候选。
- 增加 primary + secondary 落库回归测试，以及新旧 JSON parser 测试。

### P7.4. 星座 sparse 召回

目标：
- 在 P7.2/P7.3 解决“候选召回到了但确定性分数低、LLM 归属未整合多视角”的问题之后，再处理“候选没有召回到”的问题。

任务：
- 新增 ConstellationProfile Elasticsearch index。
- 索引 topic、summary、keywords、emotions、scenes、pattern_tags、central_pattern、theme_label、theme_description。
- 使用 TraceProfile 构造 sparse 查询文本。
- 将 pgvector dense topK 与 ES sparse topK 通过 RRF 融合。
- 融合候选进入 P7.1 综合评分，再进入 P7.2 codebook 边界判断。
- ES 只作为候选召回，不直接决定星座归属。

### P8. 星座画像合并与一致性优化

目标：
- 避免 `ConstellationProfile` 在持续吸收 Trace 后变长、变泛、失去辨识度。
- 为异步聚类失败提供可恢复的一致性机制。

任务：
- 设计 keywords / emotions / scenes 的频次、权重、衰减和 topN 规则。
- 设计过泛词过滤和代表性标签保留规则。
- 设计加入 Trace 后的异步 LLM 画像刷新。
- 设计代表性 Moment 的选择与更新规则。
- 设计消息队列或补偿任务，保证 TraceProfile、membership、ConstellationProfile 更新最终一致。

### P9. lonely / forming / active 状态落地

目标：
- 让算法状态表达“孤星等待同伴”“主题正在形成”“主题已经稳定”的过程。

任务：
- 讨论状态是否进入数据库字段。
- 讨论状态变化规则。
- 讨论前端是否需要感知该状态，或仅后端兼容为现有 constellation 输出。
- 讨论 lonely star 是否参与后续匹配与合并。

### P10. 观测、回归与调参

目标：
- 让匹配优化可观察、可回归、可持续调参。

任务：
- 增加 Echo 召回数量、过滤数量、最终返回数量等日志或指标。
- Echo 召回日志应优先记录 dense / ES / fused / final matches 的候选摘要，包括 `moment_id`、`trace_id`、`created_at`、短 `content_preview` 与最终 `similarity`；避免只记录“某步骤完成”的低价值流水日志。
- gRPC composite 层不记录完整 req/res，避免重复输出用户原文；业务层按场景记录可用于判断算法效果的受限 preview。
- 增加星座匹配候选数、匹配分、状态变化等日志或指标。
- 建立离线评估入口。
- 建立关键样本的回归测试或人工复核流程。

## 当前不做的事

- 不引入 insight embedding 参与 Echo 匹配。
- 不在本计划中固定 Echo 综合评分权重。
- 不在本计划中固定星座匹配评分公式。
- 不在本计划中固定 LLM rerank prompt。
- 不在本计划中直接定义数据库迁移细节。
- 不在本计划中直接修改 proto。
- 不使用手写关键词表替代 sparse search。

## 后续文档建议

后续每个阶段进入实施前，建议补充对应设计文档：

- `echo-retrieval-design.md`
- `echo-ranking-design.md`
- `echo-sparse-search-design.md`
- `trace-profile-design.md`
- `trace-profile-quality-plan.md`
- `constellation-profile-design.md`
- `constellation-matching-design.md`
- `constellation-state-design.md`
- `matching-evaluation-plan.md`
