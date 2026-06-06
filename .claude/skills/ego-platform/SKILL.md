---
name: ego-platform
description: 服务端基础设施 — Postgres/AI/JWT/日志/配置/Bootstrap 组装、sqlc 数据层。支撑所有业务域，涉及文件: server/internal/{platform,config,bootstrap}/。
---

# ego-platform

服务端基础设施 context（纯后端，无独立前端 page）。

- 后端详细 context ➔ Read `server.md`

## 快速文件索引

| 文件 | 说明 |
|------|------|
| `server/internal/config/config.go` | .env 配置加载 |
| `server/internal/platform/postgres/` | pgxpool 连接 + sqlc 生成代码 |
| `server/internal/platform/ai/client.go` | AI API 客户端（Embedding + Chat） |
| `server/internal/platform/auth/` | JWT 签发/验证 + bcrypt + gRPC 拦截器 |
| `server/internal/platform/logging/` | slog 结构化日志 |
| `server/internal/bootstrap/` | 依赖注入组装（platform.go, server.go, composite.go） |
| `server/cmd/ego/main.go` | 入口点 |
