# AGENTS.md - Starmap 模块开发指南

## 0. 全局上下文溯源

如果你是在 `server/internal/starmap/` 目录下被唤醒的 Agent，请在执行任何代码前先向上读取：

1. `../../../AGENTS.md`
2. `../../AGENTS.md`

Starmap 是 ego 的核心领域之一。不要把 Moment 创建或 ChatSession 管理混入本模块。

## 1. 模块定位

Starmap 是“星图上下文”，回答：

> 一次完整 Trace 如何被寄存为星星，并逐渐形成星座？

本模块负责 Star、Constellation、星座级 Insight、PastSelfCard、TopicPrompt，以及星图/星座详情查询。

## 2. 负责的接口范围

本模块负责实现以下前后端 RPC：

| RPC | 责任 |
| --- | --- |
| `StashTrace` | 将 Writing 中的一次 Trace 收进星图，创建 Star，并触发星座重评估 |
| `ListConstellations` | 返回星图页所需的星座与星星列表 |
| `GetConstellation` | 返回星座详情，包括观察、原话列表、PastSelfCard 和 TopicPrompt |

## 3. 模块边界

### 3.1 拥有的业务能力

- 校验 Trace 是否可寄存。
- 创建 Star。
- 将 Star 聚类到 Constellation。
- 维护星座状态：`lone`、`forming`、`formed`。
- 生成或刷新星座级 Insight。
- 生成 PastSelfCard。
- 生成 TopicPrompt。
- 组装星图和星座详情 read model。

### 3.2 数据归属

Starmap 拥有唯一写入权：

```text
stars
constellations
constellation_stars
insights
past_self_cards
topic_prompts
```

Starmap 可以通过 Writing 契约只读 Trace/Moment。

### 3.3 禁止事项

- 禁止创建、更新或删除 Moment。
- 禁止直接拥有 `moments` 或 `traces` 的写入权。
- 禁止创建或保存 ChatSession/ChatMessage。
- 禁止让前端飞星动效依赖聚类和 AI 资产生成同步完成。
- 禁止在 `domain/` 中直接调用 AI SDK、pgx、sqlc 或 proto。

## 4. 依赖规则

- 读取 Trace/Moment 必须通过 Writing 暴露的只读接口。
- Conversation 需要 PastSelfCard 时，应读取 Starmap 暴露的契约，而不是访问内部实现。
- 聚类和星座资产生成应通过领域事件或异步任务触发；MVP 可用进程内事件，后续可升级 outbox。
- 需要 AI 生成星座名称、观察、卡片或话题时，通过 Port 调用。

## 5. 常用开发命令

从 `server/` 目录运行：

```text
go test ./internal/starmap/...
go test ./...
go build ./...
```

