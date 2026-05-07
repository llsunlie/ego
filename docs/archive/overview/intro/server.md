# ego · 后端目录结构

> Go · gRPC · PostgreSQL + pgvector

## 1. 组织原则：按领域分包，非按层次分

```
       按层次分（低内聚）                  按领域分（高内聚）
       ─────────────────                  ─────────────────
       handler/moment.go                  writing/
       handler/chat.go                    ├── handler.go    ← 本领域全部 RPC 入口
       service/moment.go                  ├── service.go   ← 本领域全部业务逻辑
       service/chat.go                    └── db.go         ← 本领域全部 SQL
       db/moment.go
       db/chat.go                         chat/
                                          ├── handler.go
       ↑ 改一个功能要跨 3 层目录            ├── service.go
                                          └── db.go

                                          ↑ 改一个功能只动一个包
```

每个领域包只暴露 handler 需要的接口，内部实现（service、SQL）不对外暴露。

## 2. 目录树

```
server/
├── cmd/
│   └── ego/
│       └── main.go                       # 入口：加载配置 → 初始化 → 启动 gRPC
├── internal/
│   ├── config/
│   │   └── config.go                     # 环境变量/配置文件读取
│   ├── model/
│   │   └── model.go                      # 全量 DB 实体 struct，所有领域共享
│   ├── db/
│   │   ├── postgres.go                   # pgxpool 连接池初始化
│   │   └── migrations/                   # SQL 迁移文件
│   │       ├── 001_users.sql
│   │       ├── 002_moments.sql
│   │       ├── 003_constellations.sql
│   │       ├── 004_stars.sql
│   │       ├── 005_constellation_stars.sql
│   │       ├── 006_insights.sql
│   │       ├── 007_past_self_cards.sql
│   │       ├── 008_topic_prompts.sql
│   │       ├── 009_chat_sessions.sql
│   │       ├── 010_chat_messages.sql
│   │       └── 011_indexes.sql
│   ├── auth/
│   │   ├── jwt.go                        # JWT 签发 + 解析
│   │   └── interceptor.go                # gRPC unary interceptor：提取 JWT → 注入 ctx
│   ├── ai/
│   │   └── client.go                     # AI API 封装（Embed / Insight / ChatReply / …）
│   ├── login/
│   │   └── handler.go                    # Login RPC：查 users → bcrypt → 签发 JWT
│   ├── writing/                          # 写字 → 回声 → 观察 → 寄星
│   │   ├── handler.go                    # CreateMoment / GenerateInsight / StashTrace
│   │   ├── service.go                    # 业务逻辑编排
│   │   ├── db.go                         # SQL：moments + stars 的读写
│   │   └── echo.go                       # 向量搜索匹配回声（纯查询，不存库）
│   ├── timeline/                         # 时间线 + 记忆光点
│   │   ├── handler.go                    # ListMoments / GetRandomMoments
│   │   └── db.go                         # SQL：moments 只读查询
│   ├── starmap/                          # 星图 + 星座 + 聚类
│   │   ├── handler.go                    # ListConstellations / GetConstellation
│   │   ├── service.go                    # 详情组装逻辑
│   │   ├── db.go                         # SQL：constellations / stars / insights / cards / prompts
│   │   └── cluster.go                   # 异步聚类 + AI assets 生成（goroutine）
│   ├── chat/                             # 对话模式
│   │   ├── handler.go                    # StartChat / SendMessage
│   │   ├── service.go                    # 会话管理 + 上下文加载
│   │   └── db.go                         # SQL：chat_sessions / chat_messages
│   └── wire.go                           # 依赖注入：组装各领域 handler → gRPC 注册
├── proto/
│   └── ego/
│       └── api.proto                     # proto 源文件（前后端共享，实际放项目根目录）
├── go.mod
├── go.sum
└── Makefile                              # proto gen / migrate / run
```

## 3. 领域包内部结构

每个领域包统一采用三段式：

