---
name: ego-past
description: 「过往」feature — 前端 client page: /past + 后端 timeline 领域。历史 trace 列表、按月分组、trace 详情（moment→echo→insight 链路）。涉及文件: client/lib/features/past/ + server/internal/timeline/。
---

# ego-past

「过往」全栈 feature context。修改此功能时，agent 应同时阅读：

- 前端详细 context ➔ Read `client.md`
- 后端详细 context ➔ Read `server.md`

## 快速文件索引

### 前端 (`client/`)
| 文件 | 说明 |
|------|------|
| `client/lib/features/past/past_page.dart` | Trace 列表页（分月分组 + 无限滚动） |
| `client/lib/features/past/trace_detail_page.dart` | Trace 详情页 |
| `client/lib/features/past/providers/past_page_provider.dart` | PastPageState（分页状态管理） |
| `client/lib/features/past/widgets/trace_item.dart` | Trace 列表项组件 |
| `client/lib/data/services/ego_client.dart` | `listTraces()` 等 API |

### 后端 (`server/`)
| 文件 | 说明 |
|------|------|
| `server/internal/timeline/app/list_traces.go` | ListTraces 用例（cursor 分页） |
| `server/internal/timeline/app/get_trace_detail.go` | GetTraceDetail 用例 |
| `server/internal/timeline/app/get_random_moments.go` | GetRandomMoments 用例 |
| `server/internal/timeline/module.go` | 模块组装（复用 writing Reader） |
| `server/internal/writing/adapter/postgres/reader.go` | 跨模块 Reader 实现 |

### Proto 契约
| 文件 | 说明 |
|------|------|
| `proto/ego/api.proto` | `ListTracesReq/Res`, `GetTraceDetailReq/Res`, `TraceItem` 定义 |
