# ego · 前端信息架构（Information Architecture）

> 技术栈：Flutter · Dart

## 1. 导航与路由

### 1.1 路由结构

```
                         LoginPage (/login)
                              │
                              ▼
                        AppShell (GoRouter + StatefulShellRoute)
                                  │
              ┌───────────────────┼───────────────────┐
              ▼                   ▼                   ▼
      StatefulShellBranch  StatefulShellBranch  StatefulShellBranch
      index: 0             index: 1             index: 2
              │                   │                   │
              ▼                   ▼                   ▼
         ┌──────────┐       ┌──────────┐       ┌──────────┐
         │  NowPage │       │ PastPage │       │StarmapPage│
         │  /now    │       │ /past    │       │/starmap   │
         └──────────┘       └──────────┘       └─────┬────┘
              │                                      │
              │ overlay（页面内 Stack 浮层）          │ push
              ▼                                      ▼
         ┌──────────┐                       ┌─────────────────┐
         │ 写字区    │                       │DetailPage        │
         │ 回声卡    │                       │/starmap/detail   │
         │ 观察卡    │                       │(左上角 BackButton)│
         │ 寄存动效  │                       └────────┬────────┘
         └──────────┘                                │
                                              ┌──────┴──────┐
                                              ▼              ▼
                                        ChatDialog     TopicPrompt
                                        (showModalBottomSheet) → pop 回到 NowPage
```

### 1.2 GoRouter 路由表

```
路由路径                        页面                        导航方式
───────────────────────────────────────────────────────────────────────
/login                         LoginPage                    初始路由，登录后 redirect 到 /now
/now                           NowPage                      Tab 切换（StatefulShellRoute 保活）
/past                          PastPage                     Tab 切换（StatefulShellRoute 保活）
/starmap                       StarmapPage                  Tab 切换（StatefulShellRoute 保活）
/starmap/detail                ConstellationDetailPage      context.push，左上角 back
```

### 1.3 GoRouter 核心配置思路

```dart
// 四层路由结构 + 登录守卫
GoRouter(
  initialLocation: '/now',
  redirect: (context, state) {
    final loggedIn = AuthProvider.of(context).isLoggedIn;
    final isLoginRoute = state.matchedLocation == '/login';

    if (!loggedIn && !isLoginRoute) return '/login';
    if (loggedIn && isLoginRoute) return '/now';
    return null;
  },
  routes: [
    GoRoute(path: '/login', builder: (_, __) => const LoginPage()),
    StatefulShellRoute.indexedStack(       // Tab 保活，切回不重建
      branches: [
        StatefulShellBranch(routes: [GoRoute(path: '/now', builder: ...)]),
        StatefulShellBranch(routes: [GoRoute(path: '/past', builder: ...)]),
        StatefulShellBranch(routes: [
          GoRoute(path: '/starmap', builder: ...),
          GoRoute(path: '/starmap/detail', builder: ...),  // push 路由，不在 tab 内
        ]),
      ],
      builder: (context, navigationShell) => AppShell(navigationShell: navigationShell),
    ),
  ],
);
```

### 1.4 浮层策略

写字区、回声卡、观察卡、寄存动效、对话模式均为页面内浮层，不走路由：

| 浮层 | 实现方式 | 原因 |
|------|----------|------|
| 写字区 | `AnimatedPositioned` + `AnimatedOpacity`，在 NowPage 的 Stack 内 | 与光团联动缩小上移动画 |
| 回声卡 / 观察卡 | `AnimatedPositioned` + staggered `AnimatedSlide` | 提交后依次浮出 |
| 寄存仪式感动效 | `Overlay` / 全屏 `Stack`，`AnimationController` 编排 | 需跨组件飞行（卡片→Tab），独立动画层 |
| 对话模式 | `showModalBottomSheet`（`DraggableScrollableSheet`） | 覆盖 85%，半模态交互 |
| 信封卡（记忆光点） | `OverlayEntry` / `showDialog` 无遮罩 | 点击光点展开，点击外部收回 |

---

## 2. Widget 树

### 2.0 LoginPage — 登录

