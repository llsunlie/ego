# AGENTS.md - Platform 模块开发指南

## 0. 全局上下文溯源

如果你是在 `server/internal/platform/` 目录下被唤醒的 Agent，请在执行任何代码前先向上读取：

1. `../../../AGENTS.md`
2. `../../AGENTS.md`

Platform 是基础设施模块，不是业务上下文。全局 DDD 依赖方向和模块 ownership 规则必须优先遵守。

## 1. 模块定位

Platform 负责后端基础设施能力，给业务模块提供可注入的技术实现。

它可以实现业务模块声明的 Port，但不拥有 ego 的业务流程。

## 2. 负责的接口范围

Platform 不负责实现任何前后端业务 RPC。

它负责提供后端内部基础设施能力：

| 能力范围 | 说明 |
| --- | --- |
| `postgres` | PostgreSQL 连接池、迁移、sqlc 生成查询、事务基础设施 |
| `auth` | JWT 生成/解析、gRPC 鉴权拦截器、密码基础能力 |
| `grpc` | gRPC server plumbing、错误映射、transport 辅助能力 |
| `ai` | AI SDK adapter、prompt 模板、输出校验 |
| `eventbus` | 进程内事件总线、未来 outbox 基础设施 |

业务 RPC 必须由各自业务模块的 `adapter/grpc` 实现。

## 3. 模块边界

### 3.1 拥有的基础设施能力

- 数据库连接与 sqlc 生成物。
- JWT / bcrypt 等认证基础设施。
- AI 客户端与模型供应商适配。
- gRPC transport 层通用工具。
- 领域事件分发基础设施。

### 3.2 数据归属

Platform 可以承载数据库迁移和 sqlc 生成物，但不因此拥有业务表的写入权。

业务表写入归属仍然属于各业务模块：

| 表 | 写入 owner |
| --- | --- |
| `users` | Identity |
| `moments` / `traces` | Writing |
| `stars` / `constellations` / `insights` / `past_self_cards` / `topic_prompts` | Starmap |
| `chat_sessions` / `chat_messages` | Conversation |

### 3.3 禁止事项

- 禁止在 Platform 中创建 Moment、Trace、Star、Constellation、ChatSession。
- 禁止在 Platform 中实现 `Login`、`CreateMoment`、`StashTrace` 等业务用例。
- 禁止让业务模块直接依赖具体 AI SDK；应通过业务模块定义的 interface 注入。
- 禁止把 `platform/postgres/sqlc` 类型泄漏进业务 `domain/`。

## 4. 依赖规则

- Platform 可以依赖外部 SDK 和技术库。
- Platform 不应该依赖具体业务模块的 app 或 adapter。
- Platform 可以实现业务模块在 `domain/ports.go` 或 `app` 层声明的接口。
- 业务模块需要新基础设施能力时，应先在自己的边界内定义需要什么，再由 Platform 提供实现。

## 5. 常用开发命令

从 `server/` 目录运行：

```text
go test ./internal/platform/...
go test ./...
go build ./...
```
