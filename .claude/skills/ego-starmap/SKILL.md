---
name: ego-starmap
description: 「星图」feature — 前端 client page: /starmap + 后端 starmap + conversation 领域。星座可视化、AI 对话、Trace Stash。涉及文件: client/lib/features/starmap/ + server/internal/starmap/ + server/internal/conversation/。
---

# ego-starmap

「星图」全栈 feature context。修改此功能时，agent 应同时阅读：

- 前端详细 context ➔ Read `client.md`
- 后端详细 context ➔ Read `server.md`（含 starmap + conversation 两个领域）

## 快速文件索引

### 前端 (`client/`)
| 文件 | 说明 |
|------|------|
| `client/lib/features/starmap/starmap_page.dart` | 星空画布 + 星座渲染 |
| `client/lib/features/starmap/constellation_detail_page.dart` | 星座详情 + AI 对话 |
| `client/lib/features/starmap/providers/starmap_provider.dart` | StarmapState + pendingTopicPrompt |
| `client/lib/features/starmap/painters/star_field_painter.dart` | CustomPainter 星空绘制 |
| `client/lib/features/starmap/widgets/chat_sheet.dart` | AI 对话 UI |

### 后端 (`server/`)
| 文件 | 说明 |
|------|------|
| `server/internal/starmap/app/stash_trace.go` | StashTrace 用例（核心：AI 星座匹配 + 合并） |
| `server/internal/starmap/app/list_constellations.go` | 星座列表 |
| `server/internal/starmap/app/get_constellation.go` | 星座详情 |
| `server/internal/conversation/app/start_chat.go` | 开始 AI 对话 |
| `server/internal/conversation/app/send_message.go` | 发送消息 + AI 回复 |

### Proto 契约
| 文件 | 说明 |
|------|------|
| `proto/ego/api.proto` | `Constellation`, `Star`, `StashTrace`, `StartChat`, `SendMessage` 定义 |
