# Module 1 Frontend — Login + 项目脚手架 实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 搭建 Flutter 前端骨架，实现登录页 + 三 Tab 空壳主屏，打通 gRPC 通信与 JWT 认证。

**Architecture:** GoRouter + StatefulShellRoute 管理路由，Riverpod 管理状态，Hive 持久化 token，gRPC client 注入 JWT metadata，登录守卫控制页面访问。

**Tech Stack:** Flutter 3.x, Dart, go_router, flutter_riverpod, grpc-dart, protobuf, freezed, hive_flutter

---

### Task 1: Flutter 项目初始化 + 依赖安装

- [ ] **Step 1: 创建 Flutter 项目（在项目根目录）**

```bash
flutter create --org com.ego --project-name ego .
```

- [ ] **Step 2: 添加依赖到 pubspec.yaml**

```yaml
# pubspec.yaml 主要依赖
dependencies:
  flutter:
    sdk: flutter
  go_router: ^14.0.0
  flutter_riverpod: ^2.5.0
  grpc: ^4.0.0
  protobuf: ^3.1.0
  fixnum: ^1.1.0
  freezed_annotation: ^2.4.0
  json_annotation: ^4.9.0
  hive_flutter: ^1.1.0

dev_dependencies:
  flutter_test:
    sdk: flutter
  build_runner: ^2.4.0
  freezed: ^2.5.0
  json_serializable: ^6.8.0
```

- [ ] **Step 3: 安装依赖**

```bash
flutter pub get
```

- [ ] **Step 4: 创建基础目录结构**

```bash
mkdir -p lib/core/theme
mkdir -p lib/core/router
mkdir -p lib/core/providers
mkdir -p lib/data/services/interceptors
mkdir -p lib/data/repositories
mkdir -p lib/data/generated
mkdir -p lib/features/login
mkdir -p lib/features/now
mkdir -p lib/features/past
mkdir -p lib/features/starmap
mkdir -p lib/shared/widgets
```

- [ ] **Step 5: 确认 proto 文件存在（已在 Task 2 后端中创建）**

```bash
ls proto/ego/api.proto
```
Expected: 文件存在

- [ ] **Step 6: 生成 Dart proto 代码**

```bash
protoc --dart_out=grpc:lib/data/generated \
       --proto_path=proto \
       proto/ego/api.proto
```

- [ ] **Step 7: 添加 `client/lib/data/generated/` 到 .gitignore**

```
lib/data/generated/
```

---

### Task 2: 主题 + 色板常量

**Files:**
- Create: `client/lib/core/theme/colors.dart`
- Create: `client/lib/core/theme/theme.dart`
- Create: `client/lib/core/constants.dart`

- [ ] **Step 1: 编写色板常量**

```dart
// lib/core/theme/colors.dart
import 'package:flutter/material.dart';

class AppColors {
  // 主色调
  static const Color gold = Color(0xFFD4A853);
  static const Color warmGold = Color(0xFFE8C97A);

  // 功能色
  static const Color coldBlue = Color(0xFF7B9EC7);
  static const Color softPurple = Color(0xFF9B8EC4);

  // 背景
  static const Color darkBg = Color(0xFF0D0D1A);
  static const Color surface = Color(0xFF1A1A2E);

  // 文字
  static const Color textPrimary = Color(0xFFF0EDE5);
  static const Color textSecondary = Color(0xFF8B8B9E);
}
```

- [ ] **Step 2: 编写深色主题**

