# ego 后端具体架构与目录结构决策

本文记录基于 `docs/architecture/backend-ddd.md` 和 `docs/overview/intro/server.md` 的后端结构讨论结论，并作为后续调整 `server/` 目录的依据。

---

## 1. 结论

后端采用：

```text
模块化单体
+ DDD 限界上下文
+ 六边形端口适配器
+ platform 基础设施层
+ gRPC API
+ PostgreSQL/pgvector
```

一句话：

> `overview/intro/server.md` 的“按领域分包”方向是对的，但 `handler.go / service.go / db.go` 三件套粒度偏粗；最终目录应采用 `domain / app / adapter / platform`，以适配多人按模块 owner 开发。

---

## 2. 两份文档的取舍

### 2.1 `docs/overview/intro/server.md` 的价值

它适合早期快速落地，优点是：

- 反对按技术层横切目录。
- 强调按领域分包。
- 每个领域内部有 handler、service、db。
- 能让一个小团队快速写出 MVP。

但它的问题是：

- `internal/model/model.go` 容易变成全局上帝模型。
- `db.go` 直接在领域包里，业务与持久化细节容易混在一起。
- `writing` 同时包含 `StashTrace`，边界不够清晰。
- `login`、`auth`、`ai`、`db` 没有区分业务上下文和基础设施能力。
- 多人按模块 owner 开发时，缺少明确的模块契约和隔离层。

### 2.2 `backend-ddd.md` 的价值

它更适合长期结构，优点是：

- 以业务上下文划分模块。
- 每个模块内部区分 `domain`、`app`、`adapter`。
- 基础设施统一放 `platform`。
- proto、domain model、persistence model 分离。
- 表写入权归属清晰。
- 更适合分层 Harness 和多人协作。

### 2.3 最终选择

采用 `backend-ddd.md` 的结构作为目标架构。

`docs/overview/intro/server.md` 保留为早期草案参考，不作为最终目录规范。

---

## 3. 后端目标目录

```text
server/
├── AGENTS.md
├── Makefile
├── go.mod
├── go.sum
├── sqlc.yaml
├── .harness/
│   ├── backend-feature-index.json
│   ├── integration-progress.md
│   └── clean-state-checklist.md
│
├── cmd/
│   ├── ego/
│   │   └── main.go
│   └── migrate/
│       └── main.go
│
├── proto/
│   └── ego/
│       ├── api.pb.go
│       └── api_grpc.pb.go
│
└── internal/
    ├── bootstrap/
    ├── config/
    ├── shared/
    ├── platform/
    ├── identity/
    ├── writing/
    ├── timeline/
    ├── starmap/
    └── conversation/
```

---

## 4. 目录职责

### 4.1 `cmd/`

进程入口。

```text
cmd/
├── ego/
│   └── main.go
└── migrate/
    └── main.go
```

规则：

- 只做启动、配置加载和依赖组装调用。
- 不写业务逻辑。
- 不直接写 SQL。

### 4.2 `internal/bootstrap/`

依赖组装层。

职责：

- 创建 DB pool。
- 创建 platform 服务。
- 创建各业务模块 usecase 和 gRPC handler。
- 注册 gRPC server。

它是原 `wire.go` 的升级位置。

### 4.3 `internal/config/`

配置读取。

职责：

- 读取环境变量。
- 提供配置结构。

不负责：

- 初始化外部连接。
- 业务默认值判断。

### 4.4 `internal/shared/`

极少量共享类型。

可放：

- 通用 ID 类型。
- 通用领域错误。
- 领域事件接口。
- Clock 接口。
- 事务接口。

不放：

- `Moment`
- `Star`
- `ChatSession`
- 全量 DB model

业务对象必须回到自己的 bounded context。

### 4.5 `internal/platform/`

基础设施层。

```text
internal/platform/
├── postgres/
├── grpc/
├── auth/
├── ai/
└── eventbus/
```

职责：

- PostgreSQL 连接、迁移、sqlc。
- JWT、bcrypt。
- gRPC server/interceptor/error mapper。
- AI SDK、prompt、输出校验。
- event bus / outbox。

规则：

- platform 不表达 ego 业务流程。
- platform 可以实现业务模块声明的 interface。
- 业务模块不直接依赖具体 AI SDK 或数据库连接细节。

### 4.6 `internal/identity/`

身份上下文。

负责：

- Login。
- 自动注册。
- User。
- 密码校验编排。
- JWT 签发编排。

拥有写入：

- `users`

### 4.7 `internal/writing/`

此刻写作上下文。

负责：

- Trace。
- Moment。
- Echo。
- 当前体验中的实时 Insight。

