# ego-setting client

前端设置页面 context。

## 路由

`/setting` — 需登录，从 AppShell 左上角齿轮图标通过 `context.push()` 进入。

## 核心文件

| 文件 | 说明 |
|------|------|
| `client/lib/features/setting/setting_page.dart` | 设置页 UI，`ConsumerStatefulWidget` |
| `client/lib/shared/widgets/app_shell.dart` | 左上角 `Icons.settings_outlined` 入口 |
| `client/lib/core/router/router.dart` | `/setting` 路由（GoRoute） |
| `client/lib/data/services/ego_client.dart` | `getProfile(WidgetRef ref)` — 携带 token 调用 RPC |
| `client/lib/core/providers/auth_provider.dart` | `AuthNotifier.logout()` — 清除 token |

## 页面结构

```
SettingPage
├── AppBar（透明背景，金色「设置」标题居中，左侧返回箭头）
├── 账号信息区
│   ├── 标签「账号信息」（灰色小字）
│   ├── 手机号行：标签 + 脱敏值（138****8888）
│   └── 注册时间行：标签 + 格式化日期（2025/01/15）
└── 退出登录按钮（红色边框 + 红色文字，全宽，无确认弹窗）
```

## 状态管理

- **无独立 provider** — 页面内通过 `setState` 管理 loading/profile/error
- **登出**: `ref.read(authProvider.notifier).logout()` → 清除 token → `context.go('/login')`

## 数据流

```
1. initState() → _loadProfile()
2. ref.read(EgoClient.provider).getProfile(ref) → gRPC GetProfile
3. 成功 → setState(_profile, _loading=false)
4. 失败 → setState(_error, _loading=false)
```

## gRPC 调用

| 方法 | RPC | 认证 |
|------|-----|------|
| `getProfile(ref)` | GetProfile | Bearer token（从 authProvider 读取） |
