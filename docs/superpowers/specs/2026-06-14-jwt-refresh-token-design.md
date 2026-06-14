# JWT Refresh Token 机制设计

**日期**: 2026-06-14
**状态**: 已定稿

## 概述

从单一 JWT token 升级为双 token 体系：
- **Access token**: 短期（1h），用于日常 API 鉴权
- **Refresh token**: 长期（30d），仅用于获取新的 access token
- **简单模式**: refresh token 不轮换，每次 RefreshToken RPC 仅返回新 access token

## 决策记录

| 决策 | 选项 | 原因 |
|------|------|------|
| Refresh token 轮换 | 不轮换 | 简单实现，减少状态管理 |
| LoginRes 字段命名 | `access_token` + `refresh_token` | 语义清晰，明示双 token |
| RefreshToken 输入输出 | 仅需 refresh token，返回 access token | 最简设计 |
| 过期时间配置 | `JWT_ACCESS_EXP_HOURS` + `JWT_REFRESH_EXP_DAYS` | 灵活可调 |
| 前端刷新触发 | ego_client 调用层自动处理 | 对 UI 透明，零侵入 |

## 1. Proto 变更

### 新 RPC

```protobuf
rpc RefreshToken(RefreshTokenReq) returns (RefreshTokenRes);

message RefreshTokenReq {
  string refresh_token = 1;
}

message RefreshTokenRes {
  string access_token = 1;
}
```

### 修改现有 message

```protobuf
message LoginRes {
  string access_token = 1;   // 旧字段 "token" 改名
  string refresh_token = 2;  // 新增
}

message RegisterRes {
  string access_token = 1;
  string refresh_token = 2;
}

message ResetPasswordRes {
  string access_token = 1;
  string refresh_token = 2;
}
```

## 2. 后端 DDD 变更

### 配置 (`server/internal/config/config.go`)

- 移除 `JWTExpHours`
- 新增 `JwtAccessExpHours string` (env: `JWT_ACCESS_EXP_HOURS`, 默认 1)
- 新增 `JwtRefreshExpDays string` (env: `JWT_REFRESH_EXP_DAYS`, 默认 30)
- `.env.example` 同步更新

### platform/auth

**`jwt.go`**:
- `GenerateJWT` 保持不变，通过传入不同 `expiration` 区分 access/refresh
- 新增 `ParseJWTWithType(tokenStr, secret, expectedType)` — 验证时检查 `token_type` claim，防止 access token 冒充 refresh token 或反之
- Refresh token claims: `{ "user_id": "...", "token_type": "refresh", "iat": ..., "exp": ... }`
- Access token claims: `{ "user_id": "...", "token_type": "access", "iat": ..., "exp": ... }`

**`jwt_issuer.go`**:
- `JWTIssuer` 新增字段 `RefreshExp time.Duration`
- 新增方法 `IssueRefresh(userID string) (string, error)` — 签发 refresh token（含 `token_type: "refresh"`）
- `Issue` 签名改为签发 access token（含 `token_type: "access"`）

**`interceptor.go`**:
- `RefreshToken` 加入 `PreAuthMethods`（refresh token 本身是凭证，不走 access token 鉴权）

### identity/domain

**`errors.go`**: 新增 `ErrInvalidRefreshToken`

### identity/app

**`ports.go`**: `TokenIssuer` 接口新增 `IssueRefresh(userID string) (string, error)`

**`refresh_token.go`** (新文件):
```
RefreshTokenUseCase:
1. ParseJWTWithType(refreshToken, secret, "refresh")
2. 失败 → ErrInvalidRefreshToken
3. 检查过期（ParseJWTWithType 已含过期检查）
4. tokens.Issue(userID) → 签发新 access token
```

**`login.go`**、**`register.go`**、**`reset_password.go`**:
- 改为调用 `Issue(userID)` + `IssueRefresh(userID)`
- 返回双 token

### identity/adapter/grpc

**`handler.go`**: 新增 `RefreshToken` handler 方法，调用 `RefreshTokenUseCase`

### identity/module.go

- 新增 `RefreshTokenUseCase` 注入

### bootstrap

**`platform.go`**:
- `Platform` 新增 `JWTRefreshExp time.Duration` 字段
- `InitPlatform`: 解析 `JWT_ACCESS_EXP_HOURS` + `JWT_REFRESH_EXP_DAYS`
- `JWTIssuer` 初始化传入双过期时间

**`identity.go`**: 无需改动（identity deps 不变，TokenIssuer 已含 IssueRefresh）

**`composite.go`**: 路由 `RefreshToken` → identity handler

## 3. 前端变更

### auth_provider.dart

- `AuthState`: `token` → `accessToken`，新增 `refreshToken`
- `AuthNotifier.login`: 参数改为 `(String accessToken, String refreshToken)`
- `AuthNotifier.refreshToken(newAccessToken)`: 更新 access token
- 持久化逻辑通过 `LocalStore` 处理双 token

### local_store.dart

- 新增 `getRefreshToken()` / `setRefreshToken(token)` / `clearRefreshToken()`
- `clearToken()` → 同时清除 access + refresh

### ego_client.dart

- 新增 `refreshToken()` 方法
- 自动刷新包装器：带认证的请求捕获 `GrpcError.unauthenticated` → 读取 refresh token → 调用 RefreshToken → 更新 auth state → 重试原请求（仅重试一次）

### login_page.dart

- 处理 Login/Register/ResetPassword 新的 response 字段

## 4. 错误映射

| 领域错误 | gRPC Status | 前端处理 |
|----------|-------------|----------|
| `ErrInvalidRefreshToken` (过期/伪造/access冒充) | `Unauthenticated` | 清除所有 token → 跳转登录 |
| RefreshToken RPC 本身无 token | `Unauthenticated` | 同上 |

## 5. 数据流

```
┌─ 登录 ───────────────────────────────────────┐
│ Login RPC                                    │
│   → 签发 access(1h) + refresh(30d)            │
│   → 前端存储双 token                          │
└──────────────────────────────────────────────┘

┌─ 正常请求 ────────────────────────────────────┐
│ CreateMoment (Authorization: Bearer <access>) │
│   → interceptor ParseJWT                     │
│   → user_id → context                        │
│   → handler → 200                            │
└──────────────────────────────────────────────┘

┌─ 过期刷新 ────────────────────────────────────┐
│ CreateMoment (过期的 access token)            │
│   → interceptor → UNAUTHENTICATED            │
│   → ego_client 捕获                          │
│   → RefreshToken (refresh_token)             │
│     → ParseJWTWithType("refresh")            │
│     → Issue new access(1h)                   │
│   → 更新 access token                        │
│   → 重试 CreateMoment → 200                   │
└──────────────────────────────────────────────┘

┌─ refresh 也过期 ──────────────────────────────┐
│ RefreshToken (过期的 refresh token)           │
│   → ParseJWTWithType 失败                     │
│   → UNAUTHENTICATED                          │
│   → 前端清除所有 token                        │
│   → 跳转 /login                               │
└──────────────────────────────────────────────┘
```

## 6. 未涉及

- 数据库变更（无需。refresh token 自包含）
- 服务端黑名单/版本号（无需。简单模式不轮换）
- 前端页面 UI 改动（自动刷新对 UI 透明）