```
LoginPage (ConsumerStatefulWidget)
└── Column
    ├── AppLogo                       # 品牌标识（星空 + 光团）
    ├── AccountField                  # TextField(account)
    ├── PasswordField                 # TextField(password, obscureText)
    └── LoginButton                   # "进入" → Ego.Login → 存 token → 跳转
        └── onTap:
            ├─ Ego.Login(account, password) → token
            ├─ 本地持久化 token（Hive/SharedPreferences）
            └─ router.go('/now')
```

### 2.1 NowPage — 此刻

```
NowPage (ConsumerStatefulWidget)
└── Stack
    ├── StarryBackground            # CustomPaint + CustomPainter，随机小星闪烁
    ├── BreathingLight              # CustomPaint，shader 驱动的不规则形变光团
    │   └── status: idle | writing | echoing
    │       → idle: 全屏居中，大口呼吸
    │       → writing: 缩小上移 (AnimatedScale + AnimatedSlide)
    │       → echoing: 保持 writing 位置
    ├── MemoryDotGroup              # 3 颗记忆光点浮动在光团周围
    │   └── MemoryDot × 3           # AnimatedPositioned 漂浮 + GestureDetector(onTap 展开)
    │       └── onTap → 展开 EnvelopeCard（OverlayEntry）
    │           ├── DateLabel
    │           └── ContentText     # 过去某天的一句原话
    ├── GuideText                   # "有什么想说的吗"（AnimatedOpacity，写字前可见）
    ├── WriteButton                 # "写下此刻"（AnimatedOpacity，写字前可见）
    │   └── onTap → setStatus(writing)
    ├── WritingInput                # AnimatedPositioned 从底部滑出
    │   ├── TextField               # placeholder: "随便说点什么，这里听着……"
    │   └── SubmitButton            # "先到这" → 触发提交 + 状态 → echoing
    ├── EchoCard                    # AnimatedSlide 浮出（stagger index: 0）
    │   ├── TagLabel                # "你之前也说过类似的"
    │   ├── EchoContent             # 匹配到的历史 Moment.content
    │   ├── CandidateToggle         # "之前的你还说过 ›" 展开 2-3 条候选
    │   │   └── EchoCandidate × N   # AnimatedSize 展开
    │   └── EchoActions
    │       ├── ContinueButton      # "顺着再想想" → 金色主按钮
    │       ├── StashButton         # "✦ 收进星图" → 触发 StashAnimation
    │       └── ExitButton          # "嗯，先这样" → 重置状态到 idle
    ├── InsightCard                 # AnimatedSlide 浮出（stagger index: 1）
    │   └── "✦ 我发现" 文本
    └── StashAnimation              # 全屏 Overlay，AnimationController 编排 6 段动效
```

**NowPage 状态机：**

```
idle ──[点"写下此刻"]──▶ writing ──[点"先到这"]──▶ echoing ──[点"顺着再想想"]──▶ writing（循环）
                          │                            │
                          │                            ├──[点"收进星图"]──▶ stashing → idle
                          │                            │
                          │                            └──[点"先这样"]──▶ idle
                          │
                          └──（echoing 态 echo + insight 同时可见）
```

### 2.2 PastPage — 过往

```
PastPage (ConsumerStatefulWidget)
├── PageHeader                    # "每一次说出口的，都留在这里"
└── Expanded
    └── ListView.builder          # 按月分组的 Timeline
        └── MonthSection × N      # 每月一组
            ├── MonthLabel        # "2026年5月"
            └── MomentItem × N
                ├── DateLabel     # "5月3日"
                ├── ContentPreview # 单行截断，TextOverflow.ellipsis
                ├── ConnectionBadge # ✦ 已联结（条件渲染）
                └── GestureDetector
                    └── onTap → AnimatedSize 展开完整内容
```

### 2.3 StarmapPage — 星图

```
StarmapPage (ConsumerStatefulWidget)
└── Stack
    ├── NebulaBackground          # CustomPaint，星云渐变
    ├── InteractiveViewer         # 缩放 + 平移手势
    │   └── CustomPaint           # StarFieldPainter
    │       ├── ConstellationNode × N   # 已成型星座
    │       │   ├── BrightStar (Paint)  # 亮星，独立 AnimationController 控制闪烁
    │       │   ├── ConnectionLine      # 连线，Path + dashEffect + 呼吸 alpha
    │       │   └── NameLabel          # TextPainter，金色光晕 + backdropFilter 等效
    │       ├── FormingStarNode × N     # 正在成型
    │       │   ├── WeakPulseLight      # 弱光，正弦 alpha 脉冲
    │       │   └── HintLabel          # "隐约有什么…"
    │       └── LoneStarNode × N        # 孤星
    │           └── DimLight            # 低 alpha 静态光点
    └── GestureDetector
        └── onTap → hitTest 命中 ConstellationNode → context.push('/starmap/detail')
```

