# Echo 与星座匹配评估基线

## 文档定位

本文是 P0「建立评估样本与验收口径」的产出，用于后续 Echo 匹配与星座聚合优化的共同评估基线。

本文只定义评估对象、样本格式、首批合成样本、验收口径与人工复核流程，不定义具体召回算法、打分公式、数据库迁移、prompt 或实现方案。

## 评估目标

### Echo 评估目标

Echo 的评估问题是：

> 当前 Moment 是否召回了真正像“过去的自己在回应我”的历史 Moment？

高质量 Echo 应优先满足：

- 情绪或处境有呼应。
- 核心困扰、自我模式或表达方式有呼应。
- 不只是表层词相同。
- 不来自当前 Trace 内刚写过的内容。
- 不连续返回多条几乎相同的候选。

### 星座评估目标

星座聚合的评估问题是：

> 当前 Trace 是否被沉淀到正确的长期主题中，或被合理保留为孤星/形成中主题？

高质量星座聚合应优先满足：

- 不因短 topic 相似而误合并。
- 不因 topic 命名差异而过度拆分。
- 能识别长期反复出现的主题画像。
- 能允许孤星存在，而不是硬凑星座。
- 星座画像能随新 Trace 加入持续变准确。

## 样本集结构

后续样本可以以 Markdown 表格维护，也可以转成 JSON/CSV。建议每条样本至少包含以下字段。

### Echo 样本字段

| 字段 | 说明 |
|---|---|
| `case_id` | 样本 ID |
| `current_moment` | 当前 Moment 原文 |
| `current_trace_context` | 当前 Trace 内已有上下文，可为空 |
| `history_candidates` | 历史候选 Moment 列表 |
| `expected_positive_ids` | 应当召回的历史 Moment ID |
| `expected_negative_ids` | 不应召回的历史 Moment ID |
| `must_exclude_ids` | 必须排除的候选，如当前 Trace 内内容 |
| `reason` | 人工标注理由 |
| `tags` | 样本类型标签 |

### 星座样本字段

| 字段 | 说明 |
|---|---|
| `case_id` | 样本 ID |
| `incoming_trace` | 当前要收进星图的 Trace 摘要或原文列表 |
| `existing_constellations` | 已有星座/孤星/形成中主题 |
| `expected_action` | `join_existing` / `create_lonely` / `create_forming` / `promote_active` 等 |
| `expected_target_id` | 预期加入的星座 ID，可为空 |
| `expected_negative_ids` | 不应加入的星座 ID |
| `reason` | 人工标注理由 |
| `tags` | 样本类型标签 |

## Echo 首批评估样本

### E01 深层处境相似，表层词不同

| 字段 | 内容 |
|---|---|
| case_id | `echo.e01.deep_situation` |
| current_moment | `我不想再主动找他了，可又会一直等他有没有发现我不高兴。` |
| history h1 | `每次都是我先开口，我真的有点累。` |
| history h2 | `今天工作太累了，什么都不想做。` |
| history h3 | `我只是想让他先问问我怎么了。` |
| expected_positive_ids | `h1`, `h3` |
| expected_negative_ids | `h2` |
| reason | h1/h3 都在表达亲密关系里主动、等待、被看见的处境；h2 只有“累”的表层相似。 |
| tags | `relationship`, `surface_word_trap`, `deep_echo` |

### E02 表层关键词相同，但不是回声

| 字段 | 内容 |
|---|---|
| case_id | `echo.e02.surface_word_false_positive` |
| current_moment | `今天工作好累，感觉所有事都压在我身上。` |
| history h1 | `我今天跑步好累，但还挺开心的。` |
| history h2 | `又是我一个人把所有事情扛完，好像没人会来接住我。` |
| expected_positive_ids | `h2` |
| expected_negative_ids | `h1` |
| reason | h1 只有“累”的词面相同；h2 在“独自承担”这一处境上有回声。 |
| tags | `work`, `surface_word_trap`, `responsibility` |

### E03 当前 Trace 内内容必须排除

| 字段 | 内容 |
|---|---|
| case_id | `echo.e03.exclude_current_trace` |
| current_trace_context | `t_current: ["我不想先开口。", "可我又一直在等。"]` |
| current_moment | `其实我就是想让他主动来找我。` |
| history h1 | `trace=t_current: 可我又一直在等。` |
| history h2 | `trace=t_old: 每次关系里都是我先说，我有点撑不住。` |
| expected_positive_ids | `h2` |
| must_exclude_ids | `h1` |
| reason | 当前 Trace 内刚写过的内容不能作为“过去的自己”的 Echo。 |
| tags | `current_trace_exclusion`, `relationship` |

