---
name: ego-onboarding
description: 引导流程 feature — 前端 client page: /onboard，新用户 5 步引导。纯前端，无对应后端领域。涉及文件: client/lib/features/onboarding/。
---

# ego-onboarding

引导流程全栈 context（纯前端，无独立后端领域）。

- 前端详细 context ➔ Read `client.md`

## 快速文件索引

### 前端 (`client/`)
| 文件 | 说明 |
|------|------|
| `client/lib/features/onboarding/onboarding_page.dart` | 5 步引导 UI |
| `client/lib/features/onboarding/onboarding_data.dart` | 引导内容数据 |
| `client/lib/core/providers/onboarding_provider.dart` | 引导完成状态 |
| `client/lib/core/router/router.dart` | `/onboard` 路由 + 守卫 |
| `client/lib/core/theme/colors.dart` | 主题色 |

### 后端 (`server/`)
无独立后端领域。引导完成状态通过 `onboarding_provider` 持久化在客户端 LocalStore。
