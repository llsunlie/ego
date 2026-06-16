---
name: ego-now
description: 「此刻」首页 feature — 前端 client page: /now + 后端 writing 领域。用户写作、AI 回声、记忆匹配、收星图核心交互。涉及文件: client/lib/features/now/ + server/internal/writing/。
---

# ego-now

「此刻」全栈 feature context — ego 核心交互。修改此功能时，agent 应同时阅读：

- 前端详细 context ➔ Read `client.md`
- 后端详细 context ➔ Read `server.md`

## 快速文件索引

### 前端 (`client/`)
| 文件 | 说明 |
|------|------|
| `client/lib/features/now/now_page.dart` | 页面主体 + 状态机驱动 |
| `client/lib/features/now/providers/now_page_provider.dart` | NowPageState + NowPageNotifier |
| `client/lib/features/now/widgets/` | 8 个 widget 组件 |
| `client/lib/data/services/ego_client.dart` | `createMoment()` 等 API |

### 后端 (`server/`)
| 文件 | 说明 |
|------|------|
| `server/internal/writing/domain/types.go` | Moment, Echo, Insight, Trace |
| `server/internal/writing/app/create_moment.go` | CreateMoment 用例：创建 Moment + 混合 Echo 召回 |
| `server/internal/writing/app/echo_hybrid.go` | Dense/Sparse Echo 候选 RRF 融合 |
| `server/internal/writing/app/echo_matcher.go` | Echo 匹配（向量相似度） |
| `server/internal/writing/adapter/postgres/moment_sparse_search.go` | pg_trgm 稀疏文本召回 |
| `server/internal/writing/adapter/postgres/echo_candidate_reader.go` | pgvector dense 最近邻召回 |
| `server/internal/writing/adapter/grpc/handler.go` | gRPC Handler |
| `server/internal/writing/module.go` | 模块组装 |

### Proto 契约
| 文件 | 说明 |
|------|------|
| `proto/ego/api.proto` | `CreateMomentReq/Res`, `Echo`, `Insight` 定义 |
