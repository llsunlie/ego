# ego · 前端目录结构

> Flutter · Dart · Riverpod · gRPC

## 1. 组织原则：按特性分包，薄耦合

```
      按层次分（低内聚）                   按特性分（高内聚）
      ─────────────────                   ─────────────────
      pages/now/                          features/now/
      pages/past/                         ├── now_page.dart
      providers/now_page.dart             ├── providers/
      widgets/starry_background.dart      └── widgets/
      services/echo_service.dart              ├── breathing_light.dart
      models/moment.dart                      ├── echo_card.dart
                                              └── ...
      ↑ 改 NowPage 要在 5 个目录间跳         ↑ 一个特性自包含

      前后端共通：领域包/特性包 = handler + 状态 + UI + 查询，垂直切分
```

每个特性包内聚其页面、Widget、Provider、本地存储——外部只通过包内暴露的 Provider 和页面入口交互。

## 2. 目录树

```
lib/
├── main.dart                              # runApp + ProviderScope
├── app.dart                               # MaterialApp.router（GoRouter + Theme + Interceptors）
├── core/
│   ├── theme/
│   │   ├── theme.dart                     # ThemeData（深色主题）
│   │   └── colors.dart                    # 色板常量（暖金/冷蓝/柔紫/金色）
│   ├── router/
│   │   └── router.dart                    # GoRouter 配置 + 登录守卫 redirect
│   ├── providers/
│   │   ├── auth_provider.dart             # JWT token + user_id + isLoggedIn
│   │   └── tab_provider.dart              # 当前 selectedTabIndex
│   └── constants.dart                     # 动画时长、星星数量等常量
├── data/
│   ├── models/                            # freezed 实体（与 proto message 对应）
│   │   ├── moment.dart
│   │   ├── echo.dart
│   │   ├── insight.dart
│   │   ├── star.dart
│   │   ├── constellation.dart
│   │   ├── topic_prompt.dart
│   │   ├── past_self_card.dart
│   │   ├── chat_message.dart
│   │   └── login_result.dart
│   ├── services/
│   │   ├── ego_client.dart                # gRPC 客户端封装 + JWT metadata 注入
│   │   └── interceptors/
│   │       └── auth_interceptor.dart       # 所有请求注入 Authorization header
│   ├── repositories/
│   │   └── local_store.dart               # Hive 本地持久化（token 缓存、moment 缓存）
│   └── generated/                         # protoc 生成的 Dart 文件（只读）
│       ├── api.pb.dart
│       ├── api.pbenum.dart
│       └── api.pbgrpc.dart
├── features/
│   ├── login/
│   │   └── login_page.dart                # 登录/自动注册页面
│   ├── now/
│   │   ├── now_page.dart                  # 此刻主页（Stack 编排光团 + 写字区 + 卡片 + 动效）
│   │   ├── providers/
│   │   │   ├── now_page_provider.dart     # 状态机：idle → writing → echoing → stashing
│   │   │   └── memory_dots_provider.dart  # 3 颗记忆光点数据（GetRandomMoments）
│   │   └── widgets/
│   │       ├── breathing_light.dart       # CustomPaint + shader 不规则光团
│   │       ├── starry_background.dart     # CustomPainter 星空背景
│   │       ├── memory_dot.dart            # 记忆光点（漂浮 + 展开信封）
│   │       ├── writing_input.dart         # 写字区（TextField + 提交按钮）
│   │       ├── echo_card.dart             # 回声卡（匹配结果 + 候选）
│   │       ├── insight_card.dart          # 观察卡（✦ 我发现）
│   │       └── stash_animation.dart       # 仪式感动效（6 段编排 Overlay）
│   ├── past/
│   │   ├── past_page.dart                 # 过往时间线主页
│   │   ├── providers/
│   │   │   └── timeline_provider.dart     # ListMoments 游标分页
│   │   └── widgets/
│   │       ├── month_section.dart         # 按月分组标题
│   │       └── moment_item.dart           # 单条 Moment（展开/折叠）
│   └── starmap/
│       ├── starmap_page.dart              # 星图主页（InteractiveViewer）
│       ├── providers/
│       │   └── constellation_provider.dart # ListConstellations + detail family
│       ├── detail/
│       │   ├── detail_page.dart           # 星座详情页（Scaffold + BackButton）
│       │   ├── providers/
│       │   │   └── detail_provider.dart   # GetConstellation（按 id）
│       │   └── widgets/
│       │       ├── insight_section.dart   # ① ✦ 我发现（ExpansionTile 原话列表）
│       │       ├── chat_section.dart      # ② 和那时的自己说说话（PastSelfCard × 3）
│       │       ├── chat_dialog.dart       # 对话浮层（showModalBottomSheet）
│       │       └── topic_prompt_section.dart # ③ ✦ 我想和你聊聊（TopicPromptCard × 3）
│       └── painters/
│           ├── star_field_painter.dart    # CustomPainter 星图（星星 + 连线 + 标签）
│           └── nebula_painter.dart        # 星云背景渐变
└── shared/
    ├── widgets/
    │   ├── star_painter.dart              # 可复用星星绘制（亮星/脉冲/暗星）
    │   └── shimmer_card.dart              # 骨架屏/加载占位
    └── extensions/
        └── date_format.dart               # DateTime → "X月X日" / "2月-3月" 格式
```

