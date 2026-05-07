# ego 分层 Harness 项目结构与工作规范

本文说明 ego 在“前后端分离 + 后端 DDD 模块多人协作”场景下，如何组织 Harness 文件和日常工作流。

核心原则：

> 全局定契约，局部管执行。

也就是说：

- 根目录负责跨端、跨模块的协作规则。
- 前端维护自己的 Harness。
- 后端维护自己的端级 Harness。
- 后端每个 DDD 模块维护自己的局部 Harness。
- `proto/` 作为前后端共享契约，有独立 Harness。

这样可以减少 Git 冲突、降低 agent 上下文噪音，也能保护模块 owner 的工作边界。

---

## 1. 推荐项目结构

```text
ego/
├── AGENTS.md
├── DECISIONS.md
├── Makefile
├── .harness/
│   ├── repo-map.md
│   ├── integration-feature-index.json
│   ├── integration-progress.md
│   └── clean-state-checklist.md
│
├── proto/
│   ├── ego/
│   │   └── api.proto
│   ├── CHANGELOG.md
│   └── .harness/
│       ├── contract-rules.md
│       ├── feature_list.json
│       └── progress.md
│
├── client/
│   ├── AGENTS.md
│   ├── .harness/
│   │   ├── init.sh
│   │   ├── feature_list.json
│   │   ├── progress.md
│   │   └── clean-state-checklist.md
│   └── ...
│
└── server/
    ├── AGENTS.md
    ├── .harness/
    │   ├── backend-feature-index.json
    │   ├── integration-progress.md
    │   └── clean-state-checklist.md
    ├── cmd/
    ├── internal/
    │   ├── platform/
    │   │   ├── AGENTS.md
    │   │   ├── ARCHITECTURE.md
    │   │   ├── CONTRACT.md
    │   │   ├── .harness/
    │   │   │   ├── init.sh
    │   │   │   ├── feature_list.json
    │   │   │   ├── progress.md
    │   │   │   ├── decisions.md
    │   │   │   └── clean-state-checklist.md
    │   │   └── ...
    │   │
    │   ├── identity/
    │   │   ├── AGENTS.md
    │   │   ├── ARCHITECTURE.md
    │   │   ├── CONTRACT.md
    │   │   ├── .harness/
    │   │   │   ├── feature_list.json
    │   │   │   ├── progress.md
    │   │   │   └── clean-state-checklist.md
    │   │   └── ...
    │   │
    │   ├── writing/
    │   │   ├── AGENTS.md
    │   │   ├── ARCHITECTURE.md
    │   │   ├── CONTRACT.md
    │   │   ├── .harness/
    │   │   │   ├── feature_list.json
    │   │   │   ├── progress.md
    │   │   │   └── clean-state-checklist.md
    │   │   └── ...
    │   │
    │   ├── timeline/
    │   │   ├── AGENTS.md
    │   │   ├── ARCHITECTURE.md
    │   │   ├── CONTRACT.md
    │   │   ├── .harness/
    │   │   └── ...
    │   │
    │   ├── starmap/
    │   │   ├── AGENTS.md
    │   │   ├── ARCHITECTURE.md
    │   │   ├── CONTRACT.md
    │   │   ├── .harness/
    │   │   └── ...
    │   │
    │   └── conversation/
    │       ├── AGENTS.md
    │       ├── ARCHITECTURE.md
    │       ├── CONTRACT.md
    │       ├── .harness/
    │       └── ...
    └── ...
```

说明：

- 当前仓库可以渐进迁移到这个结构，不要求一次性补齐所有文件。
- 最先落地的应是 `AGENTS.md`、模块 `ARCHITECTURE.md`、模块 `.harness/feature_list.json` 和 `.harness/progress.md`。
- 后端模块目录应和 DDD 限界上下文保持一致。

---

## 2. 各层 Harness 职责

### 2.1 根目录 Harness

根目录是全局协作层。

它负责：

- 仓库级工作规则。
- 前后端边界。
- 后端模块 ownership。
- 跨端契约流程。
- 全局架构红线。
- 集成状态。
- 发布级状态。

它不负责：

- 记录某个后端模块的开发细节。
- 记录某个前端页面的日常进度。
- 记录某个人的本地调试过程。
- 替代模块自己的 Harness。

推荐文件：

```text
ego/
├── AGENTS.md
├── DECISIONS.md
├── Makefile
└── .harness/
    ├── repo-map.md
    ├── integration-feature-index.json
    ├── integration-progress.md
    └── clean-state-checklist.md
```

