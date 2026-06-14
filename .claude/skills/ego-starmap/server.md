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
│   ├── constellation_matcher.go             # AI 星座匹配逻辑
│   ├── constellation_asset_generator.go     # 生成星座名/描述
│   └── topic_generator.go                   # 生成话题提示
└── adapter/
    ├── grpc/handler.go                      # gRPC Handler
    ├── grpc/mapper.go                       # proto ↔ domain 映射
    ├── id/uuid.go                           # UUID 生成
    ├── postgres/
    │   ├── star_repo.go                     # StarRepository
    │   ├── star_reader.go                   # StarReader（跨模块）
    │   ├── constellation_repo.go            # ConstellationRepository
    │   └── trace_stasher.go                 # TraceStasher — 标记 trace 为已 stash
    └── ai/
        ├── constellation_matcher.go         # AI 判断 moment 归属哪个星座
        ├── constellation_asset_generator.go # AI 生成星座名称和描述
        └── topic_generator.go               # AI 生成话题提示
```

## 核心领域模型 (`domain/types.go`)

```go
Star          { ID, ConstellationID, MomentID, Brightness, CreatedAt }
Constellation { ID, UserID, Name, Description, StarCount, CreatedAt }
```

## StashTrace 用例（核心流程）(`app/stash_trace.go`)

1. 查询 trace 的所有 moments
2. 标记 trace 为 `stashed = true`（通过 `TraceStasher`）
3. **AI 星座匹配**：判断每个 moment 应归属哪个星座
4. **同名星座合并**：同名星座下的 stars 合并
5. **新星座生成**：为新主题的 moment 创建新星座 + AI 生成名称/描述
6. 创建 Star 实体关联 moment

## ListConstellations / GetConstellation 用例

- `ListConstellations`: 返回所有星座 + 总 star 数
- `GetConstellation`: 返回星座详情 + 关联的 moments + stars

## 跨模块依赖

```go
type Deps struct {
    DB       sqlc.DBTX
    AIClient *platformai.Client
}
```

依赖：
- `writing/adapter/postgres` — `Reader`（MomentReader, TraceReader）
- `writing/adapter/postgres` — `EchoRepository`, `InsightRepository`
- `platform/ai` — AI Client（用于星座匹配、名称生成、话题生成）

## Profile 向量召回

星座聚合已升级为 TraceProfile -> ConstellationProfile 的画像匹配。画像向量使用全局 embedding 配置，当前默认 `BAAI/bge-m3` / `AI_EMBEDDING_DIM=1024`。

| 表 | 向量列 | 维度 | 索引 |
|---|---|---:|---|
| `trace_profile_vectors` | `embedding` | 1024 | `idx_trace_profile_vectors_embedding_hnsw` |
| `constellation_profile_vectors` | `profile_embedding` | 1024 | `idx_constellation_profile_vectors_profile_embedding_hnsw` |
| `constellation_profile_vectors` | `centroid_embedding` | 1024 | 暂不做 topK 召回 |

候选召回时以 ConstellationProfile 的 `profile_embedding` 做 pgvector/HNSW dense topK，同时可用 Elasticsearch sparse 召回补足中文短语、标签和场景词命中，再通过 RRF 融合候选。`centroid_embedding` 仍维护为实际 TraceProfile embedding 的加权平均，当前主要用于画像统计和后续评分扩展。

## 相关文件

| 文件 | 说明 |
|------|------|
| `server/internal/platform/ai/client.go` | AI API 客户端 |
| `server/internal/writing/adapter/postgres/reader.go` | Moment/Trace Reader |
| `server/internal/starmap/adapter/postgres/trace_profile_repo.go` | TraceProfile 持久化 + vector 写入 |
| `server/internal/starmap/adapter/postgres/constellation_profile_repo.go` | ConstellationProfile 持久化 + pgvector 候选召回 |
| `server/internal/platform/postgres/migrations/011_trace_profiles.sql` | TraceProfile 迁移，`VECTOR(1024)` + HNSW |
| `server/internal/platform/postgres/migrations/012_constellation_profiles.sql` | ConstellationProfile 迁移，`VECTOR(1024)` + HNSW |
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
        └── chat_generator.go           # LLM Chat 生成实现
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