### 2.4 ConstellationDetailPage — 星座详情

```
ConstellationDetailPage (ConsumerWidget)
├── AppBar
│   ├── BackButton                 # Icons.arrow_back
│   └── Title(constellation.name)
└── SingleChildScrollView
    └── Column
        ├── InsightSection         # ① ✦ 我发现
        │   ├── InsightCard        # 金色卡片
        │   └── CollapsibleMoments # ExpansionTile → 展开 MomentItem 列表
        ├── ChatSection            # ② 和那时的自己说说话
        │   ├── PastSelfCard × 3   # 紫色卡片，最多 3 条
        │   │   └── onTap → showModalBottomSheet(ChatDialog)
        │   └── "…还有 N 条" 折叠入口（超出 3 条时）
        └── TopicPromptSection     # ③ ✦ 我想和你聊聊
            └── TopicPromptCard × 3 # 金色卡片
                └── onTap → context.pop + navigateToNow(topic: prompt)
```

### 2.5 ChatDialog — 对话模式

```
ChatDialog (StatefulWidget, showModalBottomSheet)
├── DraggableScrollableSheet
│   └── Column
│       ├── ChatBubbleList        # ListView.builder + reversed
│       │   ├── PastSelfBubble    # AI 第一人称回复（左对齐）
│       │   │   └── SourceLabel   # "参考了 X月X日 前后"
│       │   └── UserBubble        # 用户回复（右对齐）
│       └── ChatInput
│           ├── TextField
│           └── SendButton
```

---

## 3. 数据实体

### 3.1 实体定义

```dart
// ─── 核心实体 ───

class Moment {
  final String id;
  final String content;        // 用户写的原话
  final DateTime createdAt;
  final String traceId;        // 所属 Trace
  final bool connected;        // 是否已联结进星座
}

class Trace {
  final String id;
  final List<Moment> moments;   // 一轮对话的所有话语（含多轮"顺着再想想"）
  final Echo? echo;             // 最后一次回声
  final Insight? insight;       // AI 观察
  final DateTime createdAt;
  final bool stashed;           // 是否已收进星图
}

class Echo {
  final String id;
  final Moment targetMoment;    // 匹配到的历史话语
  final List<Moment> candidates; // 2-3 条候选回声
  final double similarity;
  final DateTime generatedAt;
}

class Insight {
  final String id;
  final String text;             // 第二人称观察文本
  final List<String> relatedMomentIds;
  final DateTime generatedAt;
}

class Constellation {
  final String id;
  final String name;             // 主题标签名
  final ConstellationStatus status;
  final List<Star> stars;
  final int momentCount;         // 包含的话语数
  final DateTime createdAt;
  final DateTime updatedAt;
}

// 以下实体由 GetConstellation 返回，不嵌入 Constellation 实体中

class PastSelfCard {
  final String id;
  final String title;            // 展示标题（如日期范围 "2月-3月"）
  final String openingLine;      // AI 扮演时的开场白
  final List<Moment> contextMoments; // 该时段关联的原话
}

class MomentReference {
  final String date;             // "X月X日"
  final String snippet;          // 原话片段
}

enum ConstellationStatus { formed, forming, lone }

class Star {
  final String id;
  final String traceId;          // 关联的 Trace
  final double x;                // 星图坐标 x（0-1 归一化）
  final double y;                // 星图坐标 y（0-1 归一化）
  final StarVisualState visualState;
  final double rhythm;           // 闪烁节奏参数 0-1，每个星星不同
}

enum StarVisualState { bright, pulsing, dim }

class TopicPrompt {
  final String id;
  final Moment anchorMoment;     // 锚定的原话
  final String question;         // 具体问题
  final String constellationId;
}

class ChatMessage {
  final String id;
  final ChatRole role;
  final String content;
  final List<MomentReference> referenced;  // 标注引用来源
  final DateTime timestamp;
}

enum ChatRole { user, pastSelf }
```

