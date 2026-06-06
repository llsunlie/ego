---
name: ego-svr-timeline
description: 服务端 timeline 领域 context — 只读查询，Trace 列表（分页/按月分组）、Trace 详情、随机 Moment。对应前端 client page: past。
---

# ego-cli-server-timeline

timeline 有界上下文 — 历史数据查询服务（只读）。DDD 结构：`server/internal/timeline/`

## 所属 gRPC 方法

- `ListTraces` — 分页查询 trace 列表（cursor-based pagination）
- `GetTraceDetail` — 查询 trace 的完整 moment → echo → insight 链路
- `GetRandomMoments` — 随机获取 N 个 historical moments（记忆光点）

## 模块结构 (`server/internal/timeline/`)

```
timeline/
├── module.go                      # 依赖注入
├── domain/
│   └── ports.go                   # TraceReader 接口
├── app/
│   ├── list_traces.go             # ListTraces 用例
│   ├── get_trace_detail.go        # GetTraceDetail 用例
│   └── get_random_moments.go      # GetRandomMoments 用例
└── adapter/
    ├── grpc/handler.go            # gRPC Handler
    └── grpc/mapper.go             # proto ↔ domain 映射
```

## 架构特点

timeline 是纯 **read-only** 模块：
- 没有自己的 repository — 复用 writing 模块暴露的 `MomentReader`、`TraceReader`、`EchoRepository`、`InsightRepository`
- 没有 domain types — 直接使用 writing 模块的 domain 类型
- 应用层只有 3 个查询用例

## 模块组装 (`module.go`)

```go
type Deps struct {
    DB sqlc.DBTX
}

func NewHandler(deps Deps) *timelinegrpc.Handler {
    queries := sqlc.New(deps.DB)
    reader      := writingpostgres.NewReader(queries)         // MomentReader + TraceReader
    echoRepo    := writingpostgres.NewEchoRepository(queries)
    insightRepo := writingpostgres.NewInsightRepository(queries)

    listTraces      := timelineapp.NewListTracesUseCase(reader)
    getTraceDetail  := timelineapp.NewGetTraceDetailUseCase(reader, echoRepo, insightRepo)
    getRandomMoments := timelineapp.NewGetRandomMomentsUseCase(reader)

    return timelinegrpc.NewHandler(listTraces, getTraceDetail, getRandomMoments)
}
```

直接依赖 `writing/adapter/postgres` 的实现（跨模块依赖）。

## ListTraces 用例

- cursor-based 分页（`pageSize` 默认 20）
- 返回数据由前端按月份分组展示
- `hasMore` 标识是否还有更多数据

## GetTraceDetail 用例

- 输入 `traceId` → 返回完整 timeline：
  - Trace → Moments → Echos（含 matchedMomentIds）→ Insights
- 使用 `EchoRepository` 和 `InsightRepository` 查询关联数据

## GetRandomMoments 用例

- 随机返回 N 个用户的 historical moments
- 用于 Now 页面的「记忆光点」

## 相关文件

| 文件 | 说明 |
|------|------|
| `server/internal/writing/domain/ports.go` | MomentReader, TraceReader 接口定义 |
| `server/internal/writing/adapter/postgres/reader.go` | Reader 实现（sqlc 查询） |
| `server/internal/bootstrap/timeline.go` | 顶层 wiring |