## 3. 特性包内部结构

每个 `features/<name>/` 自包含：

```
feature/
├── <name>_page.dart        # 页面入口 Widget
├── providers/              # 本特性 Riverpod Provider（不被其他特性引用）
└── widgets/                # 本特性私有 Widget
```

**约束：**
- Provider 只在本特性内使用，跨特性共享提至 `core/providers/`
- Widget 不跨特性 import，通用 UI 放 `shared/widgets/`
- 页面之间通过路由传参（`GoRouter.extra`），不直接传 Provider
- 所有 gRPC 调用经 `data/services/ego_client.dart`，特性不直接创建 grpc Client

## 4. Provider 分层

```
┌─────────────────────────────────────────────┐
│  core/providers/        （app 级全局常驻）    │
│  - authProvider          JWT / isLoggedIn    │
│  - tabProvider           selectedTabIndex    │
├─────────────────────────────────────────────┤
│  features/*/providers/   （特性级，AutoDispose）│
│  - nowPageProvider       idle|writing|ech..  │
│  - memoryDotsProvider    GetRandomMoments    │
│  - timelineProvider      ListMoments 分页    │
│  - constellationProvider ListConstellations  │
│  - detailProvider        GetConstellation    │
│  - chatProvider          SendMessage 会话    │
├─────────────────────────────────────────────┤
│  Widget-local State      （setState）        │
│  - TextField 文本                           │
│  - AnimationController                      │
│  - ExpansionTile 状态                       │
└─────────────────────────────────────────────┘
```

## 5. 跨特性交互

特性间不直接 import，通过以下方式解耦：

| 交互 | 方式 |
|------|------|
| NowPage → PastPage | 写入 Moment 后 PastPage 重新 ListMoments（Provider 刷新） |
| NowPage → StarmapPage | StashTrace 后失效 constellationProvider 缓存 |
| DetailPage → NowPage | 话题引子：`Navigator.pop() → switchTab(0)`，通过路由 extra 携带 topic |
| LoginPage → App | `authProvider.login(token)` → GoRouter redirect 自动跳 /now |

## 6. 依赖图

```
                       ┌──────────┐
                       │   main   │
                       └────┬─────┘
                  ┌─────────┼─────────┐
                  ▼         ▼         ▼
           ┌─────────┐ ┌────────┐ ┌──────────┐
           │  core/   │ │ data/  │ │ shared/  │  ← 基础设施（无特性依赖）
           │ router   │ │ models │ │ widgets  │
           │ theme    │ │ services│           │
           │ providers│ │ repos. │           │
           └────┬─────┘ └───┬────┘ └──────────┘
                │            │
    ┌───────────┼────────────┼──────────────────┐
    ▼           ▼            ▼                  ▼
┌────────┐ ┌────────┐ ┌──────────┐      ┌──────────┐
│ login  │ │  now   │ │  past    │      │ starmap  │
└────────┘ └────────┘ └──────────┘      └──────────┘
                                                 │
                                            ┌────┴────┐
                                            │  detail  │
                                            │  chat    │
                                            └─────────┘
```
- 特性间互不 import，通过路由 + Provider 缓存失效通信
- `data/` 是唯一有 gRPC 依赖的层，特性通过 `ego_client.dart` 调后端
- `core/providers/auth_provider.dart` 是唯一被所有特性依赖的 Provider（注入 token）

## 7. gRPC 客户端 + JWT 拦截器

```dart
// data/services/ego_client.dart
// 注意：生成的 gRPC stub 类名也是 EgoClient，需通过 import prefix 区分
import 'generated/api.pbgrpc.dart' as grpc;

class EgoClient {
  final grpc.EgoClient _stub;

  EgoClient(this._stub);

  static final instance = Provider<EgoClient>((ref) {
    final channel = ClientChannel(
      'localhost',
      port: 50051,
      options: const ChannelOptions(credentials: ChannelCredentials.insecure()),
    );
    return EgoClient(grpc.EgoClient(channel));
  });

  // 每次调用注入 Authorization metadata
  CallOptions _withAuth(Ref ref) {
    final token = ref.read(authProvider).token;
    return CallOptions(
      metadata: {'Authorization': 'Bearer $token'},
    );
  }
}
```

## 8. 主要 Flutter 依赖

```
用途              包
─────────────────────────────────────────
路由               go_router
状态管理            flutter_riverpod
数据类              freezed + freezed_annotation
JSON 序列化         json_serializable
gRPC               grpc + protobuf
本地存储            hive_flutter
动画                flutter_animate
```

## 9. proto 共享

```
项目根/
├── proto/
│   └── ego/
│       └── api.proto          ← 单一来源（前后端共享）
├── lib/                       ← Flutter 前端
│   └── data/generated/        ← protoc 生成（git ignore）
└── server/                    ← Go 后端
    └── proto/ego/api.proto    ← symlink → ../../proto/ego/api.proto
```
