---
name: ego-cli-now
description: 「此刻」首页 context — 路由 /now，用户写想法、接收 AI 回声、匹配过去 moment、收进星图的核心交互页面。
---

# ego-cli-now

「此刻」首页 context — 用户写想法、接收回声、收进星图的主界面。使用此 skill 后可直接讨论或修改首页。

## 路由

`/now` — AppShell 底部导航的第一个 tab（索引 0），默认首页。

## 核心文件

### 主页面
| 文件 | 说明 |
|------|------|
| `client/lib/features/now/now_page.dart` | 页面主体，组合所有 widget，管理 Overlay (StashAnimation) |
| `client/lib/features/now/providers/now_page_provider.dart` | NowPageState + NowPageNotifier，核心状态机 |

### Widget 组件 (8 个)
| 文件 | 说明 |
|------|------|
| `client/lib/features/now/widgets/starry_background.dart` | 星空背景粒子 |
| `client/lib/features/now/widgets/breathing_light.dart` | 中央呼吸光晕，根据 status 切换动画 |
| `client/lib/features/now/widgets/orbiting_satellites.dart` | 环绕光球，dimmed 参数控制 |
| `client/lib/features/now/widgets/memory_dot.dart` | 记忆光点 (MemoryDotGroup)，随机加载过去 moments |
| `client/lib/features/now/widgets/guide_section.dart` | idle 状态引导文字 + 底部 WriteButton |
| `client/lib/features/now/widgets/writing_input.dart` | 写作输入框 (writing/echoing 状态) |
| `client/lib/features/now/widgets/echo_card.dart` | 回声卡片 (EchoSection)，显示 echo + insight + 匹配的过去 moments + 操作按钮 |
| `client/lib/features/now/widgets/insight_card.dart` | AI 洞察卡片 |
| `client/lib/features/now/widgets/stash_animation.dart` | 收进星图的 Overlay 动画 (全屏) |

### 依赖的核心服务
| 文件 | 说明 |
|------|------|
| `client/lib/data/services/ego_client.dart` | gRPC API: createMoment, generateInsight, getMoments, stashTrace, getRandomMoments |
| `client/lib/features/starmap/providers/starmap_provider.dart` | pendingTopicPromptProvider（星图话题联动） |
| `client/lib/core/theme/colors.dart` | AppColors 色板 |

## 状态机 (NowPageStatus)

```
idle → writing → echoing → stashing → idle
  ↑        ↑                    |
  └── reopenWhisper()           |
  └────────────────── dismissEcho() / completeStash()
```

- **idle**: 显示星空 + 呼吸光 + 记忆光点 + 引导文字 + 底部 Write 按钮
- **writing**: 显示写作输入框（WritingInput），用户输入想法
- **echoing**: 提交后显示回声（EchoSection），含 AI 回声 + insight + 匹配的过去 moments + "收进星图"/"算了不要了" 按钮
- **stashing**: Overlay 播放入星图动画 → 调用 stashTrace API → 回到 idle

## 状态管理

**nowPageProvider** (`StateNotifierProvider<NowPageNotifier, NowPageState>`)

NowPageState 字段：
- `status`: NowPageStatus 枚举
- `currentTraceId`, `currentMomentId`: 当前对话追踪
- `echo`: pb.Echo (AI 回声)
- `insight`: pb.Insight (AI 洞察)
- `isLoading`, `error`
- `isReopen`: 是否重新打开写作
- `matchedMoments`: 匹配的历史 moments

**memoryDotsProvider** (`FutureProvider.autoDispose`) — 随机加载 3 个 moments 作为记忆光点

**pendingTopicPromptProvider** (from starmap_provider) — 监听星图话题提示，触发 startWriting

## API 调用序列

1. **submitMoment**: `createMoment(content, traceId)` → 返回 moment + echo
2. **并行**: `generateInsight(momentId, echoId)` + `getMoments(matchedIds)` (如有匹配)
3. **stash**: `stashTrace(traceId)` → 完成收星图
4. **memoryDots**: `getRandomMoments(count=3)` (idle 时展示浮动记忆点)

## 布局结构

```
Scaffold
├── Stack (Expanded)
│   ├── StarryBackground (底层)
│   ├── BreathingLight (状态驱动)
│   ├── OrbitingSatellites (dimmed when !idle)
│   ├── MemoryDotGroup (dimmed when !idle)
│   ├── GuideText (idle only, positioned below light)
│   ├── WritingInput (writing/echoing)
│   └── EchoSection (echoing)
└── WriteButton (idle only, bottom)
```

## 与星图的联动

- starmap_provider 的 `pendingTopicPromptProvider` 可触发 now 页面自动进入 writing 状态
- stash 操作将 trace 保存到星图系统
