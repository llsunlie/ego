# ego-login client

前端登录/注册页面 context。

## 路由

`/login` — 未登录用户自动重定向到此路由。

## 核心文件

| 文件 | 说明 |
|------|------|
| `client/lib/features/login/login_page.dart` | 登录页面 UI，4-step 流程 |
| `client/lib/features/login/terms_page.dart` | 服务条款页面 |
| `client/lib/features/login/privacy_page.dart` | 隐私政策页面 |
| `client/lib/core/providers/auth_provider.dart` | AuthState (token, isLoggedIn)，login/logout 方法（async） |
| `client/lib/data/repositories/local_store.dart` | Token 持久化（Android: flutter_secure_storage, Web: Hive），settings |
| `client/lib/data/services/ego_client.dart` | gRPC API 客户端：checkPhone, sendVerificationCode, register, login |
| `client/lib/data/generated/api.pbgrpc.dart` | 生成的 EgoClient gRPC stub |
| `client/lib/core/router/router.dart` | GoRouter 路由配置（含 /terms /privacy 免登录路由） |
| `client/lib/features/now/widgets/starry_background.dart` | 星空动画背景组件，被 login_page 引用 |

## 交互流程 (4 Step)

```
Step 0: 输入手机号 → CheckPhone RPC
  ├─ registered=true → Step 1 (密码登录)
  │   └─ 点击「忘记密码？」→ Step 3 (重置密码)
  └─ registered=false → 自动发送验证码 → Step 2 (验证码注册)

Step 1: 输入密码 → Login RPC → 跳转 /onboard 或 /now

Step 2: 输入验证码 + 设置密码 + 勾选同意协议 checkbox → Register RPC → 跳转 /onboard
  ├─ Checkbox: "我已阅读并同意《服务条款》和《隐私政策》"
  ├─ 《服务条款》可点击跳转 /terms
  └─ 《隐私政策》可点击跳转 /privacy

Step 3: 输入验证码 + 新密码 → ResetPassword RPC → 自动登录 → 跳转首页
```

## 状态管理

- **authProvider** (`StateNotifierProvider<AuthNotifier, AuthState>`)
  - `AuthState`: `accessToken` (String?), `refreshToken` (String?), `isLoggedIn` (bool)
  - `AuthNotifier.login(accessToken, refreshToken)` — async，写入双 token
  - `AuthNotifier.refreshAccessToken(newAccessToken)` — 更新 access token（仅内存 + 存储）
  - `AuthNotifier.logout()` — async，清除双 token
- **Token 存储** (`LocalStore` in `client/lib/data/repositories/local_store.dart`)
  - `kIsWeb` 编译时分流：Native → FlutterSecureStorage，Web → Hive
  - storage key: `access_token` / `refresh_token`
  - `getToken()` / `setToken()` — access token 存取
  - `getRefreshToken()` / `setRefreshToken()` — refresh token 存取
  - `clearToken()` — 同时清除双 token
- **自动刷新** (`EgoClient._autoRefresh` in `client/lib/data/services/ego_client.dart`)
  - 捕获 `GrpcError.unauthenticated` → 调用 RefreshToken RPC → 更新 access token → 重试原请求（仅一次）
  - 刷新失败 → 清除双 token → 路由守卫自动跳转登录页

## 页面逻辑

- `_phoneCtrl`, `_passwordCtrl`, `_codeCtrl` — 输入控制器
- `_step` — 0=手机号, 1=密码登录, 2=验证码注册, 3=重置密码
- `_agreedToTerms` — Step 2 协议勾选状态，每次进入 Step 2 重置为 false
- `_codeSentPhone` — 缓存已发验证码的手机号，避免重复发送触发频率限制（Step 2/3 共享）
- `_countdown` — 验证码倒计时秒数（倒计时保持，返回 Step 0 不重置）

## UI / Layout

- `login_page.dart` 使用 `StarryBackground` 组件（来自 `client/lib/features/now/widgets/starry_background.dart`）作为背景。
- `Scaffold` 设置 `backgroundColor: AppColors.darkBg`。
- `body` 使用 `Stack` 包裹，`StarryBackground()` 作为底层，表单内容在上层。
- `privacy_page.dart` 和 `terms_page.dart` **不使用** StarryBackground，保持简洁的普通布局。

## gRPC 调用

| 方法 | RPC | 时机 |
|------|-----|------|
| `checkPhone(phone)` | CheckPhone | Step 0 点击"下一步" |
| `sendVerificationCode(phone)` | SendVerificationCode | 新手机自动调用 / "重新发送" / 进入 Step 3 自动调用 |
| `register(phone, code, password)` | Register | Step 2 点击"注册" |
| `login(phone, password)` | Login | Step 1 点击"登录" |
| `resetPassword(phone, code, newPassword)` | ResetPassword | Step 3 点击"重置密码" |

## 错误处理

使用 `GrpcError` 类型匹配，按 status code 区分：
- `Unauthenticated` → "密码错误" / "验证码错误"
- `NotFound` → "用户不存在"
- `AlreadyExists` → "该手机号已注册"

## 路由守卫

```dart
// router.dart redirect:
if (!loggedIn && !isLoginRoute && !isTermsRoute && !isPrivacyRoute) return '/login';
if (loggedIn && isLoginRoute) return onboardingDone ? '/now' : '/onboard';
```
`/terms` 和 `/privacy` 为免登录路由，从登录页的协议链接访问。