### 3.2 实体关系

```
Moment ──(N:M)── Constellation
  │                    │
  │ 属于               │ 拥有
  ▼                    ▼
Trace                Star
  │                    │
  │ 产生               │ 关联
  ▼                    │
Echo ◄─────────────────┘
  │
  │ 伴随
  ▼
Insight
  │
  │ 关联
  ▼
TopicPrompt

PastSelfCard ──(1:N)── ChatSession ──(1:N)── ChatMessage
  │                                  │
  │ 引用 Moment                      │ 引用 MomentReference
  └──────────────────────────────────┘
```

---

## 4. 状态管理（Riverpod）

### 4.1 Provider 分层

```
┌─────────────────────────────────────────────────────────┐
│  Widget-local State（StatefulWidget.setState）           │
│  - 浮层显隐、AnimatedPositioned 偏移值                   │
│  - TextField 文本、ScrollController                     │
│  - ExpansionTile 展开/折叠                               │
│  - AnimationController（光团、星星闪烁、仪式感动效）      │
├─────────────────────────────────────────────────────────┤
│  Page-level Providers（StateNotifierProvider）           │
│  - nowPageStatusProvider                                │
│    状态: idle | writing | echoing | stashing             │
│    数据: currentTrace (Moment[]), echoCandidates[]      │
│  - memoryDotsProvider (AutoDispose, 随机 3 条)           │
│  - constellationDetailProvider (family, 按 id 查询)      │
├─────────────────────────────────────────────────────────┤
│  App-level Providers（全局常驻）                          │
│  - authProvider（JWT token + user_id + isLoggedIn）      │
│  - momentRepositoryProvider                             │
│  - constellationListProvider                            │
│  - connectedMomentIdsProvider                           │
│  - selectedTabProvider                                  │
└─────────────────────────────────────────────────────────┘
```

### 4.2 状态共享矩阵

```
状态                     生产者                      消费者
─────────────────────────────────────────────────────────────────
momentList               NowPage（提交写字）          PastPage, StarmapPage
constellations            后端 AI 聚类                StarmapPage, DetailPage
currentTrace             NowPage（写字循环）          NowPage（EchoCard, InsightCard, StashButton）
connectedMomentIds       StashTrace 后更新            PastPage（✦ 已联结标记）
selectedTabIndex         AppShell                    StashAnimation（飞行目标 Tab）
chatHistory              ChatDialog                  ChatDialog（自身消息列表）
```

### 4.3 数据流（含 gRPC 调用）

```
用户写字 → NowPageNotifier.submitMoment(content)
              │
              ├─ 1. EgoClient.CreateMoment(content) ─── gRPC → 后端保存 + 匹配回声
              ├─ 2. MomentRepository.saveLocal(moment) ── Hive 本地双写
              ├─ 3. 如果 echo 存在：
              │      EgoClient.GenerateInsight(content, echoMomentId) ─── gRPC → AI 观察
              └─ 4. 更新 currentTrace 状态
                      │
          ┌───────────┼───────────┐
          ▼           ▼           ▼
      展示回声     顺着再想想    收进星图
      (读 state)  (追加 Moment  (EgoClient.StashTrace
                   到 currentTrace)  ─── gRPC → 生成 Star + 重新聚类)
                                        │
                                        ▼
                               Constellation 状态更新
                               (formed/forming/lone 可能变更)
```

---

## 5. 渲染策略（Flutter 实现）

### 5.1 组件与渲染方案

