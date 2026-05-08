# ego · 实体关系文档

> 确认版，派生自产品文档 ego4.0.txt、IA 文档 ia.md 和核心流程 flows.md。

## 1. 实体关系图

```
User
 │
 │ 1:N
 ▼
Moment ──────────────────────────────────────────
 │
 │ 1:N (per Moment, persisted)
 ▼
Echo ─── matched_moment_ids[] + similarities[]
 │
 │ 所属（通过 trace_id）
 ▼
Trace ─── motivation + Item[]
          Item = <Moment, Echo[], Insight>
 │
 │ 1:1 (stashed)
 ▼
Star ─── topic (AI 生成)
 │
 │ N:M (通过 constellation_stars)
 ▼
Constellation
 ├── constellation_insight (TEXT)
 │    单 Star = 该 Star 的 topic
 │    多 Star = AI 重新生成
 │    前端渲染为 "✦ 我发现" 卡片
 │
 ├── PastSelfCard (前端概念，后端数据来自 Star)
 │    点击 → ChatSession
 │              │
 │              │ 1:N
 │              ▼
 │           ChatMessage
 │              │ 引用 MomentReference (date + snippet)
 │
 └── TopicPrompt[] (持久化，锚定 Moment)
     点击 → NowPage 写字 (motivation = constellation:<id>)
```

> **PastSelfCard 说明**：后端不持久化 PastSelfCard 表。前端渲染详情页时，根据 Constellation 下的 Star 列表组装"和那时的自己说说话"卡片——每个 Star 就是一张卡片，Star.topic 作为标题，Star 对应 Trace 最后一个 Moment 的内容作为 opening_line。

## 2. 实体定义

### 核心实体

| 实体 | 说明 |
|------|------|
| **User** | 用户，所有表按 user_id 隔离 |
| **Moment** | 用户写的一句原话，属于一个 Trace |
| **Trace** | 一次写作会话，一等公民，有 motivation |
| **Echo** | 对一个 Moment 的一次历史回声匹配结果 |
| **Insight** | AI 生成的对一个 Moment（结合 Echo）的第二人称观察 |
| **Star** | Trace 寄存后的星图实体，有 AI 生成的 topic |
| **Constellation** | 多个（或单个）Stars 的聚类，维护自身 insight |
| **TopicPrompt** | Constellation 延伸的话题引子，锚定某条 Moment |
| **ChatSession** | 一次"和那时的自己对话"会话（对应一个 Star） |
| **ChatMessage** | 对话中的单条消息 (user / past_self) |
| **MomentReference** | 值对象 (date + snippet)，嵌入 ChatMessage |

### 关系

| 关系 | 类型 | 说明 |
|------|------|------|
| User → Moment | 1:N | |
| User → Trace | 1:N | |
| User → Star | 1:N | |
| User → Constellation | 1:N | |
| User → ChatSession | 1:N | |
| Moment → Trace | N:1 | 按 trace_id 分组 |
| Moment → Echo | 1:N | 每条 Moment 有多个回声匹配结果 |
| Trace → Star | 1:1 | stashed 时产生 |
| Trace → Moment | 1:N | |
| Star → Constellation | N:M | 通过 constellation_stars |
| Constellation → TopicPrompt | 1:N | |
| Star → ChatSession | 1:N | 对话入口对应一个 Star |
| ChatSession → ChatMessage | 1:N | |
| ChatMessage → MomentReference | 嵌入 | JSONB |
| Echo → Moment (被匹配) | N:1 | matched_moment_ids 引用历史 Moment |
| Insight → Echo | N:1 | 基于 echo 生成 |

## 3. 持久化决策