### E04 时间过近需要降权

| 字段 | 内容 |
|---|---|
| case_id | `echo.e04_too_recent` |
| current_moment | `我又开始怀疑自己是不是不够好。` |
| history h1 | `5分钟前: 我是不是哪里做得不够好。` |
| history h2 | `45天前: 每次别人冷一点，我就会开始怀疑是不是自己不够好。` |
| expected_positive_ids | `h2` |
| expected_negative_ids | `h1` |
| reason | h1 过近，容易像重复；h2 更有“过去模式浮现”的回声感。 |
| tags | `time_distance`, `self_doubt` |

### E05 不应硬凑 Echo

| 字段 | 内容 |
|---|---|
| case_id | `echo.e05_no_echo` |
| current_moment | `今天第一次试着把阳台上的植物重新换盆。` |
| history h1 | `我昨晚看了一部电影。` |
| history h2 | `最近总是在想工作要不要换方向。` |
| expected_positive_ids | 空 |
| expected_negative_ids | `h1`, `h2` |
| reason | 没有足够呼应时应返回无 Echo，而不是硬找相似内容。 |
| tags | `no_echo`, `daily_life` |

## 星座首批评估样本

### C01 不同 topic 命名，但应聚为同一主题

| 字段 | 内容 |
|---|---|
| case_id | `constellation.c01.same_theme_different_topic` |
| incoming_trace | `["我不想每次都是我先开口。", "但我又一直在等他发现。"]` |
| existing c1 | `topic=关系里的沉默; summary=在亲密关系中等待对方主动靠近，同时不愿反复先表达需求。` |
| existing c2 | `topic=工作压力; summary=承担任务过多导致疲惫。` |
| expected_action | `join_existing` |
| expected_target_id | `c1` |
| expected_negative_ids | `c2` |
| reason | 虽然 topic 词面不同，但 Trace 与 c1 的长期主题画像一致。 |
| tags | `profile_match`, `topic_variation`, `relationship` |

### C02 topic 太泛，不应误合并

| 字段 | 内容 |
|---|---|
| case_id | `constellation.c02_generic_topic_false_merge` |
| incoming_trace | `["我跟朋友聊天时总怕自己说错话。"]` |
| existing c1 | `topic=关系沟通; summary=亲密关系中等待对方主动、害怕表达需求。` |
| existing c2 | `topic=表达焦虑; summary=在人际交流中担心说错话、被误解或显得不合适。` |
| expected_action | `join_existing` |
| expected_target_id | `c2` |
| expected_negative_ids | `c1` |
| reason | “关系沟通”过泛，不能只因 topic 相近误合并到 c1。 |
| tags | `generic_topic_trap`, `social_anxiety` |

### C03 单次新主题应保留为孤星

| 字段 | 内容 |
|---|---|
| case_id | `constellation.c03_lonely_new_theme` |
| incoming_trace | `["今天第一次认真做饭，发现切菜的时候心里很安静。"]` |
| existing c1 | `topic=工作压力; summary=长期任务压力和独自承担。` |
| existing c2 | `topic=亲密关系等待; summary=关系里等待对方主动靠近。` |
| expected_action | `create_lonely` |
| expected_target_id | 空 |
| reason | 新 Trace 没有足够相近的长期主题，不应硬凑进已有星座。 |
| tags | `lonely`, `new_theme`, `daily_life` |

### C04 两个孤星相似，可形成 forming

| 字段 | 内容 |
|---|---|
| case_id | `constellation.c04_lonely_to_forming` |
| incoming_trace | `["我又拖到最后才开始，越拖越怕。"]` |
| existing lonely l1 | `summary=一直拖延开始，直到压力变大才行动。` |
| existing c1 | `summary=亲密关系里的主动与等待。` |
| expected_action | `create_forming` |
| expected_target_id | `l1` |
| reason | incoming trace 与孤星 l1 属于同一反复模式，可形成 forming 主题。 |
| tags | `lonely_merge`, `forming`, `procrastination` |

### C05 已成型星座应随新 Trace 演化