| 组件 | Flutter 方案 | 原因 |
|------|-------------|------|
| StarryBackground | `CustomPaint` + `CustomPainter` | 200+ 小星星独立闪烁，每帧重绘，需控制 repaint 范围 |
| StarField（星图） | `CustomPaint` + `InteractiveViewer` | 缩放/平移 + 星星独立动画 + 连线 Path |
| BreathingLight | `CustomPaint` + `FragmentShader`（或纯 Dart shader 模拟） | 不规则形变 + 光核脉冲需逐像素计算 |
| MemoryDot | `AnimatedPositioned`（漂浮）+ `CustomPaint`（光晕） | 漂浮用隐式动画，光晕用 Canvas |
| EchoCard / InsightCard | `AnimatedSlide` + `AnimatedOpacity` | 标准 UI，利用 Widget 合成 |
| Timeline | `ListView.builder` + `AnimatedSize` | 按月分组，itemBuilder 懒加载 |
| StashAnimation | `Overlay` + `AnimationController` 编排 | 跨组件飞行粒子动画 |
| ChatDialog | `showModalBottomSheet` → `DraggableScrollableSheet` | Flutter 原生半模态 |

### 5.2 包依赖建议

```
用途                 推荐包                    备注
─────────────────────────────────────────────────────────
路由                 go_router                 ShellRoute + StatefulShellRoute
状态管理              flutter_riverpod          NotifierProvider + AsyncNotifierProvider
本地存储              hive + hive_flutter       已在使用，实体类需注册 TypeAdapter
数据类                freezed                   不可变实体，copyWith + json 序列化
gRPC 客户端           grpc + protobuf           Dart 生成的 stub，连接 Go 后端
动画                  flutter_animate           快速声明式动画，减少 AnimationController 样板
```

---

## 6. 动画架构

### 6.1 动画分类与实现

| 类型 | 实现 | 场景 |
|------|------|------|
| 隐式动画 | `AnimatedPositioned`, `AnimatedOpacity`, `AnimatedSlide`, `AnimatedScale` | 浮层进出、光团缩小上移、按钮显隐 |
| 显式动画 | `AnimationController` + `Tween` + `AnimatedBuilder` | 星星闪烁、光团呼吸、仪式感 6 段序列 |
| 交错动画 | 多个 `AnimationController`，`Interval` 错开 | EchoCard → InsightCard 依次浮出（stagger 150ms） |
| 粒子动画 | `CustomPainter` + `AnimationController`，每帧算粒子位置 | 仪式感动效 3-6 段（贝塞尔飞行、拖尾、星屑炸开、Tab 脉冲） |
| 物理动画 | `AnimationController` + `Curves.elasticOut` | 写字区滑出（弹簧）、记忆光点漂浮 |

### 6.2 关键动画参数

**光团呼吸（idle）**
- `AnimationController` 循环播放，`duration: 3000ms`，`Curves.easeInOut`
- 控制：`scale`（1.0 ↔ 1.15）、`opacity`（0.6 ↔ 1.0）、形变 shader 参数

**写字区滑出**
- `AnimationController(duration: 350ms, vsync: this)`
- `CurvedAnimation(parent: controller, curve: Curves.elasticOut)`
- 驱动 `AnimatedPositioned` bottom: `-height → safeArea.bottom`

**仪式感动效编排（6 段序列）**
1. 卡片发光：`Opacity` 叠加金色层，300ms
2. 三圈涟漪：`CustomPaint` 半径从 0 到 200px，循环 3 次，每圈 200ms + gap 100ms
3. 贝塞尔飞行：粒子从卡片中心 → 星图 Tab 中心，贝塞尔曲线，600ms
4. 拖尾：飞行粒子后方生成 12 个渐隐粒子，跟随主粒子路径
5. 星屑炸开：到达目标时 spawn 30 个随机方向粒子，300ms
6. Tab 脉冲：Tab icon scale 1.0→1.3→1.0，500ms `Curves.easeOut`

### 6.3 性能注意事项

- 星星闪烁使用独立的 `AnimationController`（每星一个 Ticker），Tab 不可见时 `TickerMode` 自动暂停
- `InteractiveViewer` 内星星尺寸不随缩放变化：在 `CustomPainter` 中忽略缩放矩阵的 scale 分量，仅坐标缩放（类似地图 POI 固定尺寸）
- 重绘隔离：`RepaintBoundary` 包裹 StarryBackground 和 StarField，避免 UI 浮层变化触发整画布重绘
- 回声匹配为异步操作：提交后显示 Shimmer 占位（`shimmer` 包或自己用 `CustomPaint` 画呼吸光），不阻塞 UI
- `const` Widget 最大化编译期常量优化

---

## 7. 离线与容错策略

