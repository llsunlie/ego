# P1 Echo 召回效率优化设计与实施记录

## 文档定位

本文是 P1「Echo 召回效率优化」的设计与实施记录。

目标是将 Echo 候选召回从应用层全量扫描迁移到 PostgreSQL pgvector/HNSW topK 召回，为后续 P2「Echo 匹配质量优化」提供稳定候选集。

本文只覆盖 P1，不包含 P2 的具体候选过滤、时间权重、同 Trace 去重、LLM rerank 或 Echo score 权重设计。

## 已确认约束

1. Echo 必须实时返回。
2. Echo 匹配不引入 insight embedding。
3. Echo topK 候选召回使用 pgvector/HNSW。
4. proto 暂不修改。
5. 前端继续展示当前 Echo 的 similarity 数值。
6. P1 以效率优化为主，不改变 Echo 的产品语义判断。
7. 当前 active embedding 模型为 `BAAI/bge-m3`，向量维度为 1024。
8. 接受新增独立表 `moment_embedding_vectors`。
9. 接受 `moments.embeddings` JSONB 与 pgvector 表长期并存。
10. 新增 `ECHO_RECALL_TOP_K` 配置，默认值为 10。
11. pgvector 查询失败时不允许临时回退全量扫描。
12. 历史回填采用独立命令入口。
13. vector 参数传递先使用字符串 literal，暂不引入 `pgvector-go`。

## 当前实现现状

### 数据现状

`moments` 表当前保存：

```sql
CREATE TABLE moments (
  id         UUID PRIMARY KEY,
  trace_id   UUID NOT NULL,
  user_id    UUID NOT NULL,
  content    TEXT NOT NULL,
  embeddings JSONB NOT NULL DEFAULT '[]'::JSONB,
  created_at TIMESTAMPTZ NOT NULL
);
```

其中 `embeddings` 是 JSONB 多模型数组：

```json
[
  {
    "model": "xxx",
    "embedding": [0.1, 0.2]
  }
]
```

### 代码现状

当前 `CreateMomentUseCase` 流程：

```text
生成 embedding
  -> 保存 Moment(JSONB embeddings)
  -> moments.ListByUserID(userID) 读取全量历史
  -> excludeSelf
  -> DefaultEchoMatcher 在应用层逐条计算 cosine similarity
  -> similarity >= 0.65 的结果写入 Echo
```

当前瓶颈：

- 每次 Echo 都加载用户全量历史 Moment。
- 每条历史 Moment 都在应用层解析 JSONB embedding。
- 召回和排序都发生在 Go 内存里。

## 实施方案

P1 采用 **新增独立向量表 + JSONB 兼容保留 + 数据库 topK 查询**。

不建议直接把 `moments.embeddings` 从 JSONB 改成 pgvector，也不建议立刻删除 JSONB 字段。

原因：

- 当前领域模型支持多模型 embedding，JSONB 仍有兼容价值。
- pgvector 的 HNSW 索引需要固定维度，和“多模型/未来换模型”存在天然张力。
- 独立表可以把“当前用于 Echo 召回的 active content embedding”单独索引，不破坏现有读写模型。

## 目标数据结构

新增表命名：

```text
moment_embedding_vectors
```

字段：

```sql
CREATE TABLE moment_embedding_vectors (
  moment_id  UUID NOT NULL REFERENCES moments(id) ON DELETE CASCADE,
  user_id    UUID NOT NULL,
  trace_id   UUID NOT NULL,
  model      TEXT NOT NULL,
  dim        INT NOT NULL,
  embedding  VECTOR(1024) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  PRIMARY KEY (moment_id, model)
);
```

说明：

- `moment_id`：对应 Moment。
- `user_id`：用于用户隔离过滤，避免 join 后才能过滤。
- `trace_id`：为 P2 当前 Trace 排除、同 Trace 去重预留。
- `model`：记录 active content embedding 的模型名。
- `dim`：记录维度，便于校验和未来迁移。
- `embedding`：pgvector 字段，用于 HNSW topK。
- `created_at`：用于候选排序、时间规则和回填排查。

### 维度处理

pgvector HNSW 索引要求稳定维度。当前 active embedding 模型为 `BAAI/bge-m3`，维度已确认为 1024。

处理方式：

1. 在迁移中写死 `VECTOR(1024)`。
2. 新增 `AI_EMBEDDING_DIM` 配置，默认值为 1024。
3. 在应用启动或 embedding 写入时校验实际维度与配置维度一致。
4. 如果未来更换不同维度模型，新增迁移或新表版本，不在同一 HNSW 索引中混用不同维度。

## 目标索引

迁移中启用 pgvector 扩展：