| 概念 | 持久化 | 存储位置 | 说明 |
|------|--------|---------|------|
| Moment | ✅ | `moments` | 含 embeddings JSONB |
| Trace | ✅ | `traces` | 一等公民，有 motivation |
| Echo | ✅ | `echos` | per Moment，matched_moment_ids[] |
| Insight | ✅ | `insights` | per Moment，关联 echo_id |
| Star | ✅ | `stars` | Trace 寄存后生成 |
| Constellation | ✅ | `constellations` | 含 star_ids[], topic_prompts[], constellation_insight |
| TopicPrompt | ❌ 值对象 | constellations.topic_prompts[] | 完整的提问文本 |
| ChatSession | ✅ | `chat_sessions` | 关联 star_id |
| ChatMessage | ✅ | `chat_messages` | |
| MomentReference | ❌ 值对象 | chat_messages 中 JSONB | date + snippet |
| PastSelfCard | ❌ 不存 | 前端从 Star 数据组装 | 见第 1 节说明 |

---

## 4. 表定义

> 所有表无外键约束，无 DEFAULT 值。UUID 和 TIMESTAMPTZ 由后端代码显式设置。
> user_id 由后端保证一致性。

### 4.1 users — 用户

```sql
CREATE TABLE users (
  id            UUID PRIMARY KEY,
  account       VARCHAR(100) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL
);
CREATE UNIQUE INDEX idx_users_account ON users(account);
```

### 4.2 traces — 写作会话

```sql
CREATE TABLE traces (
  id         UUID PRIMARY KEY,
  user_id    UUID NOT NULL,
  motivation VARCHAR(50) NOT NULL,   -- 'direct' | 'trace:<id>' | 'constellation:<id>'
  stashed    BOOLEAN NOT NULL,       -- 是否已收进星图
  created_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX idx_traces_user_created ON traces(user_id, created_at DESC);
```

### 4.3 moments — 用户话语

```sql
CREATE TABLE moments (
  id         UUID PRIMARY KEY,
  user_id    UUID NOT NULL,
  trace_id   UUID NOT NULL,
  content    TEXT NOT NULL,
  embeddings JSONB NOT NULL DEFAULT '[]'::JSONB,
  created_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX idx_moments_user_created ON moments(user_id, created_at DESC);
CREATE INDEX idx_moments_trace        ON moments(trace_id);
```

> **embeddings JSONB 结构：**
> ```json
> [
>   {"model": "text-embedding-004", "embedding": [0.012, -0.034, ...]}
> ]
> ```
> 设计要点：所有模型版本的 embedding 都存在同一字段，切换 model 时追加新记录即可，不动其他列。
> 余弦相似度计算在应用层进行。数据量增长后如需加速，可单独建 VECTOR 列并创建 ivfflat/hnsw 索引。

### 4.4 echos — 回声匹配结果

```sql
CREATE TABLE echos (
  id                 UUID PRIMARY KEY,
  moment_id          UUID NOT NULL,   -- 为哪个 Moment 匹配的
  user_id            UUID NOT NULL,
  matched_moment_ids UUID[] NOT NULL, -- 匹配到的历史 Moment 列表（按相似度降序）
  similarities       FLOAT[] NOT NULL, -- 对应相似度，元素与 matched_moment_ids 一一对应
  created_at         TIMESTAMPTZ NOT NULL
);
CREATE INDEX idx_echos_moment ON echos(moment_id);
```

> 设计要点：每次 CreateMoment 写入一条 Echo 记录。PastPage 查看 Trace 详情时，按 moment_id 查询对应 Echo。

### 4.5 insights — AI 观察（per-Moment）

```sql
CREATE TABLE insights (
  id                 UUID PRIMARY KEY,
  user_id            UUID NOT NULL,
  moment_id          UUID NOT NULL,   -- 为哪个 Moment 生成的观察
  echo_id            UUID NOT NULL,   -- 基于哪个 Echo 生成的
  text               TEXT NOT NULL,   -- 第二人称观察文本
  related_moment_ids UUID[] NOT NULL, -- 观察关联的 Moment id 列表
  created_at         TIMESTAMPTZ NOT NULL
);
CREATE INDEX idx_insights_moment ON insights(moment_id);
CREATE INDEX idx_insights_echo   ON insights(echo_id);
```

> 设计要点：Insight 仅负责 per-Moment 的观察。Constellation 级别的观察存于 `constellations.constellation_insight`。

### 4.6 stars — 星图实体

