# Module 1 — Login + 项目脚手架 设计文档

## 目标

做完后 app 可运行、可登录/注册、进入带底部三 Tab 的空壳主屏。一次性搭建完整的前后端基础设施链路（proto 编译、gRPC 通信、JWT 认证、DB 连接）。

## 范围

| 层 | 内容 |
|----|------|
| 协议层 | proto 编译链路（Go 端 + Dart 端）、gRPC server/client 骨架 |
| 前端 | Flutter 项目、GoRouter + StatefulShellRoute、authProvider、LoginPage、三 Tab 空壳页面 |
| 后端 | Go gRPC server、pgxpool 连接、users 表 + migration、Login RPC（JWT + bcrypt） |

涉及 RPC：`Login`。

## 协议约定

- proto 源文件：`/proto/ego/api.proto`（前后端共享单一来源）
- Go 端：`protoc` 生成到 `server/proto/ego/` 目录
- Dart 端：`protoc` 生成到 `client/lib/data/generated/` 目录
- Module 1 只在 proto 中包含 Login RPC + 相关 message（其余 RPC 后续 module 逐步添加）
- JWT 通过 gRPC metadata `Authorization: Bearer <token>` 传递
- Token 过期 30 天，任意 RPC 返回 `UNAUTHENTICATED` → 前端清 token → GoRouter redirect 到 `/login`

## 数据流

```
App 启动
  │
  ├─ authProvider 检查本地 token
  │   ├─ 无 token → GoRouter redirect → /login → LoginPage
  │   └─ 有 token → GoRouter redirect → /now → AppShell（三 Tab）
  │
  ├─ LoginPage: 输入 account + password → Ego.Login(account, password)
  │   ├─ 后端查 users 表
  │   │   ├─ 不存在 → bcrypt hash → INSERT user → 签发 JWT (created=true)
  │   │   └─ 存在 → bcrypt 验证 → 签发 JWT (created=false)
  │   └─ 前端存 token → authProvider 更新 → redirect → /now
  │
  └─ AppShell: BottomNavigationBar（此刻/过往/星图）
      ├─ Tab 0 → NowPage（空壳）
      ├─ Tab 1 → PastPage（空壳）
      └─ Tab 2 → StarmapPage（空壳）
```

## 前端文件清单

```
client/
├── pubspec.yaml
├── lib/
│   ├── main.dart                              # runApp + ProviderScope
│   ├── app.dart                               # MaterialApp.router
│   ├── core/
│   │   ├── theme/
│   │   │   ├── theme.dart                     # ThemeData 深色主题
│   │   │   └── colors.dart                    # 色板常量
│   │   ├── router/
│   │   │   └── router.dart                    # GoRouter + StatefulShellRoute + 登录守卫
│   │   ├── providers/
│   │   │   ├── auth_provider.dart             # JWT token + user_id + isLoggedIn
│   │   │   └── tab_provider.dart              # selectedTabIndex
│   │   └── constants.dart
│   ├── data/
│   │   ├── services/
│   │   │   ├── ego_client.dart                # gRPC 客户端封装 + JWT metadata 注入
│   │   │   └── interceptors/
│   │   │       └── auth_interceptor.dart       # 所有请求注入 Authorization header
│   │   ├── repositories/
│   │   │   └── local_store.dart               # Hive 本地持久化
│   │   └── generated/                         # protoc 生成
│   ├── features/
│   │   ├── login/
│   │   │   └── login_page.dart
│   │   ├── now/
│   │   │   └── now_page.dart                  # 空壳
│   │   ├── past/
│   │   │   └── past_page.dart                 # 空壳
│   │   └── starmap/
│   │       └── starmap_page.dart              # 空壳
│   └── shared/
│       └── widgets/
│           └── app_shell.dart                 # BottomNavigationBar + Tab 容器
```

## 后端文件清单

