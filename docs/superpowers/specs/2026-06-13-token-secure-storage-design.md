# Design: Token 从 Hive 明文迁移到 flutter_secure_storage

**日期**: 2026-06-13
**状态**: 已确认

## 背景

ego 客户端使用 Hive 将 JWT Token 以明文写入文件系统，Android 和 Web 上均可直接读取。需提升至平台安全存储。

## 安全目标（分端）

| 平台 | 机制 | 安全级别 |
|------|------|----------|
| Android | EncryptedSharedPreferences → Keystore (TEE) | 硬件级 |
| Web | WebCrypto AES-256-GCM + localStorage | 加密混淆级（浏览器上限） |

Web 端密钥与密文共置在 localStorage，无法达到 Android 级别的安全，但比当前明文存储强——至少需要解密步骤。

## 架构：内存缓存 + 预加载

**问题**: `flutter_secure_storage` 全部异步，但 `StateNotifier` 构造函数不能 async。

**方案**: 在 `main()` 中 `LocalStore.init()` 预加载 token 到内存缓存，`getToken()` 永远读缓存保持同步，写操作改异步。

```
main()
  ├── await LocalStore.init()
  │     ├── Hive.initFlutter()              ← settings box 保留
  │     └── _cachedToken = await secureStorage.read('token')
  └── runApp(ProviderScope(...))
        └── AuthNotifier()
              ├── _loadToken()              ← getToken() 读缓存，同步
              ├── login(token) async       ← setToken() 异步写
              └── logout() async           ← clearToken() 异步删
```

## 改动文件

### 1. `client/pubspec.yaml`

新增依赖 `flutter_secure_storage: ^10.0.0`。

### 2. `client/lib/data/repositories/local_store.dart`

- 新增 `FlutterSecureStorage _secure` 实例 + `String? _cachedToken` 缓存
- `init()`: 从 secure storage 预加载 token 到缓存；保留 Hive settings box
- 移除 `_auth` Hive box
- `getToken()` → 同步返回缓存
- `setToken(token)` → async: `_secure.write()` + 更新缓存
- `clearToken()` → async: `_secure.delete()` + 清缓存
- onboarding/starmap 方法保持 Hive 不变

### 3. `client/lib/core/providers/auth_provider.dart`

- `login()`: `void` → `Future<void> async`，`await LocalStore.setToken(token)`
- `logout()`: `void` → `Future<void> async`，`await LocalStore.clearToken()`
- `_loadToken()` / 构造函数不变

### 4. 不变文件

- `main.dart` — 已有 `await LocalStore.init()`
- `router.dart` — `ref.watch(authProvider)` 响应式
- `ego_client.dart` — 从 `authProvider.token` 读
- `onboarding_provider.dart` — 非敏感，留 Hive
- 所有 page/widget — `login()`/`logout()` fire-and-forget 兼容 `Future<void>`

## 降级行为

- 读失败 → token 为 null → 用户重新登录
- 写失败 → 未持久化 → 下次启动需重新登录

## 迁移

旧 Hive 中 token 不迁移，用户升级后重新登录。

## 约束

- Web 必须 HTTPS（WebCrypto API 要求），localhost 除外

## 验证

1. `flutter pub get` — 依赖安装
2. `flutter analyze` — 零 issue
3. 真机测试：
   - 登录 → 写 token → 杀进程重启 → 自动登录（token 从 secure storage 恢复）
   - 登出 → token 清除 → 回到登录页
   - Web 端同流程