```sql
CREATE TABLE stars (
  id         UUID PRIMARY KEY,
  user_id    UUID NOT NULL,
  trace_id   UUID NOT NULL,   -- 关联的 Trace
  topic      TEXT NOT NULL,   -- AI 生成的主题文本
  created_at TIMESTAMPTZ NOT NULL
);
CREATE UNIQUE INDEX idx_stars_trace ON stars(trace_id);  -- 一个 Trace 只能生成一颗 Star
```

### 4.7 constellations — 星座

```sql
CREATE TABLE constellations (
  id                   UUID PRIMARY KEY,
  user_id              UUID NOT NULL,
  name                 VARCHAR(100) NOT NULL,
  constellation_insight TEXT NOT NULL, -- Constellation 级 AI 观察
  star_ids             UUID[] NOT NULL, -- 包含的 Star id 列表
  topic_prompts        TEXT[] NOT NULL, -- 话题引子列表
  created_at           TIMESTAMPTZ NOT NULL,
  updated_at           TIMESTAMPTZ NOT NULL
);
CREATE INDEX idx_constellations_user ON constellations(user_id);
CREATE INDEX idx_constellations_stars ON constellations USING GIN (star_ids);
```

> 设计要点：
> - `star_ids` 直接存储 Star 引用，替代 constellation_stars 中间表。查询"某个 Star 属于哪些 Constellation"使用 `WHERE star_ids @> ARRAY[$star_id]::UUID[]`。
> - `topic_prompts` 存储完整的提问文本列表，每个元素为一句 AI 生成的话题引子（锚定的 Moment 引用已包含在文本中）。
> - 聚类触发时更新 Constellation 记录，单 Star（孤星）也是一个 Constellation 记录。

### 4.8 chat_sessions — 对话会话

```sql
CREATE TABLE chat_sessions (
  id                 UUID PRIMARY KEY,
  user_id            UUID NOT NULL,
  star_id            UUID NOT NULL,        -- 对应一个 Star（替代 PastSelfCard）
  context_moment_ids UUID[] NOT NULL,      -- 上下文 Moment 列表
  created_at         TIMESTAMPTZ NOT NULL
);
CREATE INDEX idx_chat_sessions_user ON chat_sessions(user_id);
```

### 4.9 chat_messages — 对话消息

```sql
CREATE TABLE chat_messages (
  id                 UUID PRIMARY KEY,
  user_id            UUID NOT NULL,
  session_id         UUID NOT NULL,
  role               VARCHAR(10) NOT NULL,  -- 'user' | 'past_self'
  content            TEXT NOT NULL,
  referenced_moments JSONB,                 -- MomentReference[]: [{date, snippet}, ...]
  created_at         TIMESTAMPTZ NOT NULL
);
CREATE INDEX idx_chat_messages_session ON chat_messages(session_id, created_at);
```

---

## 5. ER 概览

```
users
  │ (user_id 隔离所有表，无 FK)
  │
  ├── traces ── 1:N ── moments (embeddings JSONB)
  │     │                    │
  │     │ 1:1                ├── 1:N ── echos
  │     ▼                    │              │
  │   stars                  │              │ 关联历史 Moment UUID[]
  │     │                    │              │
  │     │                    ├── 1:N ── insights (moment_id + echo_id)
  │     │                    │
  │     │ star_ids UUID[]    │
  │     ▼                    │
  │   constellations ◄──────── (star_ids, topic_prompts TEXT[], constellation_insight)
  │
  └── chat_sessions (star_id)
        │
        │ 1:N
        ▼
      chat_messages (referenced_moments JSONB)
```

> Star 与 Constellation 的关系：`constellations.star_ids` 包含 Star UUID。一个 Star 可被多个 Constellation 引用。

## 6. Migration 顺序

```
1. users
2. traces
3. moments
4. echos
5. insights
6. stars
7. constellations
8. chat_sessions
9. chat_messages
10. INDEXES（GIN 索引最后建）
```

> 共 9 张表，无外键约束。按数据依赖顺序创建。`CREATE EXTENSION vector` 在需要 pgvector 加速时再执行。

