---
name: ego-cli-past
description: 「过往」页面 context — 路由 /past + /past/detail/:traceId，按月份浏览历史 trace 列表，查看 moment→echo→insight 详情链路。
---

# ego-cli-past

「过往」页面 context — 浏览历史写过的所有 trace，查看每条 trace 的详情（moment → echo → insight 链路）。使用此 skill 后可直接讨论或修改过往页面。

## 路由

- `/past` — AppShell 底部导航第二个 tab（索引 1），trace 列表
- `/past/detail/:traceId` — trace 详情子路由

## 核心文件

### 主页面
| 文件 | 说明 |
|------|------|
| `client/lib/features/past/past_page.dart` | Trace 列表页，按月分组，无限滚动分页 |
| `client/lib/features/past/trace_detail_page.dart` | Trace 详情页，展示完整 moment → echo → insight 链路 |
| `client/lib/features/past/providers/past_page_provider.dart` | PastPageState + PastPageNotifier，分页状态管理 |
| `client/lib/features/past/widgets/trace_item.dart` | Trace 列表项组件 |

### 依赖
| 文件 | 说明 |
|------|------|
| `client/lib/data/services/ego_client.dart` | gRPC API: listTraces, getTraceDetail, getMoments |
| `client/lib/core/providers/tab_provider.dart` | Tab 切换监听，切换到此 tab 时触发首次加载 |
| `client/lib/core/providers/auth_provider.dart` | 提供 auth token 用于 API 调用 |
| `client/lib/data/services/interceptors/auth_interceptor.dart` | authCallOptions 辅助函数 |

## 状态管理

**pastPageProvider** (`StateNotifierProvider<PastPageNotifier, PastPageState>`)

PastPageState 字段：
- `traces`: List\<pb.Trace\> — 当前已加载的 trace 列表
- `nextCursor`: String? — 分页游标
- `hasMore`: bool — 是否有更多数据
- `isLoading`: bool — 首次加载中
- `isLoadingMore`: bool — 加载下一页中
- `error`: String?

计算属性：
- `monthGroups`: Map\<String, List\<pb.Trace\>\> — 按 `{year}年{month}月` 分组
- `sortedMonthKeys`: List\<String\> — 倒序排列的月份键

## 数据流

### 列表页
1. 切换到 past tab → `tabProvider` 变化 → `loadFirstPage()`
2. `listTraces(cursor: '', pageSize: 20)` → 返回 traces + nextCursor + hasMore
3. 滚动到底部 (距底部 200px) → `loadNextPage()` → 追加数据
4. 下拉刷新 → `refresh()` → 重新加载首页

### 详情页
1. 点击 trace → `context.push('/past/detail/$traceId')`
2. `getTraceDetail(traceId)` → 返回 items (List\<TraceItem\>，每个含 moment + echos + insight)
3. 收集所有 echo 的 matchedMomentIds → `getMoments(ids)` 批量获取匹配的历史 moment
4. 展示 timeline: moment 内容 → echo 回声卡片（含候选匹配 moments 切换）→ insight 洞察

## 页面结构

### PastPage
```
Scaffold
├── _PageHeader ("每一次说出口的，都留在这里")
└── ListView.builder (分月分组)
    ├── _MonthLabel ("2024年5月")
    ├── TraceItem × N
    ├── _MonthLabel ("2024年4月")
    └── ... + 底部 loading indicator
```

### TraceDetailPage
```
Scaffold
├── AppBar ("过往详情" + 返回按钮)
└── ListView
    └── TraceItem 卡片 × N
        ├── 时间戳
        ├── Moment 内容 (斜体)
        ├── Echo 回声卡片 (_EchoCard)
        │   └── _CandidateToggle (多个候选匹配 moments 展开/收起)
        └── Insight 洞察卡片 (金色边框)
```

## TraceDetail 关键组件

- `_EchoCard`: 展示 echo 匹配的第一条 moment + "你之前也说过类似的" 标签 + 展开更多候选按钮
- `_CandidateToggle`: 展开/收起 其余匹配 moments，显示相似度百分比
- Insight 区域：渐变背景 + "✦ 我发现" + insight text

## 注意事项

- past_page 依赖 `tabProvider` 监听 tab 切换来触发加载和重置 `_hasLoaded`
- 切换到其他 tab 时 `_hasLoaded` 重置为 false，确保下次切回时重新加载
- 详情页直接使用 `client.stub` 调用 gRPC（非封装方法），手动传 auth options