| 字段 | 内容 |
|---|---|
| case_id | `constellation.c05_profile_evolution` |
| incoming_trace | `["我不是不想说，只是每次说出口都像是在求别人理解我。"]` |
| existing c1 | `status=active; topic=亲密关系中的主动与等待; summary=等待对方主动靠近，不想总由自己表达需求。` |
| expected_action | `join_existing` |
| expected_target_id | `c1` |
| reason | 新 Trace 应加入 c1，并推动星座画像吸收“表达需求像是在请求理解”这一更具体的侧面。 |
| tags | `active_update`, `profile_evolution`, `relationship` |

## Echo 验收口径

### 必须满足

- 不返回当前 Trace 内的 Moment 作为 Echo。
- 不因单个表层词相同而召回明显无关内容。
- 在无足够呼应时允许 Echo 为空。
- 同一历史 Trace 中不应连续返回多条高度重复候选，除非后续设计明确允许。
- 返回结果应有稳定排序口径，避免同样输入多次返回明显不同的 Echo。

### 优先优化

- 更偏向“处境/情绪/自我模式”的呼应，而不是普通话题相似。
- 时间距离要服务“过去的自己”的体验，而不是只按最近优先。
- Echo score 或 similarity 的含义必须可解释。

### 失败类型

| 类型 | 说明 |
|---|---|
| `false_positive_surface` | 因表层词相同召回错误内容 |
| `false_positive_current_trace` | 召回当前 Trace 内内容 |
| `false_positive_too_recent` | 召回过近内容导致像重复 |
| `false_negative_deep_echo` | 漏掉深层处境相似内容 |
| `duplicate_echoes` | 多条 Echo 来自同一连续片段，缺少代表性 |
| `forced_echo` | 无合适 Echo 时硬返回 |

## 星座验收口径

### 必须满足

- 不只根据短 topic 或展示名决定聚合。
- 能允许新主题以 lonely 形式存在。
- 不把明显不同的长期主题合并。
- 不因 topic 命名差异拆散同一长期主题。
- 星座加入新 Trace 后，应具备画像更新路径。

### 优先优化

- forming 状态应能吸收相似 lonely trace。
- active 星座应更稳定，不被弱相关 Trace 轻易污染。
- 星座画像字段应区分展示用途与匹配用途。

### 失败类型

| 类型 | 说明 |
|---|---|
| `false_merge_generic_topic` | 因泛 topic 误合并 |
| `over_split_topic_variation` | 同一主题因命名不同被拆散 |
| `forced_constellation` | 新主题被硬凑进已有星座 |
| `lonely_not_reused` | 孤星后续没有参与匹配 |
| `profile_stale` | 星座画像没有随新 Trace 演化 |
| `active_pollution` | 成熟星座被弱相关 Trace 污染 |

## 人工复核流程

1. 每次算法或召回链路变更前，选取固定样本集作为 baseline。
2. 对每条 Echo 样本记录：
   - 实际返回的 matched moment IDs。
   - 返回顺序。
   - similarity/score。
   - 是否命中预期正例。
   - 是否包含必须排除项。
3. 对每条星座样本记录：
   - 实际 action。
   - 实际目标星座。
   - 是否误合并或过度拆分。
   - 状态变化是否符合预期。
4. 标记失败类型。
5. 对比优化前后：
   - 不只看命中率，也看失败类型是否减少。
   - 若新增严重失败类型，即使总体命中率提升也不能直接通过。

## 最小通过标准

P1/P2 Echo 优化进入实现验收时，至少应满足：

- 首批 Echo 样本中不得出现 `false_positive_current_trace`。
- 首批 Echo 样本中不得出现 `forced_echo`。
- 表层词陷阱样本应优先排除错误候选。
- 深层处境样本应至少召回一个人工标注正例。

P4-P8 星座优化进入实现验收时，至少应满足：

- 首批星座样本中不得出现明显 `false_merge_generic_topic`。
- 新主题样本应能保留为 lonely 或等价兼容状态。
- topic 命名差异样本应能聚合到同一长期主题。
- 星座画像演化样本应进入画像更新流程。

## 后续扩展样本建议

- 用户长期工作压力样本。
- 亲密关系主动/等待样本。
- 自我怀疑/价值感样本。
- 创作/表达卡住样本。
- 日常轻松记录样本。
- 只有关键词重合但处境不同的负例。
- 处境相似但关键词不同的正例。
- 多个相似候选来自同一 Trace 的去重样本。
- 孤星后续形成 forming 的连续样本。
- active 星座被弱相关内容污染的防护样本。
