# Token 安全存储 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 JWT Token 从 Hive 明文迁移到 flutter_secure_storage，Android 走 Keystore，Web 走 WebCrypto 加密。

**Architecture:** 内存缓存 + main() 预加载。`LocalStore.init()` 从 secure storage 预读 token 到 `_cachedToken`，`getToken()` 同步返回缓存，`setToken()`/`clearToken()` 异步写 secure storage。非敏感数据（onboarding、starmap guide）留在 Hive `_settings` box 不变。

**Tech Stack:** Flutter, flutter_secure_storage ^10.0.0, Hive (保留用于非敏感数据)

**Commit 策略（ego 约束）:** 全部 task 完成且验证通过后一次提交，不在中间 task 提交。

---

### Task 1: 添加 flutter_secure_storage 依赖

**Files:**
- Modify: `client/pubspec.yaml`

- [ ] **Step 1: 在 pubspec.yaml 的 dependencies 中添加 flutter_secure_storage**

定位到 `hive_flutter: ^1.1.0` 行，在其后添加：

```yaml
  flutter_secure_storage: ^10.0.0
```

- [ ] **Step 2: 安装依赖**

```bash
cd client && flutter pub get
```

Expected: `exit code 0`，无错误输出。

---

### Task 2: 改造 LocalStore — 核心存储层

**Files:**
- Modify: `client/lib/data/repositories/local_store.dart`

- [ ] **Step 1: 替换文件头部 import**

将现有的：
```dart
import 'package:hive_flutter/hive_flutter.dart';
```

替换为：
```dart
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:hive_flutter/hive_flutter.dart';
```

- [ ] **Step 2: 替换类成员变量和 init()**

将现有的：
```dart
class LocalStore {
  static const _authBox = 'auth';
  static const _settingsBox = 'settings';
  static late Box _auth;
  static late Box _settings;

  static Future<void> init() async {
    await Hive.initFlutter();
    _auth = await Hive.openBox(_authBox);
    _settings = await Hive.openBox(_settingsBox);
  }
```

替换为：
```dart
class LocalStore {
  static const _settingsBox = 'settings';
  static final _secure = FlutterSecureStorage();
  static String? _cachedToken;
  static late Box _settings;

  static Future<void> init() async {
    await Hive.initFlutter();
    _settings = await Hive.openBox(_settingsBox);
    _cachedToken = await _secure.read(key: 'token');
  }
```

- [ ] **Step 3: 替换 token 存取方法**

将现有的：
```dart
  // Auth
  static String? getToken() => _auth.get('token');
  static void setToken(String token) => _auth.put('token', token);
  static void clearToken() => _auth.delete('token');
```

替换为：
```dart
  // Auth — token in secure storage, cached in memory
  static String? getToken() => _cachedToken;

  static Future<void> setToken(String token) async {
    await _secure.write(key: 'token', value: token);
    _cachedToken = token;
  }

  static Future<void> clearToken() async {
    await _secure.delete(key: 'token');
    _cachedToken = null;
  }
```

- [ ] **Step 4: settings 方法保持不变**

确认 `getOnboardingDone`、`setOnboardingDone`、`getStarmapTapGuideShown`、`setStarmapTapGuideShown` 四个方法代码不变（它们通过 `_settings` box 操作，不受影响）。

---

### Task 3: 适配 AuthNotifier — 异步 login/logout

**Files:**
- Modify: `client/lib/core/providers/auth_provider.dart`

- [ ] **Step 1: _loadToken() 保持不变**

确认 `_loadToken()` 方法代码不变（它调用 `LocalStore.getToken()`，现在读内存缓存，同步返回）。

- [ ] **Step 2: login() 改为 async**

将现有的：
```dart
  void login(String token) {
    LocalStore.setToken(token);
    state = AuthState(token: token, isLoggedIn: true);
  }
```

替换为：
```dart
  Future<void> login(String token) async {
    await LocalStore.setToken(token);
    state = AuthState(token: token, isLoggedIn: true);
  }
```

- [ ] **Step 3: logout() 改为 async**

将现有的：
```dart
  void logout() {
    LocalStore.clearToken();
    state = const AuthState();
  }
```

替换为：
```dart
  Future<void> logout() async {
    await LocalStore.clearToken();
    state = const AuthState();
  }
```

---

### Task 4: 静态验证

- [ ] **Step 1: Flutter 静态分析**

```bash
cd client && flutter analyze
```

Expected: `No issues found!`（零 issue）。

- [ ] **Step 2: 检查改动范围**

```bash
git diff --stat
```

Expected: 仅 `client/pubspec.yaml`、`client/lib/data/repositories/local_store.dart`、`client/lib/core/providers/auth_provider.dart`、`client/pubspec.lock` 四个文件变更，无其他副作用。

---

### Task 5: 提测验证

按 ego-feature Phase 5 执行：

- [ ] **Step 1: flutter analyze**（Task 4 已完成）
- [ ] **Step 2: 真机测试**（用户执行）

**手动测试清单：**

| # | 场景 | 步骤 | 预期 |
|---|------|------|------|
| 1 | 登录 + 持久化 | 输入手机号密码登录 | 进入首页 |
| 2 | 重启恢复 | 杀进程重新打开 app | 自动进入首页（token 从 secure storage 恢复），不跳登录页 |
| 3 | 登出 | 设置页 → 登出 | 回到登录页 |
| 4 | 再次登录 | 登出后重新登录 | 正常进入首页 |
| 5 | 重启验证登出 | 登出后杀进程重启 | 仍停留在登录页 |

- [ ] **Step 3: Web 端测试**

```bash
cd client && flutter run -d chrome
```

同样的测试清单过一遍（需 localhost，WebCrypto 在 localhost 下可用）。

---

### Task 6: Commit

全部验证通过后一次性提交。

- [ ] **Step 1: 提交**

```bash
git add client/pubspec.yaml client/pubspec.lock \
  client/lib/data/repositories/local_store.dart \
  client/lib/core/providers/auth_provider.dart
git commit -m "feat(client): migrate token from Hive plaintext to flutter_secure_storage

Android: EncryptedSharedPreferences (Keystore)
Web: WebCrypto AES-256-GCM + localStorage
Non-sensitive data (onboarding flag, starmap guide) stays in Hive.

Co-Authored-By: Claude <noreply@anthropic.com>"
```
