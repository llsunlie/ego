# ego 后端结构规范

后端采用 Go 模块化单体，按 DDD 限界上下文组织。本文是 Agent 进入 `server/` 后理解结构、边界和责任归属的入口文件。

## 1. 架构规则

```text
Go modular monolith
+ DDD bounded contexts
+ domain / app / adapter
+ platform infrastructure
+ gRPC
+ PostgreSQL / pgvector
```

硬规则：

- `proto` 只作为前后端 API 契约，不作为后端领域模型。
- 业务模块内部统一使用 `domain / app / adapter`。
- 基础设施能力统一放在 `internal/platform`。
- 每张业务表只有一个写入 owner。
- 模块间只通过 interface、port、read model、domain event 协作。
- 非 owner 不修改其他模块内部实现。

## 2. 目录总览

```text
server/
├── AGENTS.md                    # 后端总规则
├── Makefile                     # 后端标准命令
├── go.mod / go.sum
├── sqlc.yaml
├── .harness/                    # 后端集成级 harness
│
├── cmd/
│   ├── ego/                     # 服务启动入口
│   └── migrate/                 # 数据库迁移入口
│
├── proto/                       # Go gRPC 生成物，不手写
│   └── ego/
│       ├── api.pb.go
│       └── api_grpc.pb.go
│
└── internal/
    ├── bootstrap/               # 依赖组装
    ├── config/                  # 配置读取
    ├── shared/                  # 极少量共享基础类型
    ├── platform/                # 基础设施
    ├── identity/                # 登录 / 用户身份
    ├── writing/                 # 此刻 / Trace / Moment / Echo
    ├── timeline/                # 过往 / 记忆光点
    ├── starmap/                 # 星图 / 星座 / 星座资产
    └── conversation/            # 和过去自己对话
```

## 3. 顶层目录职责

| 路径 | 职责 | 禁止 |
| --- | --- | --- |
| `cmd/ego` | 启动后端服务 | 业务逻辑、SQL、RPC 行为实现 |
| `cmd/migrate` | 执行数据库迁移 | 业务逻辑 |
| `server/proto` | 存放 Go 生成代码 | 手写或手改 `*.pb.go` |
| `internal/bootstrap` | 组装依赖、注册 handler、连接模块 | 业务规则、SQL、AI 业务判断 |
| `internal/config` | 读取环境变量和配置 | 初始化 DB/AI client、承载业务规则 |
| `internal/shared` | 通用 ID、错误、事件接口、Clock、事务接口 | 业务实体、数据库 model 大杂烩 |
| `internal/platform` | PostgreSQL、JWT、gRPC plumbing、AI adapter、eventbus | ego 业务流程 |

`server/proto` 只保存生成物；契约源文件在根目录：

```text
proto/ego/api.proto
```

## 4. 模块责任矩阵

| 模块 | 定位 | 负责 RPC | 写入 owner | 可读取 |
| --- | --- | --- | --- | --- |
| `platform` | 基础设施 | 无业务 RPC | 无业务表 | 按基础设施需要 |
| `identity` | 身份上下文 | `Login` | `users` | `users` |
| `writing` | 此刻写作上下文 | `CreateMoment`, `GenerateInsight` | `moments`, `traces` | 当前用户历史 `moments` |
| `timeline` | 过往查询上下文 | `ListMoments`, `GetRandomMoments` | 无 | `moments` |
| `starmap` | 星图上下文 | `StashTrace`, `ListConstellations`, `GetConstellation` | `stars`, `constellations`, `constellation_stars`, `insights`, `past_self_cards`, `topic_prompts` | `traces`, `moments`，通过 Writing 契约 |
| `conversation` | 对话上下文 | `StartChat`, `SendMessage` | `chat_sessions`, `chat_messages` | `past_self_cards` 通过 Starmap 契约；`moments` 通过 Writing 契约 |

## 5. 模块边界

### 5.1 Platform

路径：`server/internal/platform`

子模块：

```text
auth        JWT、鉴权拦截器、密码基础能力
postgres    DB pool、migration、sqlc、事务基础设施
grpc        transport 层通用能力
ai          AI SDK adapter、prompt、输出校验
eventbus    领域事件分发、未来 outbox
```

规则：

- 不负责业务 RPC。
- 不拥有业务表写入权。
- 可实现业务模块声明的 port。
- 不在 platform 内编排 `CreateMoment`、`StashTrace`、`SendMessage` 等用例。