```
server/
├── cmd/ego/main.go                        # 入口
├── internal/
│   ├── config/config.go                   # 环境变量/配置
│   ├── db/
│   │   ├── postgres.go                    # pgxpool 连接池
│   │   └── migrations/
│   │       └── 001_users.sql
│   ├── auth/
│   │   ├── jwt.go                         # JWT 签发 + 解析
│   │   └── interceptor.go                 # gRPC unary interceptor
│   └── login/
│       └── handler.go                     # Login RPC
├── proto/ego/                             # protoc 生成的 Go 代码
├── go.mod
└── Makefile                               # proto gen / migrate / run
```

## 后端实现概要（伪代码）

### Step 1: 项目骨架
- `go mod init`，安装依赖（grpc, pgx, jwt, bcrypt, uuid）
- Makefile: `proto-gen`, `migrate`, `run`
- `cmd/ego/main.go`: 加载 config → 连接 pgxpool → 注册 gRPC server → Listen
- `config.go`: 环境变量读取 `DATABASE_URL`, `JWT_SECRET`, `PORT`

### Step 2: proto + gRPC 骨架
- 根目录 `proto/ego/api.proto`：Module 1 只含 Login RPC + LoginReq/LoginRes
- `make proto-gen`：protoc 生成 Go 代码到 `server/proto/ego/`
- main.go 注册 EgoServer，启动 gRPC

### Step 3: users 表 + migration
```sql
CREATE TABLE users (
  id            UUID PRIMARY KEY,
  account       VARCHAR(100) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL
);
CREATE UNIQUE INDEX idx_users_account ON users(account);
```

### Step 4: auth 包
```go
// jwt.go
func generateJWT(userID string) string {
    claims := jwt.MapClaims{
        "user_id": userID,
        "iat":     time.Now().Unix(),
        "exp":     time.Now().Add(30 * 24 * time.Hour).Unix(),
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    s, _ := token.SignedString(jwtSecret)
    return s
}

// interceptor.go
func authInterceptor(jwtSecret []byte) grpc.UnaryServerInterceptor {
    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        // Login RPC 白名单，跳过认证
        if strings.Contains(info.FullMethod, "Login") {
            return handler(ctx, req)
        }
        // 从 metadata 提取 Bearer token → parseJWT → 注入 user_id 到 ctx
        md, _ := metadata.FromIncomingContext(ctx)
        tokenStr := extractBearer(md)
        userID, err := parseJWT(tokenStr, jwtSecret)
        if err != nil { return nil, status.Error(codes.Unauthenticated, "invalid token") }
        ctx = context.WithValue(ctx, "user_id", userID)
        return handler(ctx, req)
    }
}
```

### Step 5: Login handler
```go
// login/handler.go
func (s *LoginHandler) Login(ctx context.Context, req *pb.LoginReq) (*pb.LoginRes, error) {
    row := s.db.QueryRow("SELECT id, password_hash FROM users WHERE account = $1", req.Account)

    if row == nil {
        // 自动注册
        hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
        userID := uuid.New().String()
        s.db.Exec("INSERT INTO users (id, account, password_hash, created_at) VALUES ($1, $2, $3, $4)",
            userID, req.Account, hash, time.Now())
        token := generateJWT(userID)
        return &pb.LoginRes{Token: token, Created: true}, nil
    }

    // 验证密码
    if bcrypt.CompareHashAndPassword([]byte(row["password_hash"]), []byte(req.Password)) != nil {
        return nil, status.Error(codes.Unauthenticated, "密码错误")
    }
    token := generateJWT(row["id"])
    return &pb.LoginRes{Token: token, Created: false}, nil
}
```

## 前端实现概要（伪代码）

### Step 1: 项目初始化
- `flutter create`，安装依赖（go_router, flutter_riverpod, grpc, protobuf, freezed, hive_flutter）
- proto 生成 Dart 代码到 `lib/data/generated/`
- `lib/main.dart`: `ProviderScope(child: App())`
- `lib/app.dart`: `MaterialApp.router(...)` + 深色主题

