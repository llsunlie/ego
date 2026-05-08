# writing Contract

## Owned writes

- `moments` — Writing 拥有唯一写入权
- `traces` — Writing 拥有唯一写入权
- `echos` — Writing 拥有唯一写入权
- `insights` — Writing 拥有唯一写入权（此刻会话级 Insight）

其他模块不允许直接 INSERT、UPDATE 或 DELETE 上述表。

## RPC owned by Writing

| RPC | Input | Output | Notes |
| --- | --- | --- | --- |
| `CreateMoment` | `content` + optional `trace_id` | `Moment` + `Echo` (可能 nil) | 自动创建或延续 Trace，匹配回声并持久化 |
| `GenerateInsight` | `moment_id` + `echo_id` | `Insight` | 生成并持久化当前会话级观察 |
| `ListTraces` | `cursor` + `page_size` | `Trace[]` + `next_cursor` + `has_more` | 游标分页查询用户 Trace 列表 |
| `GetTraceDetail` | `trace_id` | `Trace` + `TraceItem[]` | 返回 Trace 详情，每个 Item 包含 Moment + Echo[] + Insight? |

## Exposed to other modules

### MomentReader

```go
type MomentReader interface {
    GetByID(ctx, id string) (*domain.Moment, error)
    ListByUserID(ctx, userID string, cursor string, pageSize int32) ([]domain.Moment, string, bool, error)
    RandomByUserID(ctx, userID string, count int32) ([]domain.Moment, error)
}
```

适用模块：Timeline（`GetRandomMoments`）、Starmap、Conversation

### TraceReader

```go
type TraceReader interface {
    GetTraceByID(ctx, id string) (*domain.Trace, error)
    ListMomentsByTraceID(ctx, traceID string) ([]domain.Moment, error)
    ListTracesByUserID(ctx, userID string, cursor string, pageSize int32) ([]domain.Trace, string, bool, error)
}
```

适用模块：Starmap（`StashTrace` 时读取 Trace 下全部 Moment）

## Reads from other modules

- 通过 gRPC auth interceptor 注入的 `ctx.Value("user_id")` 获取当前用户身份（来自 Identity 模块）

## Constraints for other modules

- 禁止直接写入 `moments`、`traces`、`echos`、`insights` 表
- 禁止直接 import Writing 模块的 `adapter/postgres` 或 `domain` 内部实现
- 只允许通过本文件声明的 `MomentReader` / `TraceReader` 接口读取数据
- Moment 中 `Embeddings` 字段为内部字段，调用方不应依赖
