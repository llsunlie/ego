---
## starmap 领域

---

# ego-cli-server-starmap

starmap 有界上下文 — 星座可视化系统。DDD 结构：`server/internal/starmap/`

## 所属 gRPC 方法

- `ListConstellations` — 获取所有星座列表（含 starCount）
- `GetConstellation` — 获取单个星座详情（含 moments, stars）
- `StashTrace` — 将 trace 收进星图（触发 AI 星座匹配）

## 模块结构 (`server/internal/starmap/`)

```
starmap/
├── module.go                                # 依赖注入
├── domain/
│   ├── types.go                             # Star, Constellation 领域类型
│   ├── ports.go                             # Repository 接口
│   └── errors.go                            # 领域错误
├── app/
│   ├── stash_trace.go                       # StashTrace 用例（核心）
│   ├── list_constellations.go               # ListConstellations 用例
│   ├── get_constellation.go                 # GetConstellation 用例
│   └── stash_trace_test.go                  # 异步聚类/兼容路径测试
└── adapter/
    ├── grpc/handler.go                      # gRPC Handler
    ├── grpc/mapper.go                       # proto ↔ domain 映射
    ├── id/uuid.go                           # UUID 生成
    ├── postgres/
    │   ├── star_repo.go                     # StarRepository
    │   ├── star_reader.go                   # StarReader（跨模块）
    │   ├── constellation_repo.go            # ConstellationRepository
    │   ├── trace_profile_repo.go            # TraceProfile + vector 持久化
    │   ├── constellation_profile_repo.go    # ConstellationProfile + membership 持久化
    │   ├── constellation_sparse_search.go   # pg_trgm 星座 Profile 稀疏召回
    │   └── trace_stasher.go                 # TraceStasher — 标记 trace 为已 stash
    └── ai/
        ├── trace_profile_generator.go       # 从 Trace/Moments 生成结构化 Profile + 向量
        ├── constellation_asset_generator.go # AI 生成星座名称和描述
        ├── constellation_borderline_judge.go # 边界归属判断
        ├── constellation_profile_refiner.go # 星座成熟节点 Profile 重写
        ├── json_repair.go                   # AI JSON 输出修复
        └── debug_log.go                     # AI prompt/response 调试日志
```

## 核心领域模型 (`domain/types.go`)

```go
Star          { ID, UserID, TraceID, Topic, CreatedAt }
Constellation { ID, UserID, Topic, TopicEmbedding, Name, ConstellationInsight, StarIDs, TopicPrompts, CreatedAt, UpdatedAt }
TraceProfile  { TraceID, UserID, Topic, Summary, Keywords, Emotions, Scenes, CentralPattern, PatternTags, ProfileText, Status, ... }
ConstellationProfile { ConstellationID, UserID, Topic, Summary, Keywords, Emotions, Scenes, CentralPattern, PatternTags, ThemeCode, ProfileText, TraceCount, MomentCount, ... }
ConstellationMembership { ConstellationID, StarID, TraceID, MatchScore, MatchType, MatchDimensions, Weight, CreatedAt }
```

## StashTrace 用例（核心流程）(`app/stash_trace.go`)

1. 查询 trace 的所有 moments
2. 创建 Star 占位（topic 初始为「聚合中」），标记 trace 为 `stashed = true`
3. 后台异步生成 `TraceProfile` + `TraceProfileVector`，持久化到 `trace_profiles` / `trace_profile_vectors`
4. 候选召回：
   - dense: `ConstellationProfileRepository.FindCandidates` 通过 pgvector profile embedding 最近邻召回
   - sparse: `ConstellationSparseSearch` 通过 `constellation_profiles.search_text` + pg_trgm 召回
   - RRF 融合 dense/sparse 候选，再按画像相似度、centroid、关键词/场景/情绪/模式标签 overlap 重新打分
5. 主归属决策：
   - 强匹配直接 attach 到已有星座
   - 边界候选交给 `ConstellationBorderlineJudge`
   - 无可用匹配时创建新星座和初始 `ConstellationProfile`
6. 可附加最多 2 个 secondary 星座 membership；成熟节点（3/5/8/13/之后每 8）可触发 `ConstellationProfileRefiner`
7. 旧构造路径缺少 profile 依赖时，会走 legacy asset 生成路径，保持历史测试和调用兼容

## ListConstellations / GetConstellation 用例

- `ListConstellations`: 返回所有星座 + 总 star 数，并优先用 `constellation_profiles.theme_label/theme_code` 丰富展示字段
- `GetConstellation`: 返回星座详情 + 关联的 moments + stars

## 跨模块依赖

