# ego · 数据库设计

> PostgreSQL + pgvector · 10 张表 · 无外键约束 · 无 DEFAULT 值

## 1. ER 概览

```
users
  │ (user_id 隔离所有表，无 FK)
  │
  └── moments ────────────────────────────────────────────
        │                                                    │
        │ trace_id (UUID, 分组键)                             │
        │                                                    │
        └── stars.trace_id ──────────────────────────────────┘
              │
              │ (N:M)
              ▼
      constellation_stars ────── constellations
              │                       │
              │                       ├──(1:N)── insights
              │                       ├──(1:N)── topic_prompts
              │                       └──(1:N)── past_self_cards
              │                                       │
              │                                       └──(1:N)── chat_sessions
              │                                                       │
              │                                                       └──(1:N)── chat_messages
```

所有表通过 `user_id` 隔离用户数据，user_id 无外键约束，逻辑由后端保证。

## 2. 表定义

### 2.1 users — 用户

```sql
CREATE TABLE users (
  id            UUID PRIMARY KEY,
  account       VARCHAR(100) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL
);

CREATE UNIQUE INDEX idx_users_account ON users(account);
```

**要点：**
- 无 DEFAULT 值 — id、created_at 由后端代码在 INSERT 时显式设置
- 无外键 — 其他表的 user_id 由后端保证一致性
- password_hash 使用 bcrypt，后端生成
- Login 时 account 不存在则自动注册（INSERT）

### 2.2 moments — 用户话语

```sql
CREATE TABLE moments (
  id         UUID PRIMARY KEY,
  user_id    UUID NOT NULL,
  content    TEXT NOT NULL,
  embedding  VECTOR(768),
  trace_id   UUID NOT NULL,
  connected  BOOLEAN NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_moments_user_created ON moments(user_id, created_at DESC);
CREATE INDEX idx_moments_user_trace   ON moments(user_id, trace_id);
CREATE INDEX idx_moments_embedding    ON moments USING ivfflat (embedding vector_cosine_ops)
  WITH (lists = 50);
```

**要点：**
- `id`、`user_id`、`connected`、`created_at` 均由后端代码显式设置
- `connected` 后端 INSERT 时设为 FALSE
- `trace_id` 是前端生成的 UUID，同一轮对话的所有 Moment 共享
- `embedding` 为 NULL 时表示尚未生成，不参与回声匹配

### 2.3 insights — AI 观察（仅星座级别）

```sql
CREATE TABLE insights (
  id                 UUID PRIMARY KEY,
  user_id            UUID NOT NULL,
  constellation_id   UUID NOT NULL,
  text               TEXT NOT NULL,
  related_moment_ids UUID[] NOT NULL,
  created_at         TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_insights_user_constellation ON insights(user_id, constellation_id);
```

**要点：**
- 只存星座级别的观察（DetailPage ① "✦ 我发现"）
- `related_moment_ids` 由 AI 聚类时生成，引用关联的 Moment id
- Trace 级别的观察由 `GenerateInsight` RPC 实时生成，不持久化

### 2.4 stars — 星星

```sql
CREATE TABLE stars (
  id           UUID PRIMARY KEY,
  user_id      UUID NOT NULL,
  trace_id     UUID NOT NULL,
  x            FLOAT NOT NULL,
  y            FLOAT NOT NULL,
  visual_state VARCHAR(10) NOT NULL,
  rhythm       FLOAT NOT NULL,
  created_at   TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_stars_user_trace ON stars(user_id, trace_id);
```

**要点：**
- `id`、`user_id`、`visual_state`、`rhythm`、`created_at` 均由后端代码显式设置
- `trace_id` 由前端传入，后端保证同一 user 下 trace_id 唯一（业务逻辑约束）
- `visual_state` 后端 INSERT 时设为 `'dim'`
- `rhythm` 后端生成随机值
- 坐标由前端传入（在星图可用区域选一个不重叠的位置）

### 2.5 constellations + constellation_stars — 星座

```sql
CREATE TABLE constellations (
  id         UUID PRIMARY KEY,
  user_id    UUID NOT NULL,
  name       VARCHAR(100) NOT NULL,
  status     VARCHAR(10) NOT NULL,
  created_at TIMESTAMPTZ NOT NULL,
  updated_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE constellation_stars (
  constellation_id UUID NOT NULL,
  star_id          UUID NOT NULL,
  PRIMARY KEY (constellation_id, star_id)
);

CREATE INDEX idx_cs_star_id ON constellation_stars(star_id);
```