未连接服务器时显示"暂时不可用"提示，不提供离线降级。AI 类接口（回声匹配、观察生成、对话模式）超时或失败时静默降级，不报错。

---

## 8. 目录结构

> 详细结构见 `client.md`，以下为与架构对应的摘要。

```
lib/
├── main.dart
├── app.dart                           # MaterialApp.router + ProviderScope
├── core/
│   ├── theme/
│   │   ├── app_theme.dart             # ThemeData（深色主题）
│   │   └── app_colors.dart            # 暖金/冷蓝/柔紫/金色 色板
│   ├── router/
│   │   └── app_router.dart            # GoRouter 配置（StatefulShellRoute + 登录守卫）
│   ├── providers/
│   │   ├── auth_provider.dart         # JWT token + user_id + isLoggedIn
│   │   └── tab_provider.dart          # 当前 selectedTabIndex
│   └── constants.dart                 # 动画持续时长、星星数量等常量
├── data/
│   ├── models/                        # 所有实体类（freezed），与 proto message 对应
│   ├── repositories/
│   │   └── local_store.dart           # Hive 本地持久化（token 缓存、moment 缓存）
│   ├── services/
│   │   ├── ego_client.dart            # gRPC 客户端封装 + JWT metadata 注入
│   │   └── interceptors/
│   │       └── auth_interceptor.dart   # 所有请求注入 Authorization header
│   └── generated/                     # protoc 生成的 Dart 文件（只读）
├── features/
│   ├── login/
│   │   └── login_page.dart            # 登录/自动注册
│   ├── now/
│   │   ├── now_page.dart              # 此刻主页
│   │   ├── providers/                 # nowPageProvider + memoryDotsProvider
│   │   └── widgets/                   # breathing_light, echo_card, insight_card ...
│   ├── past/
│   │   ├── past_page.dart             # 过往时间线
│   │   ├── providers/                 # timelineProvider
│   │   └── widgets/                   # month_section, moment_item
│   └── starmap/
│       ├── starmap_page.dart          # 星图主页
│       ├── providers/                 # constellationProvider + detail family
│       ├── detail/
│       │   ├── detail_page.dart       # 星座详情页
│       │   ├── providers/
│       │   └── widgets/              # insight_section, chat_dialog ...
│       └── painters/                  # star_field_painter, nebula_painter
└── shared/
    ├── widgets/                       # 可复用组件（star_painter, shimmer_card）
    └── extensions/                    # date_format 等扩展

proto/                                # proto 源文件（项目根目录，前后端共享）
└── ego/
    └── api.proto
```

---

## 9. 页面关系总结

```
NowPage ──写下的 Moment 流入──▶ PastPage ──聚类形成──▶ StarmapPage
   │                                │                      │
   │   回声从 Past 的 Moment 匹配    │  ✦ 已联结回显        │  DetailPage 引用 Moment 原话
   │                                │                      │
   └────────────── 数据读取方向 ◀──────────────────────────┘
```

---

## 10. 前后端协议（gRPC）

> 完整 proto 定义见 `proto/ego/api.proto`

### 10.1 gRPC 客户端 + JWT 拦截器

```dart
// data/services/ego_client.dart
// 注意：生成的 gRPC stub 类名也是 EgoClient，需通过 import prefix 区分
import 'generated/api.pbgrpc.dart' as grpc;

class EgoClient {
  final grpc.EgoClient _stub;

  EgoClient(this._stub);

  // 所有请求注入 Authorization metadata
  Future<Response> _call(Function handler) async {
    final token = AuthProvider.instance.token;
    final metadata = {
      'authorization': 'Bearer $token',
    };
    return handler(metadata);
  }
}
```

后端统一从 gRPC metadata 提取 JWT → 解析 user_id → 注入 context。

### 10.2 接口清单与 UI 映射

#### 登录

| UI 行为 | RPC | 说明 |
|---------|-----|------|
| 输入 account/password 进入 | `Ego.Login` | 未注册自动创建，返回 JWT token |

#### 此刻（NowPage）

