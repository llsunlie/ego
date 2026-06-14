---
name: ego-svr-platform
description: 服务端基础设施 context — Postgres 连接、AI 客户端、JWT 认证、日志、配置、Bootstrap 组装、sqlc 数据层。支撑所有前端 client page。
---

# ego-cli-server-platform

基础设施层 — 数据库、认证、AI、日志、配置、启动组装。支撑所有有界上下文的底层能力。

## 涵盖的目录

```
server/internal/
├── config/              # 配置加载
├── platform/            # 基础设施
│   ├── postgres/        # 数据库连接 + sqlc 生成代码
│   ├── ai/              # AI API 客户端
│   ├── auth/            # JWT + bcrypt + gRPC 拦截器
│   ├── logging/         # 结构化日志
│   ├── metrics/         # Prometheus 指标
│   └── ratelimit/       # API 令牌桶限流（per-IP + per-user）
├── bootstrap/           # 启动组装（依赖注入）
└── shared/              # 共享工具
```

## config (`server/internal/config/`)

```go
type Config struct {
    DatabaseURL                string  // postgres://ego:ego@localhost:5432/ego?sslmode=disable
    JWTSecret                  string
    WebPort                    string  // 9080 (plain HTTP: gRPC-web + static files)
    WebTLSPort                 string  // 9443 (TLS HTTP when TLS_DOMAIN set)
    GRPCPort                   string  // 9444 (gRPC native, TLS when TLS_DOMAIN set)
    TLSDomain                  string  // Let's Encrypt domain, empty = TLS disabled
    CORSAllowedOrigins         string  // CORS_ALLOWED_ORIGINS, comma-separated whitelist
    WebDir                     string
    JWTExpHours                string
    LogLevel                   string
    LogFormat                  string
    LogOutput                  string
    AIAPIKey                   string
    AIBaseURL                  string
    AIEmbeddingModel           string  // default model family: BAAI/bge-m3
    AIEmbeddingDim             string  // default 1024
    AIEmbeddingAPIKey          string
    AIEmbeddingBaseURL         string
    AIChatModel                string
    AIChatAPIKey               string
    AIChatBaseURL              string
    EchoRecallTopK             string
    ElasticsearchURL           string
    ElasticsearchUser          string
    ElasticsearchPass          string
    EchoSparseEnabled          string
    EchoSparseTopK             string
    EchoHybridRRFK             string
    ConstellationSparseEnabled string
    ConstellationSparseTopK    string
    ConstellationHybridRRFK    string
}
```

从 `.env` 文件加载（`loadEnvFile()` — 从 CWD 向上搜索 `.env`）。支持行内 `#` 注释。OS 环境变量优先级高于 `.env`。
另有 `RateLimitAuthRate/Burst/PreAuthRate/PreAuthBurst/MaxBuckets` 字段（见 ratelimit 节）。

当前 embedding 默认使用 `BAAI/bge-m3`，输出维度通过 `AI_EMBEDDING_DIM` 配置，默认 `1024`。数据库迁移中的 pgvector 列已固定为 `VECTOR(1024)`，修改模型维度时必须同步迁移：

- `moment_embedding_vectors.embedding`
- `trace_profile_vectors.embedding`
- `constellation_profile_vectors.profile_embedding`
- `constellation_profile_vectors.centroid_embedding`

Echo 和星座画像召回均支持 dense pgvector + Elasticsearch sparse 两路召回，再使用 RRF 融合候选。相关开关和 topK 在 `ECHO_*`、`CONSTELLATION_*` 环境变量中配置。

`CORS_ALLOWED_ORIGINS` 是生产跨域白名单。`TLS_DOMAIN` 为空的本地开发模式会自动允许 `localhost` / `127.0.0.1` origin；生产模式下空白名单会拒绝跨域请求。

## platform/postgres (`server/internal/platform/postgres/`)