```dart
// lib/core/theme/theme.dart
import 'package:flutter/material.dart';
import 'colors.dart';

ThemeData darkTheme() {
  return ThemeData(
    brightness: Brightness.dark,
    scaffoldBackgroundColor: AppColors.darkBg,
    colorScheme: ColorScheme.dark(
      primary: AppColors.gold,
      secondary: AppColors.coldBlue,
      surface: AppColors.surface,
    ),
    textTheme: const TextTheme(
      bodyLarge: TextStyle(color: AppColors.textPrimary, fontSize: 16),
      bodyMedium: TextStyle(color: AppColors.textSecondary, fontSize: 14),
    ),
    inputDecorationTheme: InputDecorationTheme(
      filled: true,
      fillColor: AppColors.surface,
      border: OutlineInputBorder(
        borderRadius: BorderRadius.circular(12),
        borderSide: BorderSide.none,
      ),
      hintStyle: const TextStyle(color: AppColors.textSecondary),
    ),
    elevatedButtonTheme: ElevatedButtonThemeData(
      style: ElevatedButton.styleFrom(
        backgroundColor: AppColors.gold,
        foregroundColor: Colors.black,
        minimumSize: const Size(double.infinity, 52),
        shape: RoundedRectangleBorder(
          borderRadius: BorderRadius.circular(12),
        ),
      ),
    ),
  );
}
```

- [ ] **Step 3: 编写常量**

```dart
// lib/core/constants.dart
class AppConstants {
  static const String serverHost = 'localhost';
  static const int serverPort = 50051;
  static const Duration animationDuration = Duration(milliseconds: 300);
}
```

---

### Task 3: Auth Provider + Hive 本地持久化

**Files:**
- Create: `client/lib/data/repositories/local_store.dart`
- Create: `client/lib/core/providers/auth_provider.dart`

- [ ] **Step 1: 编写 Hive 本地存储**

```dart
// lib/data/repositories/local_store.dart
import 'package:hive_flutter/hive_flutter.dart';

class LocalStore {
  static const _boxName = 'auth';
  static late Box _box;

  static Future<void> init() async {
    await Hive.initFlutter();
    _box = await Hive.openBox(_boxName);
  }

  static String? getToken() => _box.get('token');
  static void setToken(String token) => _box.put('token', token);
  static void clearToken() => _box.delete('token');
}
```

- [ ] **Step 2: 编写 AuthState + AuthNotifier**

```dart
// lib/core/providers/auth_provider.dart
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../data/repositories/local_store.dart';

class AuthState {
  final String? token;
  final bool isLoggedIn;

  const AuthState({this.token, this.isLoggedIn = false});

  AuthState copyWith({String? token, bool? isLoggedIn}) {
    return AuthState(
      token: token ?? this.token,
      isLoggedIn: isLoggedIn ?? this.isLoggedIn,
    );
  }
}

class AuthNotifier extends StateNotifier<AuthState> {
  AuthNotifier() : super(const AuthState()) {
    _loadToken();
  }

  void _loadToken() {
    final token = LocalStore.getToken();
    if (token != null) {
      state = AuthState(token: token, isLoggedIn: true);
    }
  }

  void login(String token) {
    LocalStore.setToken(token);
    state = AuthState(token: token, isLoggedIn: true);
  }

  void logout() {
    LocalStore.clearToken();
    state = const AuthState();
  }
}

final authProvider = StateNotifierProvider<AuthNotifier, AuthState>((ref) {
  return AuthNotifier();
});
```

---

### Task 4: gRPC 客户端封装 + JWT 拦截器

**Files:**
- Create: `client/lib/data/services/interceptors/auth_interceptor.dart`
- Create: `client/lib/data/services/ego_client.dart`

- [ ] **Step 1: 编写 JWT 拦截器**

```dart
// lib/data/services/interceptors/auth_interceptor.dart
import 'package:grpc/grpc.dart';

/// 为 gRPC 请求注入 Authorization: Bearer <token> metadata
CallOptions authCallOptions(String? token) {
  if (token == null) return CallOptions();
  return CallOptions(
    metadata: {'Authorization': 'Bearer $token'},
  );
}
```

- [ ] **Step 2: 编写 gRPC 客户端封装**