#### AGENTS.md

全局宪法。

应包含：

- 项目整体说明。
- 全局代码与协作规则。
- 前后端分离规则。
- 后端模块 ownership 规则。
- proto 变更流程。
- 标准验证入口。
- 各端和各模块 Harness 的入口索引。

不应包含：

- 每个模块的详细任务列表。
- 每个开发者的进度。
- 大段模块实现细节。

#### DECISIONS.md

全局决策日志。

记录影响多个端或多个模块的决策，例如：

- 采用 Go 模块化单体。
- 后端按 DDD bounded context 拆分。
- proto 是前后端唯一 API 契约。
- 后端内部不使用 proto 作为领域模型。
- 模块 owner 边界策略。

#### .harness/repo-map.md

仓库地图。

建议记录：

```text
client/                    前端
proto/                     前后端 API 契约
server/internal/platform/  后端基础设施能力
server/internal/writing/   此刻、Moment、Trace、Echo
server/internal/starmap/   Star、Constellation、星图
server/internal/conversation/ 过去自己对话
```

同时记录每个目录的 owner、主要入口文档和禁止事项。

#### .harness/integration-feature-index.json

跨端、跨模块功能索引。

它只记录集成级功能，不记录模块内部细任务。

例如：

```json
{
  "id": "flow.now.write_echo_insight",
  "title": "此刻页：写字后返回回声和观察",
  "status": "in_progress",
  "owners": ["client", "server/internal/writing", "proto"],
  "contract": "proto/ego/api.proto",
  "module_features": [
    "client/.harness/feature_list.json#now.write",
    "server/internal/writing/.harness/feature_list.json#create_moment"
  ],
  "evidence": []
}
```

#### .harness/integration-progress.md

只记录联调和发布级状态。

例如：

- 当前 proto 是否已同步。
- 前端是否已接入后端接口。
- 哪些端到端流程可跑通。
- 哪些模块仍在等待对方。

不要在这里记录 writing 模块今天实现了哪个 repository。

---

### 2.2 proto Harness

`proto/` 是前后端共享契约层。

它负责：

- gRPC API 定义。
- 前后端数据结构契约。
- proto tag 兼容性。
- API 变更记录。
- 生成代码流程说明。

推荐文件：

```text
proto/
├── ego/
│   └── api.proto
├── CHANGELOG.md
└── .harness/
    ├── contract-rules.md
    ├── feature_list.json
    └── progress.md
```

#### proto/ego/api.proto

前后端接口的唯一事实来源。

规则：

- 前端理解后端响应结构时，必须以 proto 为准。
- 后端实现 RPC 时，必须以 proto 为准。
- 不允许前端根据后端内部 Go struct 猜数据结构。
- 不允许后端随意返回 proto 未定义字段。

#### proto/.harness/contract-rules.md

记录 proto 修改规则：

- 字段 tag 只能追加，不能复用。
- 删除字段要先废弃，再清理。
- 修改 RPC 入参或出参必须同步更新前后端任务。
- breaking change 必须写明迁移步骤。
- proto 变更需要更新 `CHANGELOG.md`。

#### proto/.harness/feature_list.json

只记录接口契约任务。

例如：

- 新增 `CreateMoment` RPC。
- 新增 `StashTrace` RPC。
- 给 `CreateMomentReq` 增加 `topic` 字段。

不记录后端内部实现任务。

---

### 2.3 client Harness

`client/` 是前端开发沙盒。

它负责：

- 前端页面和交互任务。
- 前端状态管理。
- 前端动画和 UI 验证。
- 根据 proto 接入后端接口。
- 前端自己的测试和构建。

推荐文件：

```text
client/
├── AGENTS.md
└── .harness/
    ├── init.sh
    ├── feature_list.json
    ├── progress.md
    └── clean-state-checklist.md
```

前端 `AGENTS.md` 应强调：

- 数据结构以 `../proto/ego/api.proto` 为准。
- 不根据后端内部实现猜字段。
- 修改接口需求时，先走 proto 变更流程。
- 前端进度写在 `client/.harness/progress.md`。
- 前端任务写在 `client/.harness/feature_list.json`。

---

### 2.4 server Harness

`server/` 是后端端级规则层。

它负责：

- Go 后端通用规则。
- DDD 分层规则。
- 后端模块 ownership。
- 后端测试入口。
- 后端集成状态。

