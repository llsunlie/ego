# P2.5 Elasticsearch Sparse Search 实施记录

## 文档定位

本文记录 P2.5 引入 Elasticsearch sparse search 的实施结果。

P2.5 的目标不是替代 P1 pgvector dense recall，而是补足 dense embedding 对具体短语、反复措辞、词面呼应不敏感的问题。

## 目标链路

```text
Moment 写入
  -> PostgreSQL moments
  -> pgvector moment_embedding_vectors
  -> Elasticsearch ego_moments index

Echo 查询
  -> 并发启动 pgvector dense topK 与 Elasticsearch sparse topK
  -> RRF 融合候选
  -> P2 EchoRanker 规则过滤与排序
  -> 保存 Echo
```

PostgreSQL 仍是事实来源，Elasticsearch 只作为检索索引。

## Elasticsearch index

当前 index：

```text
ego_moments
```

document：

```json
{
  "moment_id": "xxx",
  "user_id": "xxx",
  "trace_id": "xxx",
  "content": "每次都是我先开口，我真的有点累。",
  "created_at": "2026-06-04T10:00:00Z"
}
```

## 中文分词策略

第一版使用：

```text
content.ik      主召回字段
content.ngram   兜底字段
```

字段策略：

| 字段 | analyzer | 用途 |
|---|---|---|
| `content.ik` | `ik_max_word`，search 使用 `ik_smart` | 主 BM25 sparse search |
| `content.ngram` | char bigram/trigram | 分词失败兜底 |

查询时：

```text
content.ik boost 1.0
content.ngram boost 0.2 - 0.3
```

## RRF 融合

不直接把 cosine score 和 ES BM25 score 相加。使用 RRF 按排名融合：

```text
rrf_score = 1 / (K + dense_rank) + 1 / (K + sparse_rank)
```

默认：

```text
ECHO_HYBRID_RRF_K=60
```

`K=60` 是常见稳健默认值，用于避免某一路第一名过度碾压其他候选。

## 查询并发与失败语义

当前 `CreateMoment` 保存 Moment 后，会先 best-effort 写入 ES，然后进入 Echo 召回。Echo 召回阶段中：

- pgvector/HNSW dense topK 与 ES sparse topK 并发执行。
- ES sparse 路径包含 `SearchMomentIDs` 与 PostgreSQL `GetByIDs` 回读。
- 两路完成后再进入 RRF 融合。

失败语义：

- pgvector dense 查询失败：`CreateMoment` 返回错误，不回退全量扫描。
- ES sparse 查询失败：记录 warn，sparse 候选视为空，继续使用 dense 候选。
- ES sparse ID 回读 PostgreSQL 失败：记录 warn，sparse 候选视为空，继续使用 dense 候选。

## 召回日志口径

P2.5 的日志目标是帮助判断召回质量，而不是记录每个基础设施动作。

保留的核心日志：

- `CreateMoment: echo recall candidates`
  - 当前 Moment 的 `current_preview`
  - pgvector dense 召回候选 `dense_candidates`
  - Elasticsearch sparse 召回候选 `es_candidates`
  - RRF 融合后的 `fused_candidates`
  - 每个候选最多记录前 5 条，字段包括 `rank`、`moment_id`、`trace_id`、`created_at`、`content_preview`
  - 同时记录 `dense_candidate_count`、`sparse_candidate_count`、`fused_candidate_count`、`dense_top_k`、`sparse_top_k`、`rrf_k`

- `CreateMoment: echo final matches`
  - EchoMatcher 最终命中的历史 Moment
  - 每条记录包括 `rank`、`moment_id`、`trace_id`、`created_at`、`content_preview`、`similarity`

降噪规则：

- 不记录 ES 每次 HTTP 请求成功日志。
- 不记录 ES 写入成功日志。
- 不记录 sparse ids loaded / sparse moments loaded / hybrid merged 这类中间碎片日志。
- 不在 gRPC composite 层记录完整 req/res，避免用户原文被重复输出。
- ES 写入失败、sparse recall 失败、pgvector 查询失败仍按错误等级记录。

内容日志只记录 `content_preview`，当前限制为 48 个 rune，避免整段用户输入刷屏。

## 写入策略

第一版建议：

```text
同步 best-effort 写 ES
```

行为：

- ES 写入成功：正常。
- ES 写入失败：记录 warn/error，不阻断 CreateMoment。
- sparse recall 不可用时，记录 warn，并继续使用 pgvector dense recall。

后续生产化可改为：

```text
PostgreSQL outbox + 后台同步 ES
```

## 历史回填

新增独立命令：

```text
server/cmd/backfill-moment-search
```

职责：

- 扫描历史 `moments`。
- 构造 ES document。
- bulk index 到 `ego_moments`。
- 支持幂等重复执行。
- 输出 scanned/indexed/skipped/failed。

## 配置建议

```text
ELASTICSEARCH_URL=http://localhost:9200
ELASTICSEARCH_USERNAME=
ELASTICSEARCH_PASSWORD=
ECHO_SPARSE_RECALL_ENABLED=true
ECHO_SPARSE_RECALL_TOP_K=10
ECHO_HYBRID_RRF_K=60
```

P1 配置继续保留：

```text
ECHO_RECALL_TOP_K=10
```

## 服务依赖

已扩展 `docker-compose.yml`：

```text
elasticsearch
```

如果采用 IK analyzer，建议构建自定义镜像：

```text
docker/elasticsearch/Dockerfile
  -> base elasticsearch
  -> install analysis-ik
```

## 不在 P2.5 第一版中处理

- 不引入 ELSER。
- 不引入 LLM rerank。
- 不让 ES 成为事实来源。
- 不让 ES 写入失败阻断 CreateMoment。

## 当前实现入口

| 范围 | 文件 |
|---|---|
| Platform ES client | `server/internal/platform/elasticsearch/client.go` |
| Writing ES adapter | `server/internal/writing/adapter/elasticsearch/moment_search.go` |
| CreateMoment hybrid recall | `server/internal/writing/app/create_moment.go` |
| RRF 融合 | `server/internal/writing/app/echo_hybrid.go` |
| 历史回填 | `server/cmd/backfill-moment-search` |
| 本地服务 | `docker-compose.yml`, `docker/elasticsearch/Dockerfile` |
