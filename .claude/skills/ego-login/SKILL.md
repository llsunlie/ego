---
name: ego-login
description: 登录注册 feature — 前端 client page: /login + 后端 identity 领域。手机号+密码+短信验证码。涉及文件: client/lib/features/login/ + server/internal/identity/。
---

# ego-login

登录/注册全栈 feature context。修改此功能时，agent 应同时阅读以下文件了解前后端完整上下文：

- 前端详细 context ➔ Read `client.md`
- 后端详细 context ➔ Read `server.md`

## 快速文件索引

### 前端 (`client/`)
| 文件 | 说明 |
|------|------|
| `client/lib/features/login/login_page.dart` | 4-step 登录 UI（手机号→密码/验证码注册/忘记密码） |
| `client/lib/core/providers/auth_provider.dart` | Auth 状态管理 |
| `client/lib/data/services/ego_client.dart` | `checkPhone()` / `sendVerificationCode()` / `register()` / `login()` / `resetPassword()` |
| `client/lib/core/router/router.dart` | `/login` 路由 + 守卫 |

### 后端 (`server/`)
| 文件 | 说明 |
|------|------|
| `server/internal/identity/` | identity 有界上下文: CheckPhone/Login/Register/SendCode/ResetPassword |
| `server/internal/identity/adapter/sms/aliyun.go` | 阿里云短信认证 SDK |
| `server/internal/platform/auth/interceptor.go` | JWT 拦截器（5 个 RPC 免认证） |
| `server/internal/bootstrap/composite.go` | gRPC 方法路由 |

### Proto 契约
| 文件 | 说明 |
|------|------|
| `proto/ego/api.proto` | `CheckPhone` / `SendVerificationCode` / `Register` / `Login` / `ResetPassword` |
