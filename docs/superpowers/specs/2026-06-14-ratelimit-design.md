# Rate Limit Design Spec

## Overview

Server 当前无 API 速率限制。新增 `ratelimit` 包，实现 per-user_id + per-IP 双维度令牌桶限流，通过 config 可配置。

## 限流策略

| 维度 | 免鉴权 RPC | 鉴权 RPC |
|------|-----------|---------|
| IP | ✓ 独立令牌桶 | ✓ 独立令牌桶 |
| user_id | —（不可用） | ✓ 独立令牌桶 |

- **两个维度独立限流，任一维度超限即拒绝请求**
- 免鉴权 RPC 列表（与 `auth/interceptor.go` 的 `preAuthMethods` 保持一致）：Login, CheckPhone, SendVerificationCode, Register, ResetPassword
- 拒绝响应：gRPC status `RESOURCE_EXHAUSTED`，message 指出超限维度（如 `"rate limit exceeded: ip"`）

## Config 参数

新增 5 个环境变量（`.env` / OS env），默认值如下：

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `RATELIMIT_AUTH_RATE` | 10 | 鉴权接口令牌补充速率（tokens/sec） |
| `RATELIMIT_AUTH_BURST` | 20 | 鉴权接口桶容量（允许突发） |
| `RATELIMIT_PREAUTH_RATE` | 10 | 免鉴权接口令牌补充速率（tokens/sec） |
| `RATELIMIT_PREAUTH_BURST` | 30 | 免鉴权接口桶容量（允许突发） |
| `RATELIMIT_MAX_BUCKETS` | 500 | 桶对象数量上限，防内存泄漏 |

### 默认值设计说明

- 鉴权接口 10/20：稳态 10 req/s，允许 20 突发
- 免鉴权接口 10/30：考虑中国用户 NAT 共享 IP，稳态 10 req/s，30 突发
- 桶上限 500：限制 sync.Map 中桶对象总数，防止 IP 欺骗攻击导致内存暴涨

## 架构

### 拦截器链

```
Request → auth.UnaryServerInterceptor → ratelimit.UnaryServerInterceptor → Handler
```

- ratelimit 在 auth **之后**，因为鉴权 RPC 需要 auth 解析的 `user_id`
- auth 对 preAuth RPC 直通（不校验 token），ratelimit 对 preAuth RPC 仅按 IP 限流

### 包结构

```
server/internal/platform/ratelimit/
├── ratelimit.go         # Limiter 结构体 + Allow 方法 + 桶清理
├── ratelimit_test.go    # 单元测试
└── interceptor.go       # gRPC UnaryServerInterceptor
```

### Limiter 设计

```go
type Limiter struct {
    authRate       float64
    authBurst      int
    preAuthRate    float64
    preAuthBurst   int
    maxBuckets     int
    buckets        sync.Map  // string → *rate.Limiter
    preAuthMethods map[string]bool
}
```

- `NewLimiter(cfg)` — 从 config 创建 Limiter，启动后台清理 goroutine
- `Allow(ctx, fullMethod string) (bool, string)` — 判断请求是否允许
  - 从 context 提取 `user_id`（auth 注入）
  - 从 metadata/peer 提取客户端 IP
  - preAuth RPC → 检查 `"preauth:ip:<ip>"` 桶
  - 鉴权 RPC → 检查 `"auth:ip:<ip>"` 桶 + `"auth:user:<userID>"` 桶
  - 任一桶令牌不足 → 返回 `false` + 原因
- `Close()` — 停止清理 goroutine

### 桶管理

- Key 格式：`"preauth:ip:<ip>"`, `"auth:ip:<ip>"`, `"auth:user:<userID>"`
- 创建桶时检查总数：超过 `maxBuckets` 则先清理过期桶
- 清理后仍超上限：fail-open（放行），不创建新桶
- 后台 goroutine 每 1 分钟扫描 sync.Map，删除 5 分钟未访问的桶（`rate.Limiter` 对象）

