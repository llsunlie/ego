# writing Contract

## Owned writes

- `moments` — Writing 拥有唯一写入权
- `traces` — Writing 拥有唯一写入权

其他模块不允许直接 INSERT、UPDATE 或 DELETE `moments` 和 `traces` 行。

## RPC owned by Writing

| RPC | Input | Output | Notes |
| --- | --- | --- | --- |
| `CreateMoment` | `content` + optional `trace_id`, `topic` | `Moment` + `Echo` | 自动创建或延续 Trace，匹配回声 |
| `GenerateInsight` | `current_content` + `echo_moment_id` | `Insight` | 实时生成当前会话级观察，不持久化 |

## Exposed to other modules

### MomentReader

```go
type MomentReader interface {
    GetByID(ctx context.Context, id string) (*domain.Moment, error)
    ListByUserID(ctx context.Context, userID string, cursor string, pageSize int32) (moments []domain.Moment, nextCursor string, hasMore bool, err error)
    RandomByUserID(ctx context.Context, userID string, count int32) ([]domain.Moment, error)
}
```

适用模块：Timeline（`ListMoments`、`GetRandomMoments`）、Starmap、Conversation

### TraceReader

```go
type TraceReader interface {
    GetTraceByID(ctx context.Context, id string) (*domain.Trace, error)
    ListMomentsByTraceID(ctx context.Context, traceID string) ([]domain.Moment, error)
}
```

适用模块：Starmap（`StashTrace` 时读取 Trace 下全部 Moment）

## Reads from other modules

- 通过 gRPC auth interceptor 注入的 `ctx.Value("user_id")` 获取当前用户身份（来自 Identity 模块）
- Starmap 在 `StashTrace` 后通过领域事件更新 Moment 的 `connected` 状态

## Constraints for other modules

- 禁止直接写入 `moments`、`traces` 表
- 禁止直接 import Writing 模块的 `adapter/postgres` 或 `domain` 内部实现
- 只允许通过本文件声明的 `MomentReader` / `TraceReader` 接口读取数据
- 读取的 Moment 中 `Embedding` 字段为内部字段，调用方不应依赖