```go
type Deps struct {
    DB                      sqlc.DBTX
    AIClient                *platformai.Client
    AIEmbeddingDim          int
    ConstellationSparseOn   bool
    ConstellationSparseTopK int
    ConstellationHybridRRFK int
}
```

依赖：
- `writing/adapter/postgres` — `Reader`（MomentReader, TraceReader）
- `writing/adapter/postgres` — `EchoRepository`, `InsightRepository`
- `platform/ai` — AI Client（用于 profile 生成、边界判断、profile refine、星座资产生成）
- `platform/postgres` — pgvector + pg_trgm 支撑 dense/sparse 星座召回

## 相关文件

| 文件 | 说明 |
|------|------|
| `server/internal/platform/ai/client.go` | AI API 客户端 |
| `server/internal/platform/ai/retry.go` | AI Chat 重试策略 |
| `server/internal/platform/postgres/migrations/012_constellation_profiles.sql` | trace/constellation profile 与 vector 表 |
| `server/internal/platform/postgres/migrations/013_profile_pattern_tags.sql` | pattern_tags 字段 |
| `server/internal/platform/postgres/migrations/014_constellation_theme_codebook.sql` | theme codebook 字段 |
| `server/internal/platform/postgres/migrations/015_pgtrgm_search.sql` | pg_trgm search_text 与 GIN 索引 |
| `server/internal/writing/adapter/postgres/reader.go` | Moment/Trace Reader |
| `server/internal/platform/postgres/sqlc/stars.sql.go` | sqlc star 查询 |
| `server/internal/platform/postgres/sqlc/constellations.sql.go` | sqlc constellation 查询 |
| `server/internal/bootstrap/starmap.go` | 顶层 wiring |

---
## conversation 领域

---

# ego-cli-server-conversation

conversation 有界上下文 — 星图内的 AI 对话。DDD 结构：`server/internal/conversation/`

## 所属 gRPC 方法

- `StartChat` — 创建聊天会话（基于 Star 或已有 Session）
- `SendMessage` — 发送消息并获取 AI 回复

## 模块结构 (`server/internal/conversation/`)

```
conversation/
├── module.go                           # 依赖注入
├── domain/
│   ├── types.go                        # ChatSession, Message 领域类型
│   ├── ports.go                        # Repository 接口
│   └── errors.go                       # 领域错误
├── app/
│   ├── start_chat.go                   # StartChat 用例
│   ├── send_message.go                 # SendMessage 用例
│   ├── chat_generator.go               # AI 对话生成接口
│   └── ports.go                        # IDGenerator, ChatGenerator 接口
└── adapter/
    ├── grpc/handler.go                 # gRPC Handler
    ├── grpc/mapper.go                  # proto ↔ domain 映射
    ├── id/uuid.go                      # UUID 生成
    ├── postgres/
    │   ├── session_repo.go             # SessionRepository
    │   └── message_repo.go             # MessageRepository
    └── ai/
        └── chat_generator.go           # LLM Chat 生成实现 + 引用 snippet 截断
```

## 核心领域模型 (`domain/types.go`)

```go
ChatSession { ID, StarID, UserID, CreatedAt }
Message     { ID, SessionID, Role, Content, CreatedAt }
```

## StartChat 用例 (`app/start_chat.go`)

- 输入 `starId` + 可选 `chatSessionId`
- `chatSessionId` 为空 → 创建新 Session（加载 star 关联的 moment 作为上下文）
- `chatSessionId` 非空 → 继续已有对话
- 返回 `chatSessionId` + 开场白（AI 生成）
- `adapter/ai/chat_generator.go` 会把返回的 MomentReference snippet 截断到 30 个 rune，避免前端引用区域过长

## SendMessage 用例 (`app/send_message.go`)

- 加载历史消息作为对话上下文
- 调用 AI Chat API 生成回复
- 保存双方消息到数据库

## 跨模块依赖

```go
type Deps struct {
    DB       sqlc.DBTX
    AIClient *platformai.Client
}
```

依赖 `starmap/adapter/postgres` 的 `StarReader`（跨模块读 star 信息），以及 `writing/adapter/postgres` 的 `ChatMomentReader`（读取 moment 内容作为对话上下文）。

## 相关文件

| 文件 | 说明 |
|------|------|
| `server/internal/platform/ai/client.go` | AI Chat API 客户端 |
| `server/internal/platform/postgres/sqlc/chat_sessions.sql.go` | sqlc session 查询 |
| `server/internal/platform/postgres/sqlc/chat_messages.sql.go` | sqlc message 查询 |
| `server/internal/starmap/adapter/postgres/star_reader.go` | StarReader 实现 |
| `server/internal/writing/adapter/postgres/reader.go` | ChatMomentReader |
| `server/internal/bootstrap/chat.go` | 顶层 wiring |