### IP 提取

1. 从 gRPC metadata 取 `x-forwarded-for`，取第一个 IP（格式 `client, proxy1, ...`）
2. 回退到 `peer.Addr`（去掉端口号）

### 拦截器

```go
func UnaryServerInterceptor(limiter *Limiter) grpc.UnaryServerInterceptor
```

- 调用 `limiter.Allow()` 判断
- 允许 → `handler(ctx, req)`
- 拒绝 → 返回 `status.Error(codes.ResourceExhausted, "rate limit exceeded: <dimension>")`

## 文件变更

### Create

| 文件 | 说明 |
|------|------|
| `server/internal/platform/ratelimit/ratelimit.go` | Limiter 结构体 + Allow + 桶管理 + 清理 |
| `server/internal/platform/ratelimit/ratelimit_test.go` | 单元测试（token allow/deny、桶过期清理、maxBuckets 保护） |
| `server/internal/platform/ratelimit/interceptor.go` | gRPC UnaryServerInterceptor |

### Modify

| 文件 | 变更 |
|------|------|
| `server/internal/config/config.go` | `Config` 结构体新增 5 个字段 + `Load()` 中读取 |
| `server/internal/bootstrap/server.go` | `grpc.UnaryInterceptor(...)` → `grpc.ChainUnaryInterceptor(auth, ratelimit)` |
| `server/internal/platform/auth/interceptor.go` | 将 `preAuthMethods` map 和 `isPreAuthMethod` 函数导出（或 ratelimit 包维护自己的副本） |

### Modify（客户端）

| 文件 | 变更 |
|------|------|
| `client/lib/core/providers/grpc_error_mapper.dart` | **新增** — 统一的 gRPC 错误消息映射工具函数 |
| `client/lib/features/login/login_page.dart` | 在 `on GrpcError catch` 中增加 `resourceExhausted` 分支 |
| 各 feature page（涉及 RPC 调用的页面） | 按需更新 `GrpcError` 捕获，使用统一错误消息映射 |

### 无需变更

- proto（限流对 API 契约透明）
- 数据库

## 错误映射

### 服务端 → 客户端

| 场景 | gRPC Status | Server Message |
|------|------------|---------|
| IP 令牌桶耗尽 | RESOURCE_EXHAUSTED (8) | `"rate limit exceeded: ip"` |
| user_id 令牌桶耗尽 | RESOURCE_EXHAUSTED (8) | `"rate limit exceeded: user"` |
| 桶对象超上限 | 放行（fail-open） | — |

### 客户端 gRPC 状态码 → 用户提示

新增 `client/lib/core/providers/grpc_error_mapper.dart`：

```dart
String grpcErrorMessage(GrpcError e) {
  switch (e.code) {
    case StatusCode.resourceExhausted:
      return '请求过于频繁，请稍后再试';
    case StatusCode.unauthenticated:
      return '登录已过期，请重新登录';
    // ... 其他已有映射
    default:
      return '网络错误，请稍后重试';
  }
}
```

客户端收到 `RESOURCE_EXHAUSTED` 后**不做自动重试**（移动端自动重试会加剧限流），而是展示 Toast 提示用户手动操作。

## 测试策略

### 单元测试 (`ratelimit_test.go`)

1. **正常放行**：新 IP/user 首次请求成功
2. **令牌耗尽**：快速发送 burst+1 个请求，最后一个被拒绝
3. **双维度独立**：IP 耗尽但 user 未耗尽 → 仍被拒绝（IP 维度）
4. **免鉴权 RPC**：仅按 IP 限流，无视 user_id
5. **桶清理**：模拟过期桶被后台 goroutine 删除
6. **maxBuckets 保护**：创建超过上限的桶，验证 fail-open 和清理逻辑

### 集成验证

- `smoke.sh` 增加限流验证：连续发送超出 burst 的请求，验证收到 RESOURCE_EXHAUSTED
- `go vet ./internal/platform/ratelimit/...` 通过
