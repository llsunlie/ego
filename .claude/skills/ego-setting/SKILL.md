---
name: ego-setting
description: Settings/account info/logout feature context. Read-only user profile via GetProfile RPC, client-side logout.
---

# ego-setting

设置/账号信息全栈 feature context。修改此功能时，agent 应同时阅读以下文件了解前后端完整上下文：

- 前端详细 context ➔ Read `client.md`
- 后端详细 context ➔ Read `server.md`

## 快速文件索引

### 前端 (`client/`)
| 文件 | 说明 |
|------|------|
| `client/lib/features/setting/setting_page.dart` | 设置页 UI：脱敏手机号 + 注册时间 + 关于/版本 + 红色登出按钮 |
| `client/lib/core/version.dart` | `make version` 生成 `appVersion` 常量，来源 `git describe --tags` |
| `client/lib/shared/widgets/app_shell.dart` | 左上角齿轮图标入口（push 导航到 /setting） |
| `client/lib/core/router/router.dart` | `/setting` 路由（普通 GoRoute，需登录） |
| `client/lib/data/services/ego_client.dart` | `getProfile()` — 携带 JWT 调用 GetProfile RPC |
| `client/lib/core/providers/auth_provider.dart` | `AuthNotifier.logout()` — 清除 token + 重置状态 |

### 后端 (`server/`)
| 文件 | 说明 |
|------|------|
| `server/internal/setting/` | setting 有界上下文: GetProfile |
| `server/internal/setting/domain/ports.go` | `UserInfo` 类型 + `UserReader` 接口 |
| `server/internal/setting/domain/errors.go` | `ErrUserNotFound` |
| `server/internal/setting/app/profile.go` | `GetProfileUseCase` — 从 context 提取 user_id 后查库 |
| `server/internal/setting/adapter/grpc/handler.go` | gRPC handler + mapError |
| `server/internal/setting/adapter/postgres/user_reader.go` | `UserReader` 实现（复用 sqlc GetUserByID） |
| `server/internal/setting/module.go` | DI 装配 |
| `server/internal/bootstrap/setting.go` | `NewSettingHandler(p *Platform)` |
| `server/internal/bootstrap/composite.go` | `GetProfile` 委派给 setting handler |
| `server/internal/platform/postgres/queries/users.sql` | `GetUserByID` 查询 |
| `server/internal/platform/auth/interceptor.go` | GetProfile 不在 preAuthMethods 中 → 需 JWT |

### Proto 契约
| 文件 | 说明 |
|------|------|
| `proto/ego/api.proto` | `GetProfile(GetProfileReq) returns (GetProfileRes)` — phone + created_at |
