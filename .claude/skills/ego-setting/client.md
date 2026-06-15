# ego-setting client

前端设置页面 context。

## 路由

`/setting` — 需登录，从 AppShell 左上角齿轮图标通过 `context.push()` 进入。图标仅在 `/now`、`/past`、`/starmap` 三个根路由显示，push 子页面（如 `/past/detail/:traceId`、`/starmap/detail/:constellationId`）不显示。
`/feedback` — 需登录，从设置页「用户反馈」行 push 进入。

## 核心文件

| 文件 | 说明 |
|------|------|
| `client/lib/features/setting/setting_page.dart` | 设置页 UI: 分区 icon 行 + copy / push 交互，`ConsumerStatefulWidget` |
| `client/lib/features/setting/feedback_page.dart` | 反馈页: 多行文本输入 + 提交按钮 + loading/error 状态，`ConsumerStatefulWidget` |
| `client/lib/core/version.dart` | `make version` 生成，`appVersion` 常量来自 `git describe --tags` |
| `client/lib/shared/widgets/app_shell.dart` | 左上角 `Icons.settings_outlined` 入口，通过 `GoRouterState.of(context).uri.path` 判断仅在根路由（`/now`、`/past`、`/starmap`）显示 |
| `client/lib/features/now/widgets/starry_background.dart` | 星空背景组件 `StarryBackground`，设置页/反馈页复用 |
| `client/lib/core/router/router.dart` | `/setting`、`/feedback` 路由（GoRoute） |
| `client/lib/data/services/ego_client.dart` | `getProfile(WidgetRef ref)` + `submitFeedback(WidgetRef ref, {required String content})` — 携带 token 调用 RPC |
| `client/lib/core/providers/auth_provider.dart` | `AuthNotifier.logout()` — 清除 token |

## 页面结构

```
SettingPage (Scaffold, backgroundColor: AppColors.darkBg)
├── body: Stack
│   ├── StarryBackground() — 底层星空动画
│   └── Column/ListView — 上层内容
│       ├── AppBar（透明背景，金色「设置」标题居中，左侧返回箭头）
│       ├── 账号信息区
│       │   ├── 标签「账号信息」（灰色小字）
│       │   ├── 📱 手机号行：icon + label + 脱敏值 → 点击复制原始手机号
│       │   └── 📅 注册时间行：icon + label + 日期 → 点击复制日期文本
│       ├── 关于区
│       │   ├── 标签「关于」（灰色小字）
│       │   ├── ℹ️ 版本行：icon + label + 版本号 → 点击复制版本号
│       │   ├── 📄 服务条款行：icon + label + 右箭头 → push /terms
│       │   ├── 🛡️ 隐私政策行：icon + label + 右箭头 → push /privacy
│       │   └── 🖊️ 用户反馈行：icon + label + 右箭头 → push /feedback
│       ├── 退出登录按钮（红色边框 + 红色文字，全宽，无确认弹窗）
│       ├── Copyright + 备案号（灰色小字居中，由 Spacer 推至页面底部）
│       │   ├── Copyright © 2026 Ego 工作室 保留所有权利 — 纯文本，不可点击
│       │   └── 闽ICP备2026020313号 — 点击复制备案链接 https://beian.miit.gov.cn/ 到剪贴板
│       └── 区域间通过 SizedBox 分隔，行间通过 1px 细线分割
```

## 状态管理

- **无独立 provider** — 页面内通过 `setState` 管理 loading/profile/error
- **登出**: `ref.read(authProvider.notifier).logout()` → 清除 token → `context.go('/login')`

## 数据流

```
1. initState() → _loadProfile()
2. ref.read(EgoClient.provider).getProfile(ref) → gRPC GetProfile
3. 成功 → setState(_profile, _rawPhone, _loading=false)
4. 失败 → setState(_error, _loading=false)
```

### FeedbackPage 数据流

```
1. 用户输入文本 → 点击「提交反馈」
2. 本地校验非空（空则 SnackBar 提示）
3. setState(_state=submitting) → 按钮显示 loading
4. ref.read(EgoClient.provider).submitFeedback(ref, content:) → gRPC SubmitFeedback
5. 成功 → SnackBar「感谢你的反馈！」→ context.pop()
6. 失败 → setState(_errorMsg, _state=error) → 显示错误文本
```

## 交互

| 行 | 点击行为 |
|----|---------|
| 手机号 | 复制原始手机号到剪贴板，SnackBar 提示「手机号已复制」 |
| 注册时间 | 复制日期文本到剪贴板，SnackBar 提示「注册时间已复制」 |
| 版本 | 复制版本号到剪贴板，SnackBar 提示「版本号已复制」 |
| 服务条款 | `context.push('/terms')` → TermsPage（位于 login feature） |
| 隐私政策 | `context.push('/privacy')` → PrivacyPage（位于 login feature） |
| 用户反馈 | `context.push('/feedback')` → FeedbackPage（位于 setting feature） |

## gRPC 调用

| 方法 | RPC | 认证 |
|------|-----|------|
| `getProfile(ref)` | GetProfile | Bearer token（从 authProvider 读取） |
| `submitFeedback(ref, content:)` | SubmitFeedback | Bearer token（从 authProvider 读取） |
