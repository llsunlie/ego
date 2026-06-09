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
| `client/lib/features/setting/setting_page.dart` | 设置页 UI：icon 分区行 + copy/导航交互 + 服务条款/隐私政策/用户反馈入口 + 红色登出按钮 |
| `client/lib/features/setting/feedback_page.dart` | 反馈页：多行文本输入 + 提交到后端 |
| `client/lib/core/version.dart` | `make version` 生成 `appVersion` 常量，来源 `git describe --tags` |
| `client/lib/shared/widgets/app_shell.dart` | 左上角齿轮图标入口（push 导航到 /setting） |
| `client/lib/core/router/router.dart` | `/setting`、`/feedback` 路由（普通 GoRoute，需登录） |
| `client/lib/data/services/ego_client.dart` | `getProfile()` + `submitFeedback()` — 携带 JWT 调用 RPC |
| `client/lib/core/providers/auth_provider.dart` | `AuthNotifier.logout()` — 清除 token + 重置状态 |

### 后端 (`server/`)
| 文件 | 说明 |
|------|------|
| `server/internal/setting/` | setting 有界上下文: GetProfile, SubmitFeedback |
| `server/internal/setting/domain/ports.go` | `UserInfo` 类型 + `UserReader` 接口 + `FeedbackWriter` 接口 |
| `server/internal/setting/domain/feedback.go` | `Feedback` 实体 |
| `server/internal/setting/domain/errors.go` | `ErrUserNotFound`、`ErrFeedbackEmpty` |
| `server/internal/setting/app/profile.go` | `GetProfileUseCase` — 从 context 提取 user_id 后查库 |
| `server/internal/setting/app/feedback.go` | `SubmitFeedbackUseCase` + `IDGenerator` 接口 |
| `server/internal/setting/adapter/id/uuid.go` | UUID ID 生成器 |
| `server/internal/setting/adapter/grpc/handler.go` | gRPC handler + mapError |
| `server/internal/setting/adapter/postgres/user_reader.go` | `UserReader` 实现（复用 sqlc GetUserByID） |
| `server/internal/setting/adapter/postgres/feedback_writer.go` | `FeedbackWriter` 实现（sqlc InsertFeedback） |
| `server/internal/setting/module.go` | DI 装配 |
| `server/internal/bootstrap/setting.go` | `NewSettingHandler(p *Platform)` |
| `server/internal/bootstrap/composite.go` | `GetProfile`、`SubmitFeedback` 委派给 setting handler |
| `server/internal/platform/postgres/queries/users.sql` | `GetUserByID` 查询 |
| `server/internal/platform/postgres/queries/feedbacks.sql` | `InsertFeedback` 查询 |
| `server/internal/platform/postgres/migrations/011_feedback.sql` | feedbacks 表 migration |
| `server/internal/platform/auth/interceptor.go` | GetProfile/SubmitFeedback 不在 preAuthMethods 中 → 需 JWT |

### Proto 契约
| 文件 | 说明 |
|------|------|
| `proto/ego/api.proto` | `GetProfile` + `SubmitFeedback(SubmitFeedbackReq) returns (SubmitFeedbackRes)` — phone + created_at / id + created_at |