**要点：**
- 星星与星座是 N:M——同一颗星理论上可属于多个主题
- 孤星也是一条 constellation 记录 (status=lone) + 一条 constellation_stars
- 聚类触发时机：StashTrace 之后异步执行
- 聚类结果影响：constellation 的创建/合并/状态变更、insight 生成、past_self_cards 生成、topic_prompts 生成、moments.connected 更新

### 2.6 past_self_cards — 对话入口卡片

```sql
CREATE TABLE past_self_cards (
  id                 UUID PRIMARY KEY,
  user_id            UUID NOT NULL,
  constellation_id   UUID NOT NULL,
  title              VARCHAR(200) NOT NULL,
  opening_line       TEXT NOT NULL,
  context_moment_ids UUID[] NOT NULL,
  created_at         TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_past_self_cards_user_constellation ON past_self_cards(user_id, constellation_id);
```

### 2.7 topic_prompts — 话题引子

```sql
CREATE TABLE topic_prompts (
  id                UUID PRIMARY KEY,
  user_id           UUID NOT NULL,
  constellation_id  UUID NOT NULL,
  anchor_moment_id  UUID NOT NULL,
  question          TEXT NOT NULL,
  created_at        TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_topic_prompts_user_constellation ON topic_prompts(user_id, constellation_id);
```

### 2.8 chat_sessions + chat_messages — 对话记录

```sql
CREATE TABLE chat_sessions (
  id                 UUID PRIMARY KEY,
  user_id            UUID NOT NULL,
  past_self_card_id  UUID NOT NULL,
  context_moment_ids UUID[] NOT NULL,
  created_at         TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_chat_sessions_user_card ON chat_sessions(user_id, past_self_card_id);

CREATE TABLE chat_messages (
  id                 UUID PRIMARY KEY,
  user_id            UUID NOT NULL,
  session_id         UUID NOT NULL,
  role               VARCHAR(10) NOT NULL,
  content            TEXT NOT NULL,
  referenced_moments JSONB,
  created_at         TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_chat_messages_session ON chat_messages(user_id, session_id, created_at);
```

---

## 3. 核心查询 → RPC 映射

### Login

```
1. SELECT id, password_hash FROM users WHERE account = $1;
   → 不存在: INSERT INTO users (id, account, password_hash, created_at) VALUES (...)
   → 存在: bcrypt 验证 → 签发 JWT
```

### CreateMoment

```
1. 调用 AI embedding API → 拿到 content 的向量
2. INSERT INTO moments (id, user_id, content, embedding, trace_id, connected, created_at)
   VALUES ($id, $user_id, $content, $embedding, $trace_id, FALSE, NOW())
3. 向量搜索匹配回声（仅当前用户）：
   SELECT id, content, created_at, connected, trace_id,
          1 - (embedding <=> $embedding) AS similarity
   FROM moments
   WHERE user_id = $user_id
     AND id != $new_moment_id
     AND embedding IS NOT NULL
   ORDER BY embedding <=> $embedding
   LIMIT 4;
   -- #1 → echo.target_moment, #2-#4 → echo.candidates
4. 组装 CreateMomentRes 返回（moment + echo），回声不存库

F2 "顺着再想想" 变体：步骤 3 的向量搜索多加一个条件
  AND trace_id != $current_trace_id   -- 排除同一轮 Trace 内已写的 Moment
```

### GetRandomMoments

```sql
SELECT id, content, created_at, connected, trace_id
FROM moments
WHERE user_id = $user_id
  AND embedding IS NOT NULL
ORDER BY random()
LIMIT $count;
```

### ListMoments（游标分页）

```sql
SELECT id, content, created_at, connected, trace_id
FROM moments
WHERE user_id = $user_id
  AND ($cursor = '' OR created_at < (SELECT created_at FROM moments WHERE id = $cursor::UUID AND user_id = $user_id))
ORDER BY created_at DESC
LIMIT $page_size;
-- 前端按 created_at 自行按月分组
```

### GenerateInsight（纯 AI 调用，不写库）

```
1. SELECT content FROM moments WHERE id = $echo_moment_id AND user_id = $user_id → 拿回声内容作为 AI 上下文
2. AI 生成第二人称观察文本
3. 直接返回 GenerateInsightRes，不 INSERT
```

### StashTrace