```sql
CREATE EXTENSION IF NOT EXISTS vector;
```

建议索引：

```sql
CREATE INDEX idx_moment_embedding_vectors_user_model
ON moment_embedding_vectors(user_id, model);

CREATE INDEX idx_moment_embedding_vectors_trace
ON moment_embedding_vectors(trace_id);

CREATE INDEX idx_moment_embedding_vectors_embedding_hnsw
ON moment_embedding_vectors
USING hnsw (embedding vector_cosine_ops)
WITH (m = 16, ef_construction = 64);
```

说明：

- HNSW 用于向量近邻。
- `user_id, model` 用于过滤。
- `trace_id` 主要为 P2 预留。

## topK 查询设计

新增查询目标：

```text
输入：user_id, model, current_moment_id, query_embedding, limit
输出：按向量距离排序的历史 Moment 候选
```

建议 SQL 形态：

```sql
SELECT
  m.id,
  m.trace_id,
  m.user_id,
  m.content,
  m.embeddings,
  m.created_at,
  (1 - (mev.embedding <=> sqlc.arg(query_embedding)::vector)) AS similarity
FROM moment_embedding_vectors mev
JOIN moments m ON m.id = mev.moment_id
WHERE mev.user_id = $1
  AND mev.model = $2
  AND mev.moment_id <> $3
ORDER BY mev.embedding <=> sqlc.arg(query_embedding)::vector
LIMIT $4;
```

P1 只强制排除当前 Moment 自身。排除当前 Trace、同 Trace 去重、时间距离规则进入 P2 设计。

### pgx 参数传递建议

为减少新依赖，P1 可以先使用 vector literal 字符串：

```text
"[0.1,0.2,0.3]"
```

SQL 中通过 `sqlc.arg(query_embedding)::vector` 转换。

备选方案是引入 `pgvector-go` 类型并配置 sqlc override。该方式类型更强，但会增加依赖和 sqlc 配置复杂度。

P1 先使用 vector literal，等 P1 验证稳定后再评估是否引入 `pgvector-go`。

## 写入策略

采用 **Moment JSONB 与 pgvector 向量表双写**。

### 新 Moment 写入

当前 `MomentRepository.Create` 写入 `moments` 表。P1 需要扩展为：

```text
MomentRepository.Create
  -> 写入 moments.embeddings JSONB
  -> 写入 moment_embedding_vectors
```

使用单条 CTE SQL 保证 Moment 与向量表写入原子性：

```sql
WITH inserted AS (
  INSERT INTO moments (...)
  VALUES (...)
  RETURNING id, trace_id, user_id, created_at
)
INSERT INTO moment_embedding_vectors (...)
SELECT inserted.id, inserted.user_id, inserted.trace_id, ...
FROM inserted;
```

原因：

- 当前 repository 未显式使用事务。
- 如果分两条 SQL，可能出现 Moment 已保存但向量表未写入的半成功状态。
- CTE 能在不大改事务边界的情况下保证新 Moment 的双写一致。

如果实现时保留旧 `CreateMoment` sqlc 查询，也可以新增 `CreateMomentWithVector` 查询用于 Writing 新路径。

### 历史数据回填

P1 提供历史 Moment 的向量回填能力。

实现方式：

```text
Go 回填命令或一次性维护脚本
  -> 扫描 moments.embeddings JSONB
  -> 选择 active model 对应 embedding
  -> 校验维度
  -> upsert 到 moment_embedding_vectors
```

不推荐用纯 SQL 从 JSONB 直接回填：

- JSONB 数组转 pgvector literal 可读性差。
- 难做模型选择、维度校验和错误日志。
- 后续模型变更时仍需要 Go 侧能力。

回填命令位置为 `server/cmd/backfill-moment-vectors`，支持幂等 upsert。

## 应用层接口改造

### 新增领域端口

Writing domain 增加候选召回端口：

```go
type EchoCandidateReader interface {
    FindNearestMoments(
        ctx context.Context,
        userID string,
        currentMomentID string,
        model string,
        embedding []float32,
        limit int32,
    ) ([]Moment, error)
}
```

说明：

- P1 输出仍是 `[]Moment`，保持 `DefaultEchoMatcher` 可复用。
- P1 不在端口中暴露 pgvector、距离运算符或 SQL 细节。
- P2 如需 similarity 原始值，可再讨论是否扩展返回结构。

### CreateMomentUseCase 改造

当前：

```text
uc.moments.ListByUserID(userID)
  -> excludeSelf
  -> echo.Match(current, history)
```

P1 改为：

```text
currentEmbedding := current.Embeddings[0]
history := uc.echoCandidates.FindNearestMoments(
  userID,
  current.ID,
  currentEmbedding.Model,
  currentEmbedding.Embedding,
  topK,
)
matches := uc.echo.Match(current, history)
```

