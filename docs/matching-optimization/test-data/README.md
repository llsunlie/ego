# Matching Evaluation Test Data

本目录存放 P0 评估基线对应的机器可读测试数据。

这些数据用于后续 Echo 匹配与星座聚合优化的离线评估、人工复核或回归脚本输入。当前数据来自 `docs/matching-optimization/matching-evaluation-plan.md` 中的首批合成样本。

## 文件

| 文件 | 内容 |
|---|---|
| `echo_cases.json` | Echo 匹配评估样本 |
| `constellation_cases.json` | 星座聚合评估样本 |

## 使用约定

- `expected_positive_ids` 表示应当命中的候选。
- `expected_negative_ids` 表示不应命中的候选。
- `must_exclude_ids` 表示无论分数多高都必须排除的候选。
- `expected_action` 表示星座聚合的预期动作。
- 本目录只维护评估数据，不定义算法、权重、prompt 或数据库结构。

## 更新规则

- 新增样本时保持 `case_id` 稳定且唯一。
- 修改样本预期时，需要同步更新 `matching-evaluation-plan.md` 中的说明或理由。
- 样本应优先覆盖失败类型，而不是只覆盖成功路径。
