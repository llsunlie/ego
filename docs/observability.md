# 可观测性方案

## 架构总览

```
ego-server
  ├── slog (JSON) ──→ server.log 文件 ──→ Promtail ──→ Loki ──→ Grafana (LogQL)
  ├── /metrics 端点 ──────────────────────────────→ Prometheus ──→ Grafana (PromQL)
  └── /health 端点 ───────────────────────────────→ UptimeRobot / LB probes
```

三种信号覆盖：
- **Logs**: 结构化 JSON 日志 → Loki
- **Metrics**: 11 个 Prometheus 指标 → Prometheus
- **Traces**: 未实现（见文末 OpenTelemetry 讨论）

---

## 一、日志（Logging）

### 技术栈

- **Go 层**: `log/slog` + `uber-go/zap` + `zapslog`（Zap 引擎，slog 接口）
- **轮转**: `lumberjack` — 单文件 100MB 自动切分，保留 30 天，旧文件 gzip
- **采集**: Promtail 读取日志文件，JSON 解析后推送 Loki
- **存储**: Loki 3.4.0（TSDB + 本地文件系统）
- **查询**: Grafana Explore (LogQL)

### 日志格式

`LOG_FORMAT=json` 时输出为标准 JSON：

```json
{
  "level": "info",
  "ts": 1779888088.286,
  "msg": "ai.Chat: ok",
  "request_id": "a2c89715-...",
  "user_id": "7965e78f-...",
  "method": "/ego.Ego/Chat",
  "model": "deepseek-v4-flash",
  "prompt_tokens": 642,
  "completion_tokens": 371,
  "elapsed_ms": 6751
}
```

### 配置项

| 环境变量 | 默认值 | 说明 |
|---|---|---|
| `LOG_LEVEL` | `info` | `debug` / `info` / `warn` / `error` |
| `LOG_FORMAT` | `json` | `json` / `text` |
| `LOG_OUTPUT` | `stdout` | 文件路径或 `stdout` / `stderr` |

### 结构化字段

日志中自动携带的字段：

| 字段 | 来源 | 说明 |
|---|---|---|
| `request_id` | `auth/interceptor.go` | 每次 gRPC 请求唯一 UUID |
| `user_id` | JWT 解析 | 调用者 |
| `method` | gRPC info | 如 `/ego.Ego/CreateMoment` |
| `msg` | 业务代码 | 如 `ai.Chat: request` |
| `model` | AI client | 使用的模型名 |
| `prompt_tokens` / `completion_tokens` | AI API 响应 | token 消耗 |
| `elapsed_ms` | AI client | 调用耗时（毫秒） |

### AI 调用日志示例

```
ai.Chat: request  → 记录 model、消息数、用户最后一条消息预览（200 字符）
ai.Chat: ok       → 记录 model、prompt/completion tokens、耗时、响应预览
ai.Chat: error    → 记录错误信息、已耗时
ai.Embed: request → 记录 model、输入长度
ai.Embed: ok      → 记录 model、tokens、耗时
ai.Embed: error   → 记录错误信息
```

### 日志采集链路

```
server/.tmp/logs/server/server.log
  → docker-compose 挂载为 /var/log/ego/
    → Promtail pipeline: JSON 解析 → 提取 level/method 为 Loki 标签
      → 推送 Loki (http://loki:3100/loki/api/v1/push)
        → Grafana Explore 查询
```

### 常用 LogQL

```logql
{job="ego-server"}                          # 全部日志
{job="ego-server"} |= "ERROR"               # 关键词
{job="ego-server"} | json | level = "ERROR"  # 结构化过滤
{job="ego-server"} | msg =~ "ai\\..*"        # AI 调用日志
```

---

## 二、指标（Metrics）

### 技术栈

- **Go 层**: `prometheus/client_golang` — `promauto` 自动注册
- **暴露**: `/metrics` 端点（gRPC-web 同端口 9080）
- **采集**: Prometheus 每 15s 拉取 `host.docker.internal:9080/metrics`
- **查询**: Grafana Explore / Dashboard (PromQL)

### 指标清单（11 个）

#### HTTP 层

| 指标 | 类型 | 标签 |
|---|---|---|
| `ego_http_requests_total` | Counter | method, path, status |
| `ego_http_request_duration_seconds` | Histogram | method, path |
| `ego_http_requests_in_flight` | Gauge | - |

#### gRPC 层

| 指标 | 类型 | 标签 |
|---|---|---|
| `ego_grpc_requests_total` | Counter | method, status |
| `ego_grpc_request_duration_seconds` | Histogram | method |

#### AI Chat

| 指标 | 类型 | 标签 |
|---|---|---|
| `ego_ai_chat_total` | Counter | model, status |
| `ego_ai_chat_duration_seconds` | Histogram | model |
| `ego_ai_chat_tokens_total` | Counter | model, type (prompt/completion) |

#### AI Embedding

