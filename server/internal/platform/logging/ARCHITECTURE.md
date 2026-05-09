# logging Architecture

基于 **slog (API) + zap (backend)** + **context propagation**。`*slog.Logger` 不作为 struct 成员持有，通过 context 传递，确保每条请求日志自动携带 request_id / user_id / method。

## 架构

```
gRPC interceptor (platform/auth)
  └── baseLogger.With("request_id", "user_id", "method")
        └── logging.WithLogger(ctx, requestLogger)
              └── handler → app → adapter/postgres
                    └── logging.FromContext(ctx)
```

## 配置

| 环境变量 | 默认值 | 可选值 | 说明 |
|----------|--------|--------|------|
| `LOG_LEVEL` | `info` | `debug` / `info` / `warn` / `error` | 日志级别 |
| `LOG_FORMAT` | `text` | `text` / `json` | 输出格式 |
| `LOG_OUTPUT` | `stdout` | `stdout` / `stderr` / `/path/to/file` | 输出目标 |

**格式效果：**

`LOG_FORMAT=text` → ConsoleEncoder（彩色、时间、caller）：
```
2026-05-09T12:12:52.123+0800	INFO	logging/context_test.go:19	from context
```

`LOG_FORMAT=json` → JSONEncoder：
```json
{"level":"info","ts":1746762041.123,"msg":"from context","key":"value"}
```

`LOG_OUTPUT=/var/log/ego/app.log` → 追加写入文件（不存在则创建）。

## API

| 函数 | 说明 |
|------|------|
| `New(cfg Config) (*slog.Logger, error)` | 创建 zap-backed logger |
| `NewDefault() *slog.Logger` | 开发环境默认 logger（text、debug、caller） |
| `NewNop() *slog.Logger` | 测试用空 logger |
| `WithLogger(ctx, *slog.Logger) context.Context` | 注入 context |
| `FromContext(ctx) *slog.Logger` | 从 context 提取，无则 fallback 到 `slog.Default()` |

## 各层使用

**handler** — interceptor 已注入 ctx：
```go
logger := logging.FromContext(ctx)
logger.InfoContext(ctx, "create moment request", "trace_id", req.TraceId)
```

**app** — 不持有 logger 字段，仅记录关键业务事件：
```go
logger := logging.FromContext(ctx)
logger.InfoContext(ctx, "moment created", "moment_id", id)
```

**adapter/postgres** — DB 查询等基础设施事件用 Debug：
```go
logger := logging.FromContext(ctx)
logger.DebugContext(ctx, "saving moment", "moment_id", id)
```

**domain** — 禁止使用 logger。

**启动阶段** — 无 ctx 时用 `Platform.Logger`：
```go
p.Logger.Info("ego server starting", "port", cfg.Port)
```

## 级别约定

| 级别 | 用途 |
|------|------|
| DEBUG | SQL 细节、AI 请求参数、缓存命中 |
| INFO | 关键业务事件 |
| WARN | 可恢复异常（重试、降级） |
| ERROR | 请求失败、不可恢复错误 |

## 测试

```go
// 静默日志
ctx := logging.WithLogger(context.Background(), logging.NewNop())

// 验证日志输出
var buf bytes.Buffer
logger, _ := logging.New(logging.Config{Level: "debug", Format: "text", Output: &buf})
ctx := logging.WithLogger(context.Background(), logger)
```