```dart
// lib/data/services/ego_client.dart
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:grpc/grpc.dart';
import '../generated/api.pbgrpc.dart' as grpc;
import 'interceptors/auth_interceptor.dart';
import '../../core/constants.dart';
import '../../core/providers/auth_provider.dart';

class EgoClient {
  final grpc.EgoClient _stub;

  EgoClient(this._stub);

  static final provider = Provider<EgoClient>((ref) {
    final channel = ClientChannel(
      AppConstants.serverHost,
      port: AppConstants.serverPort,
      options: const ChannelOptions(
        credentials: ChannelCredentials.insecure(),
      ),
    );
    return EgoClient(grpc.EgoClient(channel));
  });

  /// 从 Riverpod ref 读取 token，生成带 Authorization 的 CallOptions
  CallOptions _withAuth(Ref ref) {
    final token = ref.read(authProvider).token;
    return authCallOptions(token);
  }

  Future<grpc.LoginRes> login(String account, String password) async {
    final req = grpc.LoginReq(account: account, password: password);
    return _stub.login(req);
  }
}
```

> **注意：** `grpc.EgoClient` 是 protoc 生成的 gRPC stub 类名，与我们封装的 `EgoClient` 同名。通过 `import ... as grpc` 前缀区分。`_withAuth` 在 Module 1 不会被调用（Login 不需要认证），为 Module 2 做准备。

---

### Task 5: 路由配置 + AppShell（BottomNavigationBar）

**Files:**
- Create: `client/lib/core/router/router.dart`
- Create: `client/lib/core/providers/tab_provider.dart`
- Create: `client/lib/shared/widgets/app_shell.dart`

- [ ] **Step 1: 编写 TabProvider**

```dart
// lib/core/providers/tab_provider.dart
import 'package:flutter_riverpod/flutter_riverpod.dart';

class TabNotifier extends StateNotifier<int> {
  TabNotifier() : super(0);
  void setIndex(int index) => state = index;
}

final tabProvider = StateNotifierProvider<TabNotifier, int>((ref) {
  return TabNotifier();
});
```

- [ ] **Step 2: 编写 AppShell**

```dart
// lib/shared/widgets/app_shell.dart
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../core/providers/tab_provider.dart';

class AppShell extends ConsumerWidget {
  final StatefulNavigationShell navigationShell;

  const AppShell(this.navigationShell, {super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final tabIndex = ref.watch(tabProvider);

    return Scaffold(
      body: navigationShell,
      bottomNavigationBar: BottomNavigationBar(
        currentIndex: tabIndex,
        onTap: (index) {
          ref.read(tabProvider.notifier).setIndex(index);
          navigationShell.goBranch(index);
        },
        items: const [
          BottomNavigationBarItem(
            icon: Icon(Icons.wb_sunny_outlined),
            activeIcon: Icon(Icons.wb_sunny),
            label: '此刻',
          ),
          BottomNavigationBarItem(
            icon: Icon(Icons.history),
            activeIcon: Icon(Icons.history_toggle_off),
            label: '过往',
          ),
          BottomNavigationBarItem(
            icon: Icon(Icons.auto_awesome),
            activeIcon: Icon(Icons.auto_awesome),
            label: '星图',
          ),
        ],
      ),
    );
  }
}
```

- [ ] **Step 3: 编写 GoRouter 配置**

```dart
// lib/core/router/router.dart
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../core/providers/auth_provider.dart';
import '../../features/login/login_page.dart';
import '../../features/now/now_page.dart';
import '../../features/past/past_page.dart';
import '../../features/starmap/starmap_page.dart';
import '../../shared/widgets/app_shell.dart';

final routerProvider = Provider<GoRouter>((ref) {
  final authState = ref.watch(authProvider);

  return GoRouter(
    initialLocation: '/now',
    redirect: (context, state) {
      final loggedIn = authState.isLoggedIn;
      final isLoginPage = state.matchedLocation == '/login';

      if (!loggedIn && !isLoginPage) return '/login';
      if (loggedIn && isLoginPage) return '/now';
      return null;
    },
    routes: [
      GoRoute(
        path: '/login',
        builder: (context, state) => const LoginPage(),
      ),
      StatefulShellRoute.indexedStack(
        builder: (context, state, navigationShell) =>
            AppShell(navigationShell),
        branches: [
          StatefulShellBranch(
            routes: [
              GoRoute(
                path: '/now',
                builder: (context, state) => const NowPage(),
              ),
            ],
          ),
          StatefulShellBranch(
            routes: [
              GoRoute(
                path: '/past',
                builder: (context, state) => const PastPage(),
              ),
            ],
          ),
          StatefulShellBranch(
            routes: [
              GoRoute(
                path: '/starmap',
                builder: (context, state) => const StarmapPage(),
              ),
            ],
          ),
        ],
      ),
    ],
  );
});
```