### 5.2 Identity

路径：`server/internal/identity`

职责：

- 登录。
- 自动注册。
- 密码校验编排。
- JWT 签发编排。

禁止：

- 读写写作、星图、对话数据。
- 把身份逻辑以外的业务流程放入本模块。

### 5.3 Writing

路径：`server/internal/writing`

职责：

- 创建或延续 Trace。
- 追加 Moment。
- 生成 embedding。
- 匹配 Echo。
- 生成当前体验 Insight。
- 提供 Trace/Moment 只读契约。

禁止：

- 写入星图和对话相关表。
- 实现 `StashTrace`。
- 让 Echo 成为星图资产。

### 5.4 Timeline

路径：`server/internal/timeline`

职责：

- 查询过往 Moment。
- 游标分页。
- 随机读取记忆光点 Moment。

禁止：

- 创建、更新、删除 Moment。
- 更新 `connected` 状态。
- 写入任何业务表。

### 5.5 Starmap

路径：`server/internal/starmap`

职责：

- 将 Trace 寄存为 Star。
- 聚类 Constellation。
- 维护星座状态。
- 生成星座级 Insight。
- 生成 PastSelfCard。
- 生成 TopicPrompt。
- 组装星图和星座详情 read model。

禁止：

- 创建或更新 Moment 内容。
- 创建 ChatSession / ChatMessage。
- 同步阻塞前端飞星动效等待聚类或 AI 资产完成。

### 5.6 Conversation

路径：`server/internal/conversation`

职责：

- 创建或恢复 ChatSession。
- 保存用户消息。
- 构建 PastSelfContext。
- 调用 PastSelfResponder。
- 校验引用来源。
- 保存 past-self 回复。

禁止：

- 创建 Moment。
- 修改 Constellation。
- 生成 PastSelfCard / TopicPrompt。
- 返回无引用来源且无越界说明的 past-self 回复。

## 6. 业务模块内部结构

所有业务模块采用同一结构：

```text
internal/{module}/
├── AGENTS.md             # 模块 agent 工作规则
├── ARCHITECTURE.md       # 模块内部结构
├── CONTRACT.md           # 模块对外契约
├── .harness/             # 模块级 harness
├── domain/               # 领域模型与领域规则
├── app/                  # 用例编排
└── adapter/
    ├── grpc/             # gRPC handler 与 proto mapper
    └── postgres/         # repository / read model 实现
```

| 层 | 放什么 | 禁止 |
| --- | --- | --- |
| `domain` | Entity、Value Object、Aggregate、Domain Service、Repository interface、Domain Event、业务错误 | proto、pgx、sqlc、grpc/status、platform/ai、config |
| `app` | 用例编排、事务边界、调用 repository interface、调用 port、发布事件 | SQL 字符串、具体 AI SDK、proto request/response |
| `adapter/grpc` | handler、proto mapper、request/response 转换 | 领域规则 |
| `adapter/postgres` | repository 实现、read model、sqlc/pgx 适配、持久化模型转换 | 业务决策 |

## 7. 模型边界

| 模型 | 位置 | 使用范围 |
| --- | --- | --- |
| Proto DTO | `server/proto/ego/*.pb.go` | 只在 `adapter/grpc` 和启动注册代码中使用 |
| Domain Model | `server/internal/{module}/domain` | 表达业务概念和规则 |
| Persistence Model | `server/internal/{module}/adapter/postgres`、`server/internal/platform/postgres/sqlc` | 仅持久化层内部使用 |

规则：

- Proto DTO 不进入 `domain`。
- Domain Model 不带数据库 tag。
- Persistence Model 进入 `app/domain` 前必须转换。

## 8. 依赖方向

允许：

```text
adapter → app → domain
app → domain
adapter/grpc → server/proto
adapter/postgres → platform/postgres
bootstrap → 各模块 app/adapter
platform → 外部 SDK / 基础设施库
```

禁止：

```text
domain → adapter
domain → platform
domain → proto
domain → sqlc / pgx
platform → 业务模块 app/adapter
业务模块 → 其他模块内部实现
```

## 9. 模块协作方式

允许：

- Go interface。
- domain port。
- application interface。
- read model。
- domain event。

禁止：

- 用 proto 类型作为后端模块间模型。
- import 其他模块内部 repository。
- 修改其他模块拥有的数据表。
- 复制其他模块内部实现逻辑。