```
postgres/
├── postgres.go            # pgxpool 连接（Connect 函数）
├── migrations/            # schema 迁移，含 pgvector/HNSW 向量表
└── sqlc/                  # sqlc 生成的类型安全 SQL 代码
    ├── db.go              # DBTX 接口 + Queries 结构体
    ├── models.go          # sqlc 数据模型
    ├── users.sql.go       # User 查询
    ├── traces.sql.go      # Trace 查询
    ├── moments.sql.go     # Moment 查询
    ├── echos.sql.go       # Echo 查询
    ├── insights.sql.go    # Insight 查询
    ├── stars.sql.go       # Star 查询
    ├── constellations.sql.go # Constellation 查询
    ├── chat_sessions.sql.go  # ChatSession 查询
    └── chat_messages.sql.go  # ChatMessage 查询
```

`Connect(databaseURL)` → 创建 `*pgxpool.Pool`，Ping 验证连通性。

向量检索相关迁移：

| Migration | 表 | 向量列 | 索引 |
|---|---|---|---|
| `010_moment_embedding_vectors.sql` | `moment_embedding_vectors` | `embedding VECTOR(1024)` | `idx_moment_embedding_vectors_embedding_hnsw` |
| `011_trace_profiles.sql` | `trace_profile_vectors` | `embedding VECTOR(1024)` | `idx_trace_profile_vectors_embedding_hnsw` |
| `012_constellation_profiles.sql` | `constellation_profile_vectors` | `profile_embedding VECTOR(1024)`, `centroid_embedding VECTOR(1024)` | `idx_constellation_profile_vectors_profile_embedding_hnsw` |

## platform/ai (`server/internal/platform/ai/`)

```
ai/
├── config.go        # AI 配置（Embedding + Chat 分离配置）
├── client.go        # 统一 AI 客户端（嵌入 + 对话 + 相似度）
└── similarity.go    # 余弦相似度计算
```

`Client` 提供：
- `Embed(ctx, texts)` — 文本向量嵌入
- `Chat(ctx, messages)` — LLM 对话
- `Similarity(a, b)` — 余弦相似度

Embedding 维度由 `platform.Config.EmbeddingDim` 校验并向下游写库链路传递。当前默认模型为 `BAAI/bge-m3`，因此迁移和回填脚本都以 1024 维为准。

## platform/auth (`server/internal/platform/auth/`)

```
auth/
├── bcrypt.go        # bcrypt 密码哈希（实现 identity 的 PasswordHasher）
├── jwt_issuer.go    # JWT 签发（实现 identity 的 TokenIssuer）
├── jwt.go           # JWT 验证 + gRPC 一元拦截器
└── interceptor.go   # gRPC auth 拦截器（提取 token → 注入 context）
```

## platform/logging (`server/internal/platform/logging/`)

```go
type Config struct { Level, Format, OutputPath string }
func New(cfg Config) (*slog.Logger, error)
```

- 支持 `json` / `text` 格式
- Context-based 日志（`logging.FromContext(ctx)`）
- 通过 gRPC 拦截器将 logger 注入 context

## platform/metrics (`server/internal/platform/metrics/`)

Prometheus 指标注册和 HTTP handler。

## platform/ratelimit (`server/internal/platform/ratelimit/`)

```
ratelimit/
├── ratelimit.go         # Limiter — 令牌桶管理 + atomic 计数 + 后台清理
├── ratelimit_test.go    # 单元测试
└── interceptor.go       # gRPC UnaryServerInterceptor + IP 提取
```

**限流策略**：
- preAuth RPC（Login/CheckPhone/SendVerificationCode/Register/ResetPassword）→ 仅按 IP 限流
- 鉴权 RPC → per-IP + per-user_id 双维度独立令牌桶，任一耗尽即拒绝
- 拒绝返回 `RESOURCE_EXHAUSTED`，message 为中文「请求过于频繁，请稍后再试」

**配置**（`.env`，无默认值，不设则使用内部回退值）：

