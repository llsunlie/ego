# AGENTS.md - Identity 模块开发指南

## 0. 全局上下文溯源

如果你是在 `server/internal/identity/` 目录下被唤醒的 Agent，请在执行任何代码前先向上读取：

1. `../../../AGENTS.md`
2. `../../AGENTS.md`

全局规则优先于本文件。本文件只定义 Identity 模块的局部边界。

## 1. 模块定位

Identity 是“身份上下文”，只回答一个问题：

> 当前请求是谁发出的？

本模块负责用户登录、自动注册、密码校验编排，以及 JWT 签发编排。当前已有实现位于 `adapter/grpc/handler.go`，后续新业务应逐步拆到 `app/` 和 `domain/`。

## 2. 负责的接口范围

本模块负责实现以下前后端 RPC：

| RPC | 责任 |
| --- | --- |
| `Login` | 用户登录；账号不存在时自动注册；返回 JWT token 和 `created` 标记 |

本模块不负责其他 RPC。

## 3. 模块边界

### 3.1 拥有的业务能力

- account/password 登录。
- account 不存在时自动注册。
- bcrypt 密码校验编排。
- JWT token 签发编排。
- 用户身份数据的创建与读取。

### 3.2 数据归属

Identity 拥有唯一写入权：

```text
users
```

其他模块只能通过认证上下文获得 `user_id`，不能直接读取密码 hash 或修改用户身份数据。

### 3.3 禁止事项

- 禁止写入 `moments`、`stars`、`constellations`、`chat_sessions`、`chat_messages`。
- 禁止在 Identity 中实现写作、星图、时间线或对话业务。
- 禁止把 proto 类型传入 `domain/`。
- 禁止在 `domain/` 中依赖 pgx、sqlc、gRPC status、JWT SDK 或 bcrypt。

## 4. 依赖规则

- 可以使用 `platform/auth` 提供的 JWT 和密码相关基础设施能力。
- 可以通过 `adapter/postgres` 或过渡期的 `platform/postgres/sqlc` 访问 `users`。
- 新增业务规则时，优先放入 `domain/` 或 `app/`，不要继续堆进 gRPC handler。

## 5. 常用开发命令

从 `server/` 目录运行：

```text
go test ./internal/identity/...
go test ./...
go build ./...
```

