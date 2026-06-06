---
name: ego-cli-starmap
description: 「星图」页面 context — 路由 /starmap + /starmap/detail/:constellationId，星座可视化画布、星座详情、与过去自我的 AI 对话。
---

# ego-cli-starmap

「星图」页面 context — 星座可视化、星座详情、与过去自我的对话。使用此 skill 后可直接讨论或修改星图页面。

## 路由

- `/starmap` — AppShell 底部导航第三个 tab（索引 2），星空画布
- `/starmap/detail/:constellationId` — 星座详情 + AI 对话

## 核心文件

### 主页面
| 文件 | 说明 |
|------|------|
| `client/lib/features/starmap/starmap_page.dart` | 星图画布，星星/星座定位渲染，点击检测，引导提示 |
| `client/lib/features/starmap/constellation_detail_page.dart` | 星座详情：星座信息 + moments 列表 + AI 对话 |
| `client/lib/features/starmap/providers/starmap_provider.dart` | StarmapState + StarmapNotifier + pendingTopicPromptProvider |

### 组件 & 数据
| 文件 | 说明 |
|------|------|
| `client/lib/features/starmap/painters/star_field_painter.dart` | CustomPainter：绘制星空背景 + 星座连线 + 星星闪烁 |
| `client/lib/features/starmap/data/star_position.dart` | `StarPositionEngine.placeAll()` — 星座布局算法，hit testing |
| `client/lib/features/starmap/widgets/chat_sheet.dart` | AI 对话底部弹出层 |
| `client/lib/features/starmap/widgets/insight_section.dart` | 洞察展示区域 |
| `client/lib/features/starmap/widgets/past_self_card.dart` | 过去自我的卡片 |
| `client/lib/features/starmap/widgets/topic_prompt_card.dart` | 话题提示卡片，点击可触发 now 页面开始写作 |

### 依赖
| 文件 | 说明 |
|------|------|
| `client/lib/data/services/ego_client.dart` | gRPC API: listConstellations, getConstellation, startChat, sendMessage |
| `client/lib/core/providers/tab_provider.dart` | Tab 切换监听 |
| `client/lib/core/providers/auth_provider.dart` | Auth token |
| `client/lib/data/repositories/local_store.dart` | 本地存储 (tap guide shown 标记) |
| `client/lib/core/theme/colors.dart` | AppColors 色板 |

## 状态管理

**starmapProvider** (`StateNotifierProvider<StarmapNotifier, StarmapState>`)

StarmapState 字段：
- `constellations`: List\<pb.Constellation\> — 星座列表
- `totalStarCount`: int — 总星星数
- `isLoading`: bool
- `error`: String?

**pendingTopicPromptProvider** (`StateProvider<String?>`) — 跨页面通信：从 starmap 触发 now 页面自动写作

## 数据流

### 星图画布
1. 切换到 starmap tab → `loadConstellations()`
2. `listConstellations()` → 返回 constellations + totalStarCount
3. `StarPositionEngine.placeAll(constellations)` → 计算每个星座在画布上的位置
4. `StarFieldPainter` 自定义绘制：星星点 + 星座连线 + 闪烁动画 (twinkleCtrl)
5. 点击检测 (`_tryTap`)：计算屏幕坐标 → 世界坐标转换，检测命中星座 → push detail 路由

### 星座详情
1. `getConstellation(constellationId)` → 返回 constellation + moments + stars
2. 展示星座信息 + 关联的 moments 列表
3. AI 对话：`startChat(starId)` → `sendMessage(chatSessionId, content)` (chat_sheet.dart)

### 与 Now 页面联动
- topic_prompt_card 点击 → 设置 `pendingTopicPromptProvider` → now 页面监听此 provider → `startWriting()`

## 页面结构

### StarmapPage
```
Scaffold
├── Header ("已有 N 颗星")
└── Expanded
    └── LayoutBuilder
        └── GestureDetector (onTapUp → _tryTap)
            └── Stack
                ├── CustomPaint (StarFieldPainter — 星空 + 星座)
                └── TapGuide overlay ("点击星座或星星查看详情")
```

### ConstellationDetailPage
```
Scaffold
├── AppBar (星座名 + 返回)
└── body
    └── SingleChildScrollView
        ├── 星座信息
        ├── InsightSection
        ├── Moments 列表
        └── ChatSheet (底部弹出 AI 对话)
```

## 关键动画

- `_twinkleCtrl` (AnimationController, 5s 循环) — 驱动星星闪烁
- `_showTapGuide` + Timer (3s 自动消失) — 首次进入星图的点击引导
- 引导状态通过 LocalStore 持久化（只显示一次）

## StarPositionEngine 核心逻辑

- `worldWidth / worldHeight`: 虚拟世界坐标
- `placeAll()`: 使用固定 seed 的随机算法分配星座位置，确保每次加载位置一致
- hit testing: `hitRadius * scale` 范围内检测星星或标签区域

## 注意事项

- starmap state 通过 `tabProvider` 监听 tab 切换触发加载（同 past 模式）
- `pendingTopicPromptProvider` 是跨 page 通信机制，定义在 starmap_provider 但被 now_page 消费
- 星座详情页的 AI 聊天通过 chat_sheet (底部弹出层) 实现