拥有写入：

- `moments`
- 未来可新增 `traces`

不负责：

- Star。
- Constellation。
- ChatSession。

### 4.8 `internal/timeline/`

过往查询上下文。

负责：

- 过往列表。
- 记忆光点随机读取。

性质：

- 查询模块。
- 不拥有 Moment 写入权。

### 4.9 `internal/starmap/`

星图上下文。

负责：

- StashTrace。
- Star。
- Constellation。
- 星座级 Insight。
- PastSelfCard。
- TopicPrompt。
- 星座详情组装。

拥有写入：

- `stars`
- `constellations`
- `constellation_stars`
- `insights`
- `past_self_cards`
- `topic_prompts`

### 4.10 `internal/conversation/`

对话上下文。

负责：

- ChatSession。
- ChatMessage。
- 和过去自己的对话。
- AI 回复引用来源校验。

拥有写入：

- `chat_sessions`
- `chat_messages`

---

## 5. 业务模块内部结构

每个业务模块统一采用：

```text
internal/{module}/
├── AGENTS.md
├── ARCHITECTURE.md
├── CONTRACT.md
├── .harness/
│   ├── feature_list.json
│   ├── progress.md
│   └── clean-state-checklist.md
├── domain/
├── app/
└── adapter/
    ├── grpc/
    └── postgres/
```

### 5.1 `domain/`

业务核心。

可包含：

- Entity。
- Value Object。
- Aggregate。
- Repository interface。
- Domain Service。
- Domain Event。
- 业务错误。

禁止依赖：

- proto。
- pgx/sqlc。
- grpc/status。
- platform/ai。
- config。

### 5.2 `app/`

用例编排层。

可包含：

- LoginUseCase。
- CreateMomentUseCase。
- StashTraceUseCase。
- SendMessageUseCase。
- 查询 usecase。
- 事务编排。
- 调用 domain port。

不写：

- SQL 字符串。
- 具体 AI SDK 调用。
- proto 请求响应。

### 5.3 `adapter/grpc/`

gRPC 输入适配器。

职责：

- 接收 proto request。
- 参数校验。
- 转换为 command/query。
- 调用 app usecase。
- 转换为 proto response。

### 5.4 `adapter/postgres/`

PostgreSQL 输出适配器。

职责：

- 实现 domain repository interface。
- 调用 sqlc/pgx。
- 将 persistence model 转成 domain model。

---

## 6. 不采用全局 `model/model.go`

`docs/overview/intro/server.md` 中的 `internal/model/model.go` 不作为最终方案。

原因：

- 它会变成所有模块都依赖的共享模型。
- 表字段变化会影响不相关模块。
- 容易让 domain model 和 DB model 混在一起。
- 不利于模块 owner 独立开发。

最终规则：

```text
proto model       只在 adapter/grpc 出现
domain model      放 internal/{module}/domain
persistence model 放 internal/{module}/adapter/postgres 或 platform/postgres/sqlc
```

---

## 7. 后端模块间协作

后端内部模块之间不使用 proto 作为领域模型。

推荐协作方式：

- Go interface。
- domain port。
- application interface。
- read model。
- domain event。

例如：

```text
writing/domain/ports.go
  定义 EmbeddingProvider

platform/ai
  实现 EmbeddingProvider

bootstrap/wire.go
  注入给 writing usecase
```

模块间默认只读取对方：

- `ARCHITECTURE.md`
- `CONTRACT.md`
- 公开 interface

不直接读取或修改对方内部实现。

---

## 8. 迁移策略

当前已有结构：

```text
internal/auth
internal/config
internal/db
internal/login
```

迁移目标：

```text
internal/platform/auth
internal/platform/postgres
internal/identity
```

迁移步骤：

1. `internal/auth` 移到 `internal/platform/auth`。
2. `internal/db` 移到 `internal/platform/postgres`。
3. `internal/login` 移到 `internal/identity/adapter/grpc`。
4. 更新 imports。
5. 新建 `internal/bootstrap` 骨架。
6. 新建 `writing/timeline/starmap/conversation` 骨架。
7. 跑 `go test ./...`。

---

## 9. 最终判断

后端目录结构应该服务于两个目标：

1. 让代码表达 ego 的业务边界。
2. 让不同开发者和 agent 能按模块独立工作。

因此最终方案不是按 `handler/service/db` 横切，也不是用一个全局 model 包串起来，而是：

> 每个业务上下文拥有自己的 domain/app/adapter；platform 提供基础设施能力；bootstrap 负责依赖组装；proto 只作为 client-server API 契约。