---

### Task 6: LoginPage

**Files:**
- Create: `client/lib/features/login/login_page.dart`

- [ ] **Step 1: 编写 LoginPage**

```dart
// lib/features/login/login_page.dart
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../core/providers/auth_provider.dart';
import '../../data/services/ego_client.dart';

class LoginPage extends ConsumerStatefulWidget {
  const LoginPage({super.key});

  @override
  ConsumerState<LoginPage> createState() => _LoginPageState();
}

class _LoginPageState extends ConsumerState<LoginPage> {
  final _accountCtrl = TextEditingController();
  final _passwordCtrl = TextEditingController();
  bool _loading = false;
  String? _error;

  @override
  void dispose() {
    _accountCtrl.dispose();
    _passwordCtrl.dispose();
    super.dispose();
  }

  Future<void> _login() async {
    setState(() {
      _loading = true;
      _error = null;
    });

    try {
      final client = ref.read(EgoClient.provider);
      final res = await client.login(
        _accountCtrl.text.trim(),
        _passwordCtrl.text,
      );
      ref.read(authProvider.notifier).login(res.token);
      // GoRouter redirect 自动跳 /now
    } catch (e) {
      setState(() {
        _error = e.toString().contains('密码错误')
            ? '密码错误'
            : '登录失败，请重试';
      });
    } finally {
      if (mounted) {
        setState(() => _loading = false);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.symmetric(horizontal: 32),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                // Logo 占位
                Container(
                  width: 80,
                  height: 80,
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    border: Border.all(
                      color: Theme.of(context).colorScheme.primary,
                      width: 2,
                    ),
                  ),
                  child: Center(
                    child: Text(
                      'ego',
                      style: TextStyle(
                        fontSize: 24,
                        color: Theme.of(context).colorScheme.primary,
                      ),
                    ),
                  ),
                ),
                const SizedBox(height: 48),
                // 输入框
                TextField(
                  controller: _accountCtrl,
                  decoration: const InputDecoration(
                    hintText: '账号',
                    prefixIcon: Icon(Icons.person_outline),
                  ),
                  textInputAction: TextInputAction.next,
                ),
                const SizedBox(height: 16),
                TextField(
                  controller: _passwordCtrl,
                  obscureText: true,
                  decoration: const InputDecoration(
                    hintText: '密码',
                    prefixIcon: Icon(Icons.lock_outline),
                  ),
                  textInputAction: TextInputAction.done,
                  onSubmitted: (_) => _login(),
                ),
                if (_error != null) ...[
                  const SizedBox(height: 16),
                  Text(
                    _error!,
                    style: const TextStyle(color: Colors.redAccent),
                  ),
                ],
                const SizedBox(height: 32),
                // 登录按钮
                ElevatedButton(
                  onPressed: _loading ? null : _login,
                  child: _loading
                      ? const SizedBox(
                          height: 20,
                          width: 20,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : const Text('进入', style: TextStyle(fontSize: 16)),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
```

---

### Task 7: 三个空壳页面

**Files:**
- Create: `client/lib/features/now/now_page.dart`
- Create: `client/lib/features/past/past_page.dart`
- Create: `client/lib/features/starmap/starmap_page.dart`

- [ ] **Step 1: NowPage 空壳**

```dart
// lib/features/now/now_page.dart
import 'package:flutter/material.dart';

class NowPage extends StatelessWidget {
  const NowPage({super.key});

  @override
  Widget build(BuildContext context) {
    return const Scaffold(
      body: Center(
        child: Text(
          '有什么想说的吗',
          style: TextStyle(fontSize: 18),
        ),
      ),
    );
  }
}
```

- [ ] **Step 2: PastPage 空壳**