### topK 配置

新增配置：

```text
ECHO_RECALL_TOP_K=10
```

默认值为 10。

配置读取和注入路径：

- `config` 层读取 `ECHO_RECALL_TOP_K`，默认 10。
- `bootstrap` 解析并注入 Writing module。
- Writing module 注入 `CreateMomentUseCase`。

P1 起配置化，方便评估调参。

## 适配器改造

### Postgres adapter

新增：

```text
writing/adapter/postgres/EchoCandidateReader
```

采用独立 reader，避免 MomentRepository 继续膨胀。

### 查询实现

P1 暂未新增 sqlc 查询，而是在 Postgres adapter 中使用 raw SQL：

```text
CreateMomentWithVector CTE
FindNearestMomentsByEmbedding
UpsertMomentEmbeddingVector
```

原因是 vector 参数暂采用字符串 literal，并避免 P1 同时引入 sqlc override 或 `pgvector-go`。

## 兼容与回退策略

P1 推荐提供两层兼容：

1. 数据兼容：
   - 保留 `moments.embeddings` JSONB。
   - 新增 `moment_embedding_vectors` 作为检索索引表。

2. 运行兼容：
   - 如果当前 Moment 没有 embedding，Echo 仍返回 nil。
   - 如果 pgvector 查询失败，不允许临时回退应用层全量扫描。

已确认决策：

- pgvector 查询失败时直接作为 Echo 匹配错误暴露给当前用例。
- 不做全量扫描回退，避免掩盖向量表缺失、回填不完整或索引异常。
- 失败时必须记录 error 日志，便于排查。

## 测试计划

### 单元测试

- vector literal 构造函数测试。
- embedding 维度校验测试。
- `CreateMomentUseCase` 使用 topK reader 而不是全量 reader 的 mock 测试。
- 当前 Moment 无 embedding 时不调用 topK reader。

### Postgres 集成测试

- migration 后 `vector` extension 可用。
- `moment_embedding_vectors` 可以写入。
- HNSW 查询按距离返回候选。
- `FindNearestMoments` 只返回同用户、同模型候选。
- `FindNearestMoments` 排除当前 Moment。
- 回填 upsert 幂等。

### P0 样本回归

P1 不要求完全解决 P0 中所有质量问题，但至少要保证：

- 候选召回能覆盖 P0 Echo 样本中人工正例。
- 不因 topK 过小导致明显正例进不了候选集。
- 后续 P2 有足够候选可用于质量重排。

## 验收标准

P1 完成后应满足：

1. 新 Moment 写入时，content embedding 同步进入 pgvector 向量表。已实现。
2. 历史 Moment 可通过回填进入向量表。已实现。
3. Echo 候选召回不再调用 `ListByUserID` 全量加载作为主路径。已实现。
4. Echo 候选由数据库 topK 返回。已实现。
5. 现有 API/proto 不变。已满足。
6. `moments.embeddings` JSONB 保留，Timeline/Starmap/Conversation 现有读取不受影响。
7. 相关单元测试与 Postgres 集成测试通过。

## 推荐实施顺序

待确认后建议按以下顺序执行：

1. 补充 migration：`CREATE EXTENSION vector`、新增向量表和索引。
2. 增加 sqlc 查询：向量写入、topK 查询、回填 upsert。
3. 增加 vector literal / 维度校验工具。
4. 扩展 Postgres adapter，实现 `EchoCandidateReader`。
5. 改造 `CreateMomentUseCase`，从 topK reader 获取候选。
6. 接入 `ECHO_RECALL_TOP_K` 配置。
7. 增加历史回填命令。
8. 补测试。
9. 用 P0 test-data 做候选覆盖人工复核。

## 已确认问题

| 问题 | 决策 |
|---|---|
| 当前 active embedding 模型向量维度 | `BAAI/bge-m3` / 1024 |
| 是否接受新增独立表 `moment_embedding_vectors` | 是 |
| 是否接受 `moments.embeddings` JSONB 与 pgvector 表长期并存 | 是 |
| `ECHO_RECALL_TOP_K` 默认值 | 10 |
| pgvector 查询失败时是否允许临时回退全量扫描 | 否 |
| 历史回填入口 | 独立命令 |
| vector 参数传递 | 字符串 literal，暂不引入 `pgvector-go` |

## 不在 P1 中处理

- insight embedding。
- Echo 强/弱/无分档。
- 当前 Trace 排除规则。
- 同 Trace 候选去重。
- 时间距离权重。
- keyword overlap。
- LLM rerank。
- Echo score 权重重定义。
- proto 修改。
