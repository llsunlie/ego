# P5 TraceProfile 质量验证与调优

本文记录 P5 的执行边界和当前落地产物。

P5 的目标不是替换星座聚合，而是先确认 P4 旁路生成的 TraceProfile 是否足够稳定、具体、克制，可作为后续 ConstellationProfile 匹配的输入。

## 验证目标

TraceProfile 质量复核重点：

- `topic` 是否短、稳定、直接，避免诗化或过泛。
- `summary` 是否只概括输入中已有内容，不补背景。
- `keywords` 是否具体，避免“事情、感觉、问题”等泛词。
- `emotions` 是否有明确证据，不把普通事件强行情绪化。
- `scenes` 是否贴近用户原文场景。
- `central_pattern` 是否可为空，不强行制造“核心冲突”。
- `representative_moment_index` 是否能映射到输入 moment；持久化后的 `representative_moment_id` 是否来自输入 moment。
- 输出中是否避免心理诊断、治疗化或人格推断词。

## 当前样本

新增固定样本：

`docs/matching-optimization/test-data/trace_profile_cases.json`

首批覆盖：

- 亲密关系里的主动与等待。
- 新的日常体验，不应强行制造冲突。
- 入职延迟后的重新安排。
- 朋友聊天中的表达顾虑。

这些样本用于人工复核真实 LLM 输出，也可作为后续离线评估脚本输入。

## 当前自动化覆盖

新增 Go 纯函数测试：

`server/internal/starmap/adapter/ai/trace_profile_generator_test.go`

覆盖内容：

- JSON 解析允许模型返回 markdown fence 包裹的 JSON。
- `representative_moment_index` 越界时回落到兼容 ID；仍不可用时回落到第一个 moment。
- list 字段去空、去重、限长。
- user prompt 保持 moment 顺序，并在 motivation 为空时不输出空字段。
- `profile_text` 使用结构化字段和代表原文构造。
- fallback profile 允许 `central_pattern` 为空。

这些测试不调用真实 LLM，也不调用 embedding 服务。

## 人工复核流程

1. 使用真实书写流程 stash trace。
2. 查看日志中的 `TraceProfileGenerator: generated profile`。
3. 对照 `trace_profile_cases.json` 中的 `expected_quality` 复核字段质量。
4. 如果发现过泛、过度推断或代表 moment 错误，先调整 prompt 或字段规范化规则。
5. 再运行 `go test ./internal/starmap/...` 做回归。

## 非目标

P5 不做：

- 不替换当前 topic-based constellation clustering。
- 不新增 ConstellationProfile 数据结构。
- 不实现 TraceProfile 与星座画像的匹配评分。
- 不让 fallback profile 直接参与正式星座匹配决策。

## 下一步

P5 通过后，再进入 ConstellationProfile 设计与星座匹配替换：

- 设计 ConstellationProfile 字段。
- 讨论 profile embedding 与 centroid embedding 的职责。
- 设计从现有 constellation topic 到画像结构的迁移路径。
- 设计 TraceProfile -> ConstellationProfile 的候选召回与评分流程。