```
1. INSERT INTO stars (id, user_id, trace_id, x, y, visual_state, rhythm, created_at)
   VALUES ($id, $user_id, $trace_id, $x, $y, 'dim', random(), NOW())
2. 异步触发聚类（仅当前用户范围）：
   - 向量聚合 + AI → 确定该星归属
   - 新建 constellation 或加入已有 constellation
   - 更新星座 status
   - 生成/更新 insight、past_self_cards、topic_prompts
   - 调整 moments.connected
```

### ListConstellations

```sql
SELECT c.*, s.*
FROM constellations c
JOIN constellation_stars cs ON cs.constellation_id = c.id
JOIN stars s ON s.id = cs.star_id
WHERE c.user_id = $user_id
ORDER BY c.updated_at DESC;
-- 后端组装
```

### GetConstellation

```
6 条查询并行执行：
SELECT * FROM constellations WHERE id = $1 AND user_id = $user_id;
SELECT * FROM insights WHERE constellation_id = $1 AND user_id = $user_id;
SELECT * FROM past_self_cards WHERE constellation_id = $1 AND user_id = $user_id;
SELECT * FROM topic_prompts WHERE constellation_id = $1 AND user_id = $user_id;
-- 全量原话（② 主题里说过的话）：
SELECT m.* FROM moments m
JOIN stars s ON s.trace_id = m.trace_id AND s.user_id = $user_id
JOIN constellation_stars cs ON cs.star_id = s.id
WHERE m.user_id = $user_id AND cs.constellation_id = $1;
-- past_self_card 和 topic_prompt 中引用的单个 Moment：
SELECT * FROM moments WHERE id = ANY($referenced_moment_ids) AND user_id = $user_id;
-- 后端组装 GetConstellationRes
```

### StartChat

```
新建：
  INSERT INTO chat_sessions (id, user_id, past_self_card_id, context_moment_ids, created_at)
  VALUES ($id, $user_id, $past_self_card_id, $context_moment_ids, NOW())
  AI 生成开场白 → 组装 opening ChatMessage

恢复：
  SELECT * FROM chat_messages WHERE session_id = $1 AND user_id = $user_id ORDER BY created_at;
  -- 返回 history + 最后一条 AI 消息作为 opening
```

### SendMessage

```
1. INSERT INTO chat_messages (id, user_id, session_id, role, content, created_at)
   VALUES ($id, $user_id, $session_id, 'user', $content, NOW())
2. AI 生成 past_self 回复（含 referenced_moments）
3. INSERT INTO chat_messages (id, user_id, session_id, role, content, referenced_moments, created_at)
   VALUES ($id, $user_id, $session_id, 'past_self', $content, $referenced_moments, NOW())
4. 返回 reply
```

---

## 4. 数据量估算

单用户 1 年：

| 表 | 预估量 | 说明 |
|---|--------|------|
| users | 1 | — |
| moments | ~1,000 | 每天 2-3 条 |
| stars | ~200 | 约 1/5 的 Trace 被寄存 |
| constellations | ~15-30 | 聚类后 |
| insights | ~15-30 | 每星座 1 条 |
| past_self_cards | ~30-60 | 每星座 2-3 张 |
| topic_prompts | ~30-60 | 每星座 2-3 条 |
| chat_sessions | ~50-100 | 偶尔使用 |
| chat_messages | ~500-1,000 | 每会话约 10 条 |

初期无需分区。moments 达 10 万级（用户数 × 单用户量）后按月分区：

```sql
CREATE TABLE moments (...) PARTITION BY RANGE (created_at);
CREATE TABLE moments_2026_05 PARTITION OF moments
  FOR VALUES FROM ('2026-05-01') TO ('2026-06-01');
```

---

## 5. pgvector 配置

```sql
CREATE EXTENSION vector;

-- embedding 维度取决于嵌入模型，建表时确定，后续可 ALTER
-- text-embedding-3-small (OpenAI): 1536
-- text-embedding-004 (Gemini):    768
```

向量索引策略：

| 阶段 | 数量 | 方案 |
|------|------|------|
| < 10K 条 | 不建索引 | 暴力搜索足够 |
| 10K-100K | ivfflat | `WITH (lists = 50)` |
| > 100K | hnsw | 构建慢，查询更快 |

---

## 6. Migration 顺序

```
1. CREATE EXTENSION vector
2. users
3. moments
4. constellations
5. stars
6. constellation_stars
7. insights
8. past_self_cards
9. topic_prompts
10. chat_sessions
11. chat_messages
12. INDEXES（向量索引最后建）
```

所有表无外键约束，按数据依赖顺序创建即可。
