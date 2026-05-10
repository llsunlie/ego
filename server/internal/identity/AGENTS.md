# AGENTS.md - Identity 模块开发指南

## 0. 全局上下文溯源

如果你是在 `server/internal/identity/` 目录下被唤醒的 Agent，请在执行任何代码前先向上读取：

1. `../../../AGENTS.md`
2. `../../AGENTS.md`

Identity 是"我是谁"和"我能进来吗"的上下文。

## 1. 模块定位

Identity 负责用户身份认证，回答：

> 用户如何以最小摩擦力进入 ego 空间？

本模块负责 User 注册、密码验证和 JWT 签发。

## 2. 负责的接口范围

| RPC | 责任 |
| --- | --- |
| `Login` | Auto-register (account + password) 或登录验证，返回 JWT token 和 created 标志 |

## 3. 模块边界

### 3.1 拥有的业务能力

- 密码哈希与验证。
- 用户持久化（首次登录即注册）。
- JWT token 签发。

### 3.2 数据归属

```text
users
```

### 3.3 禁止事项

- 禁止自行生成 Moment 或 Trace。
- 禁止在 `domain/` 中直接调用 pgx、sqlc 或 proto。

## 4. 架构与装配

- **两级装配**：`identity/module.go` 负责组装 adapter、app use case、gRPC handler；`bootstrap/identity.go` 注入 DB + Hasher + Tokens 等进程级资源或外部能力。
- **外部能力注入**：`PasswordHasher` 和 `TokenIssuer` 是基础设施能力，由 bootstrap 注入，identity 只依赖自身 app 层声明的接口。
- **ID 生成归属模块**：UUID 生成器位于 `adapter/id/uuid.go`，满足 `app.IDGenerator` 接口。

## 5. 常用开发命令

从 `server/` 目录运行：

```text
go test ./internal/identity/...
go test ./...
go build ./...
```