推荐文件：

```text
server/
├── AGENTS.md
└── .harness/
    ├── backend-feature-index.json
    ├── integration-progress.md
    └── clean-state-checklist.md
```

后端 `AGENTS.md` 应包含：

- `domain` 不允许依赖 proto、pgx、sqlc、grpc。
- `app` 负责编排用例。
- `adapter` 负责 gRPC、Postgres 等技术适配。
- `platform` 提供基础设施能力。
- 模块之间通过 interface、port、read model 或 domain event 协作。
- 后端内部不使用 proto 作为领域模型。
- 非 owner 不修改其他模块实现。

---

### 2.5 后端模块 Harness

后端模块是最重要的局部执行沙盒。

每个模块由一个 owner 或一个小组负责，其他人默认不修改该模块实现。

推荐结构：

```text
server/internal/{module}/
├── AGENTS.md
├── ARCHITECTURE.md
├── CONTRACT.md
├── .harness/
│   ├── init.sh
│   ├── feature_list.json
│   ├── progress.md
│   ├── decisions.md
│   └── clean-state-checklist.md
└── ...
```

#### 模块 AGENTS.md

模块工作入口。

应包含：

- 本模块职责。
- 本模块禁止事项。
- 本模块开发命令。
- 本模块测试命令。
- 本模块 Harness 文件入口。

#### ARCHITECTURE.md

模块内部架构说明。

应包含：

- 模块属于哪个 bounded context。
- 模块拥有的实体、聚合和值对象。
- 模块内部包结构。
- 主要用例。
- 依赖哪些外部 port。
- 哪些表由本模块写入。

#### CONTRACT.md

模块对外契约。

这是其他模块最应该看的文件。

应包含：

- 对外暴露的 interface。
- 可被其他模块读取的 query/read model。
- 领域事件。
- 本模块写入的数据归属。
- 其他模块不能做什么。

例如 `platform/CONTRACT.md` 可以说明：

```text
platform/ai 提供：
- EmbeddingProvider
- MomentInsightGenerator
- PastSelfResponder

业务模块只能依赖接口，不允许直接依赖具体 AI SDK client。
```

例如 `writing/CONTRACT.md` 可以说明：

```text
writing 提供：
- TraceReader
- MomentReader
- CreateMomentUseCase

writing 拥有 moments/traces 的写入权。
其他模块只能只读 Moment，不允许直接创建或更新 Moment。
```

#### .harness/feature_list.json

模块自己的任务清单。

只记录当前模块内部任务。

不要记录其他模块任务。

#### .harness/progress.md

模块自己的进度。

由于只有模块 owner 维护这个文件，它能避免多人同时改根级 progress 造成冲突。

#### .harness/decisions.md

模块级决策日志。

只记录影响本模块内部实现的决策。

影响多个模块的决策应上移到根 `DECISIONS.md` 或 `server/DECISIONS.md`。

---

## 3. 后端模块 ownership 规则

### 3.1 基本规则

后端模块以目录为 ownership 单位。

例如：

```text
server/internal/platform/      owner: platform developer
server/internal/writing/       owner: writing developer
server/internal/starmap/       owner: starmap developer
server/internal/conversation/  owner: conversation developer
```

规则：

- owner 可以修改自己模块内的实现。
- 非 owner 默认不修改其他模块实现。
- 需要跨模块能力时，先读对方 `CONTRACT.md`。
- 如果契约不足，提出契约变更，而不是直接改对方实现。
- 跨模块变更需要在相关模块 Harness 中记录。

### 3.2 允许读取的内容

开发当前模块时，可以读取其他模块：

```text
AGENTS.md
ARCHITECTURE.md
CONTRACT.md
公开 interface 文件
公开 read model 定义
```

不建议读取：

```text
其他模块内部 repository 实现
其他模块内部 service 私有逻辑
其他模块 adapter/postgres 细节
其他模块未公开 helper
```

除非当前任务明确是跨模块重构，并且已经得到 owner 同意。

### 3.3 禁止行为

开发当前模块时禁止：

- 为了让测试通过直接修改其他 owner 模块实现。
- 绕过对方公开 interface 访问内部结构。
- 直接写其他模块拥有的数据表。
- 在当前模块复制粘贴其他模块内部逻辑。
- 用 proto 类型作为后端内部领域模型。

---

## 4. 后端模块间协作规则