| 参数 | 回退值 | 说明 |
|------|--------|------|
| `RATELIMIT_AUTH_RATE` | 10 | 鉴权接口 tokens/sec |
| `RATELIMIT_AUTH_BURST` | 20 | 鉴权接口 桶容量 |
| `RATELIMIT_PREAUTH_RATE` | 10 | 免鉴权接口 tokens/sec |
| `RATELIMIT_PREAUTH_BURST` | 30 | 免鉴权接口 桶容量 |
| `RATELIMIT_MAX_BUCKETS` | 500 | 最大桶对象数，超限 fail-open |
| `RATELIMIT_CLEANUP_INTERVAL` | 60 | 桶清理间隔（秒） |
| `RATELIMIT_BUCKET_TTL` | 300 | 空闲桶过期时间（秒） |

**日志**（结构化 slog）：

| 事件 | 级别 | 内容 |
|------|------|------|
| 启动 | INFO | `ratelimit started` — 全部 config 参数 |
| 限流拒绝 | WARN | `ratelimit denied` — method, ip, user_id, dim |
| 桶超限放行 | WARN | `ratelimit fail-open` — max, current, key |
| 定时清理 | INFO | `ratelimit cleanup` — removed, remaining（仅清理数 >0） |

**拦截器链顺序**（`bootstrap/server.go`）：`auth → ratelimit`。
ratelimit 在 auth 之后，利用 auth 注入的 `user_id` 做 per-phone 限流。
拦截器同时将客户端 IP 注入 request logger（`"ip"` 字段）。

## bootstrap (`server/internal/bootstrap/`)

应用启动时的依赖注入层：

```
bootstrap/
├── platform.go    # InitPlatform(cfg) — 初始化 Pool + Logger + AIClient + JWT
├── server.go      # NewServer(cfg, platform, handler) — 创建 gRPC/gRPC-web server + CORS whitelist
├── composite.go   # EgoHandler — 组合所有 module handler，按 RPC 方法路由
├── identity.go    # NewIdentityHandler(platform)
├── writing.go     # NewWritingHandler(platform)
├── timeline.go    # NewTimelineHandler(platform)
├── starmap.go     # NewStarmapHandler(platform)
└── chat.go        # NewChatHandler(platform)
```

**启动流程** (`cmd/ego/main.go`):
1. `config.Load()` — 加载配置
2. `bootstrap.InitPlatform(cfg)` — 创建 Pool + Logger + AIClient + JWT
3. `bootstrap.NewXxxHandler(p)` — 创建各 module handler
4. `bootstrap.NewEgoHandler(...)` — 组合为 composite handler
5. `bootstrap.NewServer(cfg, p, handler)` — 创建 gRPC server
6. `server.Serve()` — 启动 gRPC + gRPC-web

`server.go` 使用 `grpcweb.WithOriginFunc(makeOriginChecker(cfg))` 做跨域检查：

- same-origin 请求始终允许。
- `TLS_DOMAIN` 为空时允许本地开发 origin。
- 其他跨域请求必须命中 `CORS_ALLOWED_ORIGINS`。

## DDD 架构约定

每个业务 domain 遵循统一的三层结构：
```
<domain>/
├── module.go       # Wire 函数：接收 Deps，返回 Handler
├── domain/         # 领域类型 + 接口定义（无外部依赖）
│   ├── types.go    # 实体/聚合根
│   ├── ports.go    # Repository 接口
│   └── errors.go   # 领域错误
├── app/            # 应用服务/用例（依赖 domain 接口）
│   ├── ports.go    # 应用层接口（IDGenerator 等）
│   └── <usecase>.go
└── adapter/        # 基础设施实现（protobuf handler, postgres repo, AI client）
    ├── grpc/       # gRPC handler + mapper
    ├── postgres/   # Repository 实现
    ├── ai/         # AI 适配器
    └── id/         # ID 生成器
```