```
领域包/
├── handler.go    # gRPC 入口：参数校验 → 调 service → 组装 pb 响应（薄层）
├── service.go    # 业务逻辑编排，调用本包的 db.go 和 ai/client.go
└── db.go         # 本领域 SQL（pgx 查询），输入/输出均为 model.go 中的 struct
```

**约束：**
- **handler** 不写业务逻辑，不直接调 `db.go`
- **service** 不写 SQL 字符串，只调 `db.go` 暴露的函数
- **db.go** 不做业务判断，只执行 SQL 返回 struct
- 跨包只能调对方包暴露的 `service` 层函数，不能直接调 `db.go`

## 4. 表与领域包映射

```
领域包        写入表                       读取表（跨包只读）
────────────────────────────────────────────────────────────
login         users                       —
writing       moments, stars              —
timeline      —                           moments
starmap       constellations,             stars, moments
              constellation_stars,
              insights,
              past_self_cards,
              topic_prompts
chat          chat_sessions,              past_self_cards, moments
              chat_messages
```

- **写入权单一**：每张表只有一个包执行 INSERT/UPDATE/DELETE
- **读取可跨包**：starmap 读 moments、chat 读 past_self_cards 都是只读查询，无需经过 writing 包中转

## 5. 跨包依赖图

```
                          ┌──────────┐
                          │  config  │
                          └────┬─────┘
                               │
                    ┌──────────┼──────────┐
                    ▼          ▼          ▼
              ┌────────┐ ┌────────┐ ┌────────┐
              │  db    │ │  ai    │ │  auth  │  ← 基础设施（无业务依赖）
              └───┬────┘ └───┬────┘ └───┬────┘
                  │          │          │
     ┌────────────┼──────────┼──────────┼────────────────┐
     ▼            ▼          ▼          ▼                ▼
┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐    ┌─────────┐
│  login  │ │ writing │ │timeline │ │ starmap │    │  chat   │
└─────────┘ └────┬─────┘ └─────────┘ └────┬─────┘    └─────────┘
                 │                        │
                 │  go s.clusterStar()    │
                 └────────────────────────┘
              （writing 异步触发 starmap 聚类）
```

- 领域包之间无直接 import，通过 goroutine 异步解耦
- 所有包依赖 `model`（共享 struct 定义），不形成循环
- `auth` interceptor 在 gRPC 层注入，领域包无需显式调用

## 6. wire.go 依赖注入示意

```go
func InitializeServer(cfg *config.Config) *grpc.Server {
    pool := db.Connect(cfg.DatabaseURL)
    aiClient := ai.NewClient(cfg.AIKey)
    jwtService := auth.NewJWTService(cfg.JWTSecret)

    // 按依赖顺序组装
    loginHandler := login.NewHandler(pool, jwtService)
    writingHandler := writing.NewHandler(pool, aiClient)
    timelineHandler := timeline.NewHandler(pool)
    starmapHandler := starmap.NewHandler(pool)
    chatHandler := chat.NewHandler(pool, aiClient)

    // writing 的 StashTrace 需要触发 starmap 聚类
    writingHandler.OnStash = starmapHandler.ClusterStarAsync

    return grpc.NewServer(
        grpc.UnaryInterceptor(auth.NewInterceptor(jwtService)),
        // 注册各 handler ...
    )
}
```

## 7. 关键 Go 依赖

```
用途              包
─────────────────────────────────────────
gRPC             google.golang.org/grpc
Proto            google.golang.org/protobuf
JWT              github.com/golang-jwt/jwt/v5
bcrypt           golang.org/x/crypto/bcrypt
pgx              github.com/jackc/pgx/v5  （pgxpool 连接池）
pgvector         github.com/pgvector/pgvector-go
errgroup         golang.org/x/sync/errgroup
UUID             github.com/google/uuid
```

## 8. main.go 启动流程

```
1. config.Load()           → 读取配置
2. db.Connect(cfg)         → pgxpool 连接 PostgreSQL
3. ai.NewClient(cfg)       → 初始化 AI API 客户端
4. auth.NewJWTService(cfg) → JWT 签发/校验器
5. InitializeServer(cfg)   → wire.go 组装所有 handler
6. server.Serve(lis)       → 启动监听
```