### 4.1 proto 的边界

`proto/` 只作为 client-server API 契约。

后端内部模块之间不通过 proto 模型协作。

后端内部应使用：

- Go interface
- domain port
- application interface
- read model
- domain event

### 4.2 模块间调用方式

推荐优先级：

1. 当前模块定义自己需要的 port。
2. 其他模块或 platform 提供 adapter 实现。
3. 只读场景使用明确 read model。
4. 跨边界异步协作用 domain event。
5. 避免直接 import 其他模块内部实现。

### 4.3 示例：writing 使用 platform AI

`writing` 需要 embedding，不应该直接依赖 OpenAI/Gemini client。

推荐方式：

```text
writing/domain/ports.go
  定义 EmbeddingProvider interface

platform/ai
  实现 EmbeddingProvider

bootstrap/wire.go
  把 platform 实现注入 writing usecase
```

这样：

- writing 表达自己的业务需求。
- platform 负责技术实现。
- 双方通过 interface 协作。
- writing agent 不需要阅读 platform 内部代码。

### 4.4 示例：starmap 读取 writing 的 Moment

`starmap` 需要读取 Trace 下的 Moment 来聚类。

推荐方式：

```text
writing/CONTRACT.md
  声明 TraceReader / MomentReader

starmap/app/StashTraceUseCase
  依赖 TraceReader interface

writing/adapter 或 app
  提供只读实现
```

不推荐：

```text
starmap 直接写 moments 表
starmap 修改 writing repository
starmap 复制 writing SQL
```

---

## 5. Agent 工作流

### 5.1 开发前端

推荐工作目录：

```text
client/
```

读取顺序：

```text
../AGENTS.md
./AGENTS.md
./.harness/progress.md
./.harness/feature_list.json
../proto/ego/api.proto
```

规则：

- 前端任务状态只更新 `client/.harness`。
- API 字段以 `proto/ego/api.proto` 为准。
- 如果需要改接口，切换到 proto 变更流程。

### 5.2 开发后端模块

以 `writing` 为例。

推荐工作目录：

```text
server/internal/writing/
```

读取顺序：

```text
../../../AGENTS.md
../../AGENTS.md
./AGENTS.md
./ARCHITECTURE.md
./CONTRACT.md
./.harness/progress.md
./.harness/feature_list.json
```

规则：

- 只修改 `server/internal/writing/` 及明确允许的共享入口。
- 如果需要 platform 能力，只读 `platform/CONTRACT.md` 和公开 interface。
- 如果需要 proto 变更，先走 proto Harness。
- 完成后更新 writing 自己的 progress 和 feature_list。

### 5.3 开发 platform

推荐工作目录：

```text
server/internal/platform/
```

读取顺序：

```text
../../../AGENTS.md
../../AGENTS.md
./AGENTS.md
./ARCHITECTURE.md
./CONTRACT.md
./.harness/progress.md
./.harness/feature_list.json
```

规则：

- platform 负责基础设施能力，不写业务流程。
- platform 对外能力必须写进 `CONTRACT.md`。
- 如果新增 AI、DB、事件、配置等基础设施能力，要说明业务模块如何通过 interface 使用。

### 5.4 修改 proto

推荐工作目录：

```text
proto/
```

读取顺序：

```text
../AGENTS.md
./.harness/contract-rules.md
./.harness/progress.md
./.harness/feature_list.json
./ego/api.proto
```

规则：

- 先改 proto。
- 更新 `proto/CHANGELOG.md`。
- 更新 `proto/.harness/feature_list.json`。
- 通知 client 和 server 对应模块更新各自任务。
- breaking change 必须写迁移说明。

---

## 6. 进度文件维护规则

### 6.1 不同层级写不同进度

| 层级 | 文件 | 写什么 |
| --- | --- | --- |
| 根目录 | `.harness/integration-progress.md` | 跨端、跨模块集成状态 |
| proto | `proto/.harness/progress.md` | API 契约变更状态 |
| client | `client/.harness/progress.md` | 前端开发状态 |
| server | `server/.harness/integration-progress.md` | 后端模块集成状态 |
| backend module | `server/internal/{module}/.harness/progress.md` | 模块内部开发状态 |

### 6.2 避免 Git 冲突

日常开发时：

- 前端开发者只改 `client/.harness/progress.md`。
- writing 开发者只改 `server/internal/writing/.harness/progress.md`。
- platform 开发者只改 `server/internal/platform/.harness/progress.md`。