```dart
// lib/features/past/past_page.dart
import 'package:flutter/material.dart';

class PastPage extends StatelessWidget {
  const PastPage({super.key});

  @override
  Widget build(BuildContext context) {
    return const Scaffold(
      body: Center(
        child: Text(
          '每一次说出口的，都留在这里',
          style: TextStyle(fontSize: 18),
        ),
      ),
    );
  }
}
```

- [ ] **Step 3: StarmapPage 空壳**

```dart
// lib/features/starmap/starmap_page.dart
import 'package:flutter/material.dart';

class StarmapPage extends StatelessWidget {
  const StarmapPage({super.key});

  @override
  Widget build(BuildContext context) {
    return const Scaffold(
      body: Center(
        child: Text(
          '已有 0 颗星',
          style: TextStyle(fontSize: 18),
        ),
      ),
    );
  }
}
```

---

### Task 8: App 入口 — main.dart + app.dart

**Files:**
- Create: `client/lib/app.dart`
- Modify: `client/lib/main.dart`

- [ ] **Step 1: 编写 app.dart**

```dart
// lib/app.dart
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'core/router/router.dart';
import 'core/theme/theme.dart';

class App extends ConsumerWidget {
  const App({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final router = ref.watch(routerProvider);

    return MaterialApp.router(
      title: 'ego',
      theme: darkTheme(),
      routerConfig: router,
      debugShowCheckedModeBanner: false,
    );
  }
}
```

- [ ] **Step 2: 编写 main.dart**

```dart
// lib/main.dart
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:hive_flutter/hive_flutter.dart';
import 'app.dart';
import 'data/repositories/local_store.dart';

void main() async {
  WidgetsFlutterBinding.ensureInitialized();

  // 初始化 Hive 并加载本地 token
  await LocalStore.init();

  runApp(
    const ProviderScope(
      child: App(),
    ),
  );
}
```

- [ ] **Step 3: 验证编译**

```bash
flutter pub get && flutter analyze
```
Expected: 无编译错误

---

### Task 9: 端到端验证

- [ ] **Step 1: 确保后端运行**

```bash
# 终端 1: 启动后端
cd server && make run
```
Expected: `gRPC server listening on :50051`

- [ ] **Step 2: 启动 Flutter 前端**

```bash
# 终端 2:
flutter run
```

- [ ] **Step 3: 验证登录页显示**

Expected:
- App 启动后显示登录页
- 深色背景
- Logo 占位圆圈 + "ego" 文字
- 账号输入框、密码输入框
- "进入" 按钮

- [ ] **Step 4: 验证自动注册 + 登录**

操作:
1. 输入新账号 "testuser" + 密码 "123456"
2. 点击 "进入"

Expected:
- 请求发送到后端
- 自动注册成功（`created: true`）
- 自动跳转到 `/now`，显示 "有什么想说的吗"
- 底部三 Tab 栏可见："此刻" | "过往" | "星图"

- [ ] **Step 5: 验证免登录（token 持久化）**

操作: 关闭 app，重新打开

Expected:
- 直接进入 `/now` 主屏（不经过登录页）
- token 从 Hive 正确加载

- [ ] **Step 6: 验证 Tab 切换**

操作: 点击底部 Tab 切换

Expected:
- "此刻" → "有什么想说的吗"
- "过往" → "每一次说出口的，都留在这里"
- "星图" → "已有 0 颗星"

- [ ] **Step 7: 验证错误密码**

操作:
1. 退出登录（清 token）或清除 app 数据
2. 输入 "testuser" + 错误密码 "wrong"
3. 点击 "进入"

Expected:
- 显示 "密码错误"
- 留在登录页

---

## 验收标准

- [ ] `flutter run` 启动前端，显示登录页
- [ ] 输入新 account + password → 登录成功 → 进入三 Tab 主屏
- [ ] 关闭 app 重新打开 → 直接进入主屏（token 有效免登录）
- [ ] 切换到"过往"/"星图" Tab → 显示空壳占位文字
- [ ] 输入错误密码 → 提示错误，留在登录页
- [ ] 深色主题应用正确，金色主色调