| UI 行为 | RPC | 说明 |
|---------|-----|------|
| 进入主屏 | `Ego.GetRandomMoments` | 获取 3 条随机历史话语作为记忆光点 |
| 点击"写下此刻" | — | 纯前端动画，光团缩小上移 |
| 提交写字 → 回声出现 | `Ego.CreateMoment` | 保存话语 + 服务端匹配回声（含 2-3 候选） |
| 回声出现后 → 观察出现 | `Ego.GenerateInsight` | 与 CreateMoment 并行调用，AI 生成第二人称观察 |
| 点"顺着再想想" → 继续写 | 回到写字区 → 再调 `Ego.CreateMoment`（同 trace_id） | 新内容保存为新 Moment，追加到当前 Trace |
| 点"✦ 收进星图" | `Ego.StashTrace` | 提交 Trace → 生成 Star + 重评星座状态 |
| 点"嗯，先这样" | — | 纯前端，重置 idle 状态 |
| 点记忆光点 | — | 纯前端，已缓存数据展开信封 |

**CreateMoment 调用时序：**

```
用户点"先到这"
      │
      ├─ Show shimmer loading
      ├─ Ego.CreateMoment(content, trace_id)  ─────┐
      │   → moment + echo                           │ 后端: 保存 + 语义匹配回声
      │   → 渲染 EchoCard                           │
      ├─ Ego.GenerateInsight(content, echo_id) ─────┐
      │   → insight                                 │ 后端: AI 生成观察（可并行）
      │   → 渲染 InsightCard                        │
      └─ Hide shimmer
```

#### 过往（PastPage）

| UI 行为 | RPC | 说明 |
|---------|-----|------|
| 进入/下拉加载更多 | `Ego.ListMoments` | 游标分页，返回扁平列表，前端按月分组 |
| ✦ 已联结标记 | Moment.connected 字段 | 由 `Ego.StashTrace` 后后端更新 |

#### 星图（StarmapPage）

| UI 行为 | RPC | 说明 |
|---------|-----|------|
| 进入星图 | `Ego.ListConstellations` | 返回所有星座（含星星坐标、视觉状态） |
| 点击星座 | `Ego.GetConstellation` | 返回星座详情全量数据（insight + moments + pastSelfCards + topicPrompts） |
| push 到详情页 | 携带 constellation id，详情页 initState 时调用 | `Ego.GetConstellation` |

#### 星座详情页 + 对话模式

| UI 行为 | RPC | 说明 |
|---------|-----|------|
| 打开详情页 | `Ego.GetConstellation(id)` | 获取 4 块数据（① 观察 ② 对话卡片 ③ 原话列表 ④ 话题引子） |
| 点击"和那时的自己说说话"卡片 | `Ego.StartChat(card_id, context_moment_ids)` | 后端构建 past-self 上下文，返回 sessionId + AI 开场白 |
| 发送消息 | `Ego.SendMessage(session_id, content)` | AI 以 past-self 第一人称回复，附带引用标注 |

#### 话题引子

| UI 行为 | RPC | 说明 |
|---------|-----|------|
| 进入详情页 | 通过 `Ego.GetConstellation` 返回 | topicPrompts 随详情页一起返回 |
| 点击话题引子卡片 | — | 前端 pop + navigateToNow(topic: prompt)，将话题带入写字区 |

### 10.3 RPC 调用频率与缓存策略

| RPC | 调用频率 | 缓存策略 |
|-----|---------|---------|
| `Login` | 用户登录/注册 | 不缓存，token 本地持久化 |
| `GetRandomMoments` | 每次进 NowPage / 下拉刷新 | 不缓存，每次随机 |
| `CreateMoment` | 每次提交写字 | 不缓存（写操作），结果存入 Hive |
| `GenerateInsight` | 每次提交写字 | 不缓存，失败时前端隐藏 InsightCard |
| `StashTrace` | 用户手动触发 | 不缓存（写操作），成功后刷新 constellations |
| `ListMoments` | 进入 PastPage + 滚动加载更多 | 分页结果缓存于 Hive + Provider |
| `ListConstellations` | 进入 StarmapPage | Provider 缓存，StashTrace 后失效 |
| `GetConstellation` | 进入详情页 | Provider family 缓存（按 id） |
| `StartChat` / `SendMessage` | 用户手动触发 | 不缓存，聊天记录本地 Hive 保存 |

### 10.4 网络不可用

未连接服务器时显示"暂时不可用"提示，不提供离线降级。