只有做联调、发布、跨模块任务时，才更新根级或 server 级 progress。

### 6.3 feature_list 维护规则

每个 `feature_list.json` 只维护当前层级的任务。

例如：

- `client/.harness/feature_list.json` 放前端页面和交互任务。
- `server/internal/writing/.harness/feature_list.json` 放 Writing 用例任务。
- `proto/.harness/feature_list.json` 放接口契约任务。
- 根 `.harness/integration-feature-index.json` 只索引跨端流程。

不要把所有任务集中到根目录。

---

## 7. 验证规范

### 7.1 局部验证

模块开发者优先跑模块级验证。

例如：

```text
server/internal/writing/.harness/init.sh
go test ./internal/writing/...
```

前端开发者优先跑：

```text
client/.harness/init.sh
```

或前端自己的 test/build 命令。

### 7.2 端级验证

当模块任务完成后，应跑端级验证。

后端：

```text
cd server
go test ./...
go build ./...
```

前端：

```text
cd client
make check
```

实际命令以各自 `AGENTS.md` 和 `Makefile` 为准。

### 7.3 集成验证

跨端流程完成后，更新根级集成状态。

例如：

- Login 前后端联调通过。
- CreateMoment 前端可以正常收到 Echo。
- StashTrace 后星图刷新成功。

集成验证证据写入：

```text
ego/.harness/integration-progress.md
```

必要时同步更新：

```text
ego/.harness/integration-feature-index.json
```

---

## 8. 干净收尾规范

每次开发会话结束前，应至少完成：

```text
- 当前层级 feature_list.json 状态已更新
- 当前层级 progress.md 已更新
- 已运行可运行的验证命令
- 未运行验证时已说明原因
- 没有临时调试代码
- 没有跨 owner 的未授权修改
- 如果改了契约，已更新 proto 相关文档
- 如果影响跨模块集成，已更新集成 progress
```

模块 owner 的日常任务只需要清理本模块 Harness。

跨模块任务需要同时清理相关模块 Harness 和集成 Harness。

---

## 9. 什么时候需要上移到更高层 Harness

### 9.1 保持在模块层

以下内容留在模块 Harness：

- 模块内部重构。
- 模块内部测试状态。
- 模块内部实现决策。
- 模块内部任务拆分。

### 9.2 上移到 server 层

以下内容上移到 `server/.harness` 或 `server/AGENTS.md`：

- 后端 DDD 规则变化。
- 多个后端模块共同遵守的规范。
- 后端端级验证命令变化。
- 后端模块集成状态。

### 9.3 上移到根层

以下内容上移到根 Harness：

- 前后端契约流程变化。
- 跨端功能联调状态。
- 发布级状态。
- 影响整个仓库的架构决策。
- owner 划分变化。

### 9.4 上移到 proto 层

以下内容进入 `proto/` Harness：

- RPC 新增、删除、改名。
- message 字段变化。
- 字段兼容性规则。
- breaking change 迁移。

---

## 10. 推荐落地顺序

不需要一次性创建所有文件。

推荐按以下顺序落地：

1. 根 `AGENTS.md`
2. 根 `.harness/repo-map.md`
3. `server/AGENTS.md`
4. `client/AGENTS.md`
5. `proto/.harness/contract-rules.md`
6. 后端各模块 `ARCHITECTURE.md`
7. 后端各模块 `CONTRACT.md`
8. 后端各模块 `.harness/feature_list.json`
9. 后端各模块 `.harness/progress.md`
10. client/proto/server/root 的 progress 和 clean-state checklist

优先保证：

- agent 知道自己在哪个边界内工作。
- owner 不互相踩文件。
- proto 变更有固定流程。
- 模块间调用只看契约，不看内部实现。
- 完成状态有证据。

---

## 11. 总结

ego 的 Harness 不应该是单点式的。

更适合采用：

```text
Root Harness
  管全局规则、跨端集成、发布状态

Proto Harness
  管前后端 API 契约

Client Harness
  管前端执行

Server Harness
  管后端端级规则

Backend Module Harness
  管 DDD 模块内部开发
```

这套结构的目标是：

- 减少多人同时修改同一个进度文件的 Git 冲突。
- 降低 agent 上下文污染。
- 保护模块 owner 的边界。
- 让跨模块协作通过契约发生。
- 让每个开发者和每个 agent 都能在自己的沙盒里稳定推进。