### Step 2: 基础设施层

```dart
// core/providers/auth_provider.dart
class AuthState { String? token; String? userId; bool isLoggedIn; }
class AuthNotifier extends StateNotifier<AuthState> {
  void login(String token) {
    // 持久化到 Hive
    // 更新 state: token, isLoggedIn = true
  }
  void logout() {
    // 清 Hive
    // 清 state
  }
}

// data/services/ego_client.dart
import 'generated/api.pbgrpc.dart' as grpc;
class EgoClient {
  final grpc.EgoClient _stub;
  CallOptions _withAuth(Ref ref) {
    final token = ref.read(authProvider).token;
    return CallOptions(metadata: {'Authorization': 'Bearer $token'});
  }
  Future<LoginRes> login(account, password) => _stub.login(LoginReq(account, password));
}
```

### Step 3: 路由 + AppShell

```dart
// core/router/router.dart
GoRouter(
  initialLocation: '/now',
  redirect: (context, state) {
    final loggedIn = ref.read(authProvider).isLoggedIn;
    if (!loggedIn && state.matchedLocation != '/login') return '/login';
    if (loggedIn && state.matchedLocation == '/login') return '/now';
    return null;
  },
  routes: [
    GoRoute(path: '/login', builder: (_, __) => const LoginPage()),
    StatefulShellRoute.indexedStack(
      branches: [
        StatefulShellBranch(routes: [GoRoute('/now', builder: → NowPage())]),
        StatefulShellBranch(routes: [GoRoute('/past', builder: → PastPage())]),
        StatefulShellBranch(routes: [GoRoute('/starmap', builder: → StarmapPage())]),
      ],
      builder: (context, shell) => AppShell(shell),
    ),
  ],
)

// shared/widgets/app_shell.dart
Scaffold(
  body: navigationShell,
  bottomNavigationBar: BottomNavigationBar(
    items: [此刻, 过往, 星图],
    onTap: (index) { tabProvider.setIndex(index); navigationShell.goBranch(index); },
  ),
)
```

### Step 4: 四个页面

```dart
// features/login/login_page.dart
Column(
  AppLogo,
  TextField(controller: accountCtrl),
  TextField(controller: passwordCtrl, obscureText: true),
  ElevatedButton('进入', onTap: () async {
    final res = await egoClient.login(accountCtrl.text, passwordCtrl.text);
    ref.read(authProvider.notifier).login(res.token);
    // redirect 自动跳 /now
  }),
)

// features/now/now_page.dart → Scaffold(body: Center(child: Text('有什么想说的吗')))
// features/past/past_page.dart → Scaffold(body: Center(child: Text('每一次说出口的，都留在这里')))
// features/starmap/starmap_page.dart → Scaffold(body: Center(child: Text('已有 0 颗星')))
```

## 验收标准

- [ ] `make run` 启动后端，gRPC 服务监听 :50051
- [ ] `flutter run` 启动前端，显示登录页
- [ ] 输入新 account + password → 登录成功 → 进入三 Tab 主屏（返回 `created: true`）
- [ ] 关闭 app 重新打开 → 直接进入主屏（token 有效免登录）
- [ ] 切换到"过往"/"星图" Tab → 显示空壳占位文字
- [ ] 输入错误密码 → 报错，留在登录页

## 后续 module 预览

| # | Module | 核心 RPC | 依赖 |
|---|--------|----------|------|
| 2 | 写字 + 回声 | CreateMoment | Module 1 |
| 3 | 观察 + 收进星图 | GenerateInsight, StashTrace | Module 2 |
| 4 | 记忆光点 + 时间线 | GetRandomMoments, ListMoments | Module 1 |
| 5 | 星图 + 星座详情 | ListConstellations, GetConstellation | Module 3 |
| 6 | 和那时的自己说说话 | StartChat, SendMessage | Module 5 |
