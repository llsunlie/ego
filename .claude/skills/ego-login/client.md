# ego-login client

前端登录/注册页面 context。

## 路由

`/login` — 未登录用户自动重定向到此路由。

## 核心文件

| 文件 | 说明 |
|------|------|
| `client/lib/features/login/login_page.dart` | 登录页面 UI，3-step 流程 |
| `client/lib/core/providers/auth_provider.dart` | AuthState (token, isLoggedIn)，login/logout 方法 |
| `client/lib/data/services/ego_client.dart` | gRPC API 客户端：checkPhone, sendVerificationCode, register, login |
| `client/lib/data/generated/api.pbgrpc.dart` | 生成的 EgoClient gRPC stub |
| `client/lib/core/router/router.dart` | GoRouter 路由配置 |

## 交互流程 (4 Step)

```
Step 0: 输入手机号 → CheckPhone RPC
  ├─ registered=true → Step 1 (密码登录)
  │   └─ 点击「忘记密码？」→ Step 3 (重置密码)
  └─ registered=false → 自动发送验证码 → Step 2 (验证码注册)

Step 1: 输入密码 → Login RPC → 跳转 /onboard 或 /now

Step 2: 输入验证码 + 设置密码 → Register RPC → 跳转 /onboard

Step 3: 输入验证码 + 新密码 → ResetPassword RPC → 自动登录 → 跳转首页
```

## 状态管理

- **authProvider** (`StateNotifierProvider<AuthNotifier, AuthState>`)
  - `AuthState`: `token` (String?), `isLoggedIn` (bool)
  - `AuthNotifier.login(token)` — 保存 token 到 LocalStore
  - `AuthNotifier.logout()` — 清除 token

## 页面逻辑

- `_phoneCtrl`, `_passwordCtrl`, `_codeCtrl` — 输入控制器
- `_step` — 0=手机号, 1=密码登录, 2=验证码注册, 3=重置密码
- `_codeSentPhone` — 缓存已发验证码的手机号，避免重复发送触发频率限制（Step 2/3 共享）
- `_countdown` — 验证码倒计时秒数（倒计时保持，返回 Step 0 不重置）

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
if (!loggedIn && !isLoginRoute) return '/login';
if (loggedIn && isLoginRoute) return onboardingDone ? '/now' : '/onboard';
```
