# Go Server 内置 TLS (Let's Encrypt)

## 背景

当前全链路无 TLS 加密。需要让 Go Server 内置 TLS 支持，使用 Let's Encrypt 自动签发证书。

## 需求

- Go Server 内置 TLS，不依赖外部反向代理
- 使用 Let's Encrypt 自动签发和续期证书
- 通过 `TLS_DOMAIN` 环境变量配置域名，空值时退回到明文模式
- 三端口架构：明文 web、TLS web、gRPC

## 环境策略

| 环境 | TLS_DOMAIN | 行为 |
|------|-----------|------|
| Dev (localhost) | "" (空) | 三端口全明文 |
| Test | `test.myego.online` | WEB_TLS_PORT + GRPC_PORT 启用 TLS |
| Prod | `myego.online` | WEB_TLS_PORT + GRPC_PORT 启用 TLS |

## 端口规划

| 端口 | 配置项 | 协议 | 服务内容 |
|------|--------|------|----------|
| 9080 | `WEB_PORT` | HTTP 明文 | gRPC-web + 静态文件 |
| 9443 | `WEB_TLS_PORT` | HTTPS (TLS_DOMAIN 时) | gRPC-web + 静态文件 |
| 9444 | `GRPC_PORT` | gRPC (TLS_DOMAIN 时) | gRPC 原生 |

nginx 映射：80→9080, 443→9443

## 技术选型

使用 `golang.org/x/crypto/acme/autocert`：
- 证书自动签发和续期
- TLS-ALPN-01 challenge 自动完成域名验证
- 证书缓存到本地 `certs/` 目录

## 设计

### Config 变更

- 删除 `Port`
- 新增 `WebTLSPort` (env `WEB_TLS_PORT`, default 9443)
- 新增 `GRPCPort` (env `GRPC_PORT`, default 9444)
- `WebPort` 保持不变 (env `WEB_PORT`, default 9080)
- `TLSDomain` (env `TLS_DOMAIN`, 空=TLS 禁用)

### Server 变更

- `NewServer`: TLS_DOMAIN 非空时创建 `autocert.Manager` → `*tls.Config`
- `Serve()`:
  - gRPC 监听 `GRPCPort` (goroutine)，TLS 由 tlsConfig 控制
  - Web 明文监听 `WEB_PORT` (goroutine)
  - Web TLS 监听 `WEB_TLS_PORT` (主线程)，TLS 由 tlsConfig 控制

## 变更文件

1. `server/internal/config/config.go`
2. `server/internal/bootstrap/server.go`
3. `server/.env.example`
4. `server/cmd/ego/main.go`
5. `smoke.sh`
