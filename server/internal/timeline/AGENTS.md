# AGENTS.md - Timeline 模块开发指南

## 0. 全局上下文溯源

如果你是在 `server/internal/timeline/` 目录下被唤醒的 Agent，请在执行任何代码前先向上读取：

1. `../../../AGENTS.md`
2. `../../AGENTS.md`


Timeline 是查询上下文，核心边界是"只读"。

## 1. 模块定位

Timeline 负责"过往"页面和记忆光点盲盒。

它回答：

> 用户如何回看自己过去说过的话？

本模块不拥有 Moment 的创建权，只负责面向前端的 Moment 读取视图。

## 2. 负责的接口范围

本模块负责实现以下前后端 RPC：

| RPC | 责任 |
| --- | --- |
| `ListTraces` | 返回当前用户的 Trace 列表，支持游标分页 |
| `GetTraceDetail` | 返回 Trace 详情，聚合 Moment + Echo + Insight |
| `GetRandomMoments` | 返回当前用户随机历史 Moment，用于记忆光点盲盒 |

## 3. 模块边界

### 3.1 拥有的业务能力

- 按时间倒序读取 Trace。
- 支持 cursor/page_size 分页。
- 聚合 Trace 关联的 Moment、Echo、Insight。
- 随机读取 N 条历史 Moment。

### 3.2 数据归属

Timeline 不拥有任何业务表写入权。

允许读取：

```text
traces
moments
echos
insights
```

所有表写入权属于 Writing，Timeline 通过 `writing/adapter/postgres` 的 read adapter 获取数据。

### 3.3 禁止事项

- 禁止创建、更新或删除 Trace、Moment、Echo、Insight。
- 禁止更新 `stashed` 状态。
- 禁止写入星图、星座、聊天相关表。
- 禁止在 Timeline 中实现 Echo 匹配、StashTrace 或 Chat。
- 禁止绕过用户隔离读取其他用户数据。

## 4. 架构与装配

- **两级装配**：`timeline/module.go` 负责组装本模块的 read adapter、app use case、gRPC handler；`bootstrap/timeline.go` 只注入 DB 等进程级资源。
- **只读直通**：Timeline 无复杂业务逻辑，app 层 use case 负责默认值（pageSize=20，count=3）和 TraceItem 组装。
- **Domain ports**：Timeline 定义自己的只读端口（MomentReader、TraceReader、EchoReader、InsightReader），由 writing/adapter/postgres 的 Reader 实现。
- **Mapper 在 adapter/grpc**：proto 映射逻辑在 `mapper.go`，handler 纯粹做 ctx→input→output→pb 转换。

## 5. 常用开发命令

从 `server/` 目录运行：

```text
go test ./internal/timeline/...
go test ./...
go build ./...
```
