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
| P3 | Echo 结果语义与前端兼容 | 保持 proto 基本兼容，同时让 similarity/score 含义清晰 | Echo score 口径文档、前端展示兼容说明 | 待设计 |
| P4 | TraceProfile 持久化 | 为星座聚合提供稳定的 Trace 级画像 | TraceProfile 生成、存储、回填任务 | 待设计 |
| P5 | ConstellationProfile 改造 | 让星座从短 topic 聚合升级为长期主题画像聚合 | 星座画像结构、兼容旧展示字段 | 待设计 |
| P6 | 星座匹配流程升级 | 用 TraceProfile 匹配 ConstellationProfile | 候选召回、评分、加入/新建/暂存流程 | 待设计 |
| P7 | lonely / forming / active 状态落地 | 表达星座从孤星到稳定主题的形成过程 | 状态规则、列表/详情返回兼容 | 待讨论 |
| P8 | 星座画像持续演化 | 避免星座被第一次 topic 固化 | 加入 Trace 后的同步统计更新与异步画像更新 | 待设计 |
| P9 | 观测、回归与调参 | 可观察优化效果并持续校准 | 日志、指标、离线评估脚本、回归用例 | 待设计 |

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
- 讨论 TraceProfile 的生成时机。
- 讨论 TraceProfile 的输入材料范围。
- 设计 TraceProfile 持久化结构。
- 增加 TraceProfile 生成与读取能力。
- 处理历史 Trace 的回填策略。

### P5. ConstellationProfile 改造

目标：
- 将星座从短 topic 驱动改为长期主题画像驱动。

任务：
- 讨论 ConstellationProfile 字段范围。
- 讨论展示字段与算法字段的兼容关系。
- 设计 profile embedding 与 centroid embedding 的职责边界。
- 设计从现有 Constellation 数据迁移到新画像结构的路径。

### P6. 星座匹配流程升级

目标：
- 用 TraceProfile 匹配 ConstellationProfile，替代当前 topic embedding 直接匹配。

任务：
- 讨论候选星座召回方式。
- 讨论匹配评分信号，但不在本任务表中固定具体公式。
- 讨论 LLM judgement 是否进入最终匹配流程。
- 设计加入已有星座、新建星座、保留孤星的决策边界。
- 更新 StashTrace、ListConstellations、GetConstellation 相关测试。

### P7. lonely / forming / active 状态落地

目标：
- 让算法状态表达“孤星等待同伴”“主题正在形成”“主题已经稳定”的过程。

任务：
- 讨论状态是否进入数据库字段。
- 讨论状态变化规则。
- 讨论前端是否需要感知该状态，或仅后端兼容为现有 constellation 输出。
- 讨论 lonely star 是否参与后续匹配与合并。

### P8. 星座画像持续演化

目标：
- 避免星座长期被第一次生成的 topic 或画像锁死。

任务：
- 讨论 Trace 加入星座后的同步更新内容。
- 讨论异步画像更新任务。
- 讨论更新失败、重试、幂等与补偿机制。
- 讨论代表性 Moment 的选择与更新规则。

### P9. 观测、回归与调参

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
- `constellation-profile-design.md`
- `constellation-state-design.md`
- `matching-evaluation-plan.md`
