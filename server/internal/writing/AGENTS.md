# AGENTS.md - Writing 模块开发指南

## 0. 全局上下文溯源

如果你是在 `server/internal/writing/` 目录下被唤醒的 Agent，请在执行任何代码前先向上读取：

1. `../../../AGENTS.md`
2. `../../AGENTS.md`


## 1. 模块定位

Writing 是"此刻写作上下文"，回答：

> 用户刚写下的话如何成为 Moment，并与过去产生回声？

本模块负责 Trace、Moment、Echo、Insight 的全部生命周期。

## 2. 负责的接口范围

本模块负责实现以下前后端 RPC：

| RPC | 责任 |
| --- | --- |
| `CreateMoment` | 创建或延续 Trace，保存 Moment，匹配回声并持久化 Echo |
| `GenerateInsight` | 基于 Moment + Echo 生成并持久化当前体验的"我发现" |
| `ListTraces` | 游标分页查询用户的 Trace 列表 |
| `GetTraceDetail` | 返回 Trace 详情，聚合 Moment + Echo[] + Insight |

## 3. 模块边界

### 3.1 拥有的业务能力

- 创建 Trace，或延续已有 Trace。
- 追加 Moment。
- 生成或接收 Moment embedding（多模型向量组，JSONB 存储）。
- 匹配当前用户历史 Moment 作为 Echo（持久化到 echos 表）。
- 生成并持久化当前会话级 Insight。
- 为其他模块提供 Moment/Trace 的只读契约（MomentReader / TraceReader）。

### 3.2 数据归属

Writing 拥有唯一写入权：

```
traces
moments
echos
insights
```

其他模块（Starmap、Timeline、Conversation）只能通过明确的只读契约读取，不允许直接创建或更新。

## 4. 常用开发命令

从 `server/` 目录运行：

```
go test ./internal/writing/...
go test ./...
go build ./...
```

从项目根目录运行端到端 smoke 测试：

```
./smoke.sh
```

smoke.sh 会启动 Docker PostgreSQL、执行迁移、编译服务、用 grpcurl 走完整核心流程。

## 5. 关键设计决策

- **embedding 存为 JSONB**：`[]EmbeddingEntry` 序列化为 JSONB，支持多模型向量组。初期在应用层做余弦相似度匹配，后续可切换 pgvector。
- **Echo 持久化**：Echo 从临时值对象变为持久化实体，支持 GetTraceDetail 回查。
- **Insight 持久化**：Writing 生成的会话级 Insight 持久化到 insights 表，与 Starmap 的星座级 Insight 区分。
- **AI stub 默认可用**：bootstrap/writing.go 中 stub 返回固定向量和全量匹配，确保无 AI 服务时逻辑仍可走通。
