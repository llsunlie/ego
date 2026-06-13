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
    DatabaseURL           string  // postgres://ego:ego@localhost:5432/ego?sslmode=disable
    JWTSecret             string
    WebPort               string  // 9080 (plain HTTP: gRPC-web + static files)
    WebTLSPort            string  // 9443 (TLS HTTP when TLS_DOMAIN set)
    GRPCPort              string  // 9444 (gRPC native, TLS when TLS_DOMAIN set)
    TLSDomain             string  // Let's Encrypt domain, empty = TLS disabled
    WebDir                string
    JWTExpHours           string
    LogLevel              string
    LogFormat             string
    AIAPIKey              string
    AIBaseURL             string
    AIEmbeddingModel      string
    AIEmbeddingAPIKey     string
    AIEmbeddingBaseURL    string
    AIChatModel           string
    AIChatAPIKey          string
    AIChatBaseURL         string
}
```

从 `.env` 文件加载（`loadEnvFile()` — 从 CWD 向上搜索 `.env`）。支持行内 `#` 注释。OS 环境变量优先级高于 `.env`。
另有 `RateLimitAuthRate/Burst/PreAuthRate/PreAuthBurst/MaxBuckets` 字段（见 ratelimit 节）。

## platform/postgres (`server/internal/platform/postgres/`)

```
postgres/
├── postgres.go            # pgxpool 连接（Connect 函数）
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

**配置**（`.env`，无默认值，不设则使用内部回退值 10/20/10/30/500）：
```
RATELIMIT_AUTH_RATE=10       # 鉴权接口 tokens/sec
RATELIMIT_AUTH_BURST=20      # 鉴权接口 桶容量
RATELIMIT_PREAUTH_RATE=10    # 免鉴权接口 tokens/sec
RATELIMIT_PREAUTH_BURST=30   # 免鉴权接口 桶容量
RATELIMIT_MAX_BUCKETS=500    # 最大桶对象数
```

**拦截器链顺序**（`bootstrap/server.go`）：`auth → ratelimit`。
ratelimit 在 auth 之后，利用 auth 注入的 `user_id` 做 per-phone 限流。
拦截器同时将客户端 IP 注入 request logger（`"ip"` 字段）。

## bootstrap (`server/internal/bootstrap/`)

应用启动时的依赖注入层：

```
bootstrap/
├── platform.go    # InitPlatform(cfg) — 初始化 Pool + Logger + AIClient + JWT
├── server.go      # NewServer(cfg, platform, handler) — 创建 gRPC server
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