| 指标 | 类型 | 标签 |
|---|---|---|
| `ego_ai_embed_total` | Counter | model, status |
| `ego_ai_embed_duration_seconds` | Histogram | model |
| `ego_ai_embed_tokens_total` | Counter | model |

#### 综合

| 指标 | 类型 | 标签 |
|---|---|---|
| `ego_ai_calls_in_flight` | Gauge | - |

### 指标埋点位置

```
server/internal/platform/metrics/metrics.go  ← 指标定义
server/internal/platform/ai/client.go        ← AI 调用埋点
server/internal/bootstrap/server.go          ← HTTP/gRPC 中间件埋点
```

### 常用 PromQL

```promql
# AI Chat QPS
rate(ego_ai_chat_total{status="ok"}[5m])

# AI Chat P99 延迟
histogram_quantile(0.99, rate(ego_ai_chat_duration_seconds_bucket[5m]))

# token 消耗速率（按类型）
rate(ego_ai_chat_tokens_total[5m])

# 整体 HTTP QPS
rate(ego_http_requests_total[1m])

# HTTP P99
histogram_quantile(0.99, rate(ego_http_request_duration_seconds_bucket[5m]))

# 按状态码分组
sum(rate(ego_http_requests_total[1m])) by (status)

# 当前并发 AI 调用
ego_ai_calls_in_flight
```

---

## 三、健康检查

`GET /health` → `{"status":"ok"}`，供 UptimeRobot 或负载均衡器探测。

---

## 四、Docker 服务

| 服务 | 镜像 | 端口 |
|---|---|---|
| Prometheus | `prom/prometheus:v3.2.0` | 9090 |
| Loki | `grafana/loki:3.4.0` | 3100 |
| Promtail | `grafana/promtail:3.4.0` | - |
| Grafana | `grafana/grafana:12.1.0` | 3200:3000 |

启动方式：

```powershell
.\restart-monitoring.ps1                     # 单独启动监控栈
.\start.ps1 -WithMonitoring                  # 全栈启动（含监控）
.\restart-monitoring.ps1 -Down                # 停止并移除监控栈
```

配置文件：

| 文件 | 用途 |
|---|---|
| `monitoring/prometheus.yml` | Prometheus 抓取配置 |
| `monitoring/loki-config.yml` | Loki 存储配置 |
| `monitoring/promtail-config.yml` | Promtail 采集 + JSON 解析 pipeline |
| `monitoring/grafana-datasources.yml` | Grafana 自动配置数据源 |

---

## 五、OpenTelemetry 讨论（未实现）

### 是什么

CNCF 开源可观测性标准，统一采集 **Traces + Metrics + Logs** 三种信号。核心用途是**分布式调用链可视化**。

### 与现有方案的关系

```
                  ┌─ Prometheus（存储/查询）      ← 保留
OTel SDK（埋点） ──┼─ Jaeger/Tempo（调用链存储） ← 新增
                  └─ Loki（日志存储）            ← 保留
```

- OTel 是**埋点标准**，不是存储后端
- Prometheus、Loki 继续保留，OTel 取而代之的是手写的 `promauto` + `slog` 代码
- 新增 Jaeger 或 Tempo 存储和查看调用链

### 需要新增

| 层面 | 内容 | 工作量 |
|---|---|---|
| Go 依赖 | `otel`、`otelgrpc` 等 5-6 个包 | 小 |
| 代码改造 | gRPC interceptor 换 `otelgrpc`，业务 Span 替代手写日志/指标 | 中 |
| Docker 服务 | Jaeger（`jaegertracing/all-in-one`）或 Grafana Tempo | 小 |
| Grafana 配置 | 添加 Jaeger/Tempo 数据源 | 小 |

### 调用链效果

每次 gRPC 请求自动生成 Span 树：

```
CreateMoment ──────────────────────────── 300ms
 ├── auth interceptor ─ 1ms
 ├── CreateMoment ───── 5ms
 │   ├── Embedding API ─────── 161ms
 │   └── DB INSERT ── 3ms
 ├── matchEcho ─────── 20ms
 │   ├── DB SELECT ──── 12ms
 │   └── 相似度计算 ─── 8ms
 └── serialize ─────── 1ms
```

### 推荐方案（如实施）

Jaeger all-in-one：一个容器同时提供 OTLP 接收端 + 存储 + UI，不依赖外部存储。UI 端口 `16686`，Go SDK 通过 OTLP gRPC 协议（`4317`）发送数据。

### ROI 评估

- 当前规模：单进程后端，`request_id` 已能串联单次请求的日志
- OTel 价值：调用链甘特图可视化，一眼定位瓶颈（DB 慢还是 AI API 慢）
- 代价：6 个新依赖 + 1 个容器 + 代码改造
- 结论：**当前 Prometheus + Loki 已覆盖 80% 需求**，OTel 在性能诊断、调用链可视化场景有增量价值，暂不实施
