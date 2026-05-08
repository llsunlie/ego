# writing Architecture

Bounded context: Writing — 此刻写作上下文。

> 用户刚写下的话如何成为 Moment，并与过去产生回声？

## 1. 领域模型 (Domain Model)

### 1.1 聚合根: Trace

一条 Trace 代表一次连续的思考会话。用户首次写下 Moment 时自动创建 Trace，后续写下的 Moment 可以选择延续已有 Trace（"顺着再想想"）。

```go
type Trace struct {
    ID         string    // UUID
    UserID     string    // 所属用户
    Motivation string    // 来源：'direct' | 'trace:<id>' | 'constellation:<id>'
    Stashed    bool      // 是否已收进星图（由 Starmap 更新）
    CreatedAt  time.Time
}
```

生命周期：`created`（motivation=direct）→ `active`（持续追加 Moment）→ `stashed`（用户寄存进星图后，由 Starmap 更新）

### 1.2 实体: Moment

用户写下的单条话语，归属于某个 Trace。

```go
type Moment struct {
    ID         string            // UUID
    TraceID    string            // 所属 Trace
    UserID     string            // 所属用户（冗余，便于跨模块只读查询）
    Content    string            // 用户写的原话
    Embeddings []EmbeddingEntry  // 多模型向量组，JSONB 存储，不对外暴露
    CreatedAt  time.Time
}

type EmbeddingEntry struct {
    Model     string    `json:"model"`     // 模型名，如 'text-embedding-3-small'
    Embedding []float32 `json:"embedding"` // 向量
}
```

- `Embeddings` 由 `EmbeddingGenerator` 端口生成，JSONB 持久化后可用于 Echo 匹配。
- Embedding 不通过 API 暴露给客户端。

### 1.3 实体: Echo

Echo 是匹配结果，现在已持久化到 `echos` 表。

```go
type Echo struct {
    ID               string    // UUID
    MomentID         string    // 为哪个 Moment 匹配的
    UserID           string    // 所属用户
    MatchedMomentIDs []string  // 匹配到的历史 Moment ID（按相似度降序）
    Similarities     []float64 // 对应相似度
    CreatedAt        time.Time
}
```

### 1.4 实体: Insight

Writing 模块负责生成"此刻会话级"Insight，区别于 Starmap 的星座级 Insight。持久化到 `insights` 表。

```go
type Insight struct {
    ID               string    // UUID
    UserID           string    // 所属用户
    MomentID         string    // 基于哪个 Moment 生成
    EchoID           string    // 基于哪个 Echo 生成
    Text             string    // 第二人称观察文本
    RelatedMomentIDs []string  // 关联的 Moment ID 列表
    CreatedAt        time.Time
}
```

### 1.5 值对象: MatchedMoment / TraceItem

```go
type MatchedMoment struct {
    MomentID   string
    Similarity float64
}

// TraceItem 聚合 Moment + Echo[] + Insight（用于 GetTraceDetail 响应）
type TraceItem struct {
    Moment  Moment
    Echos   []Echo
    Insight *Insight
}
```

## 2. 领域端口 (Domain Ports)

定义在 `domain/ports.go`。

```go
// TraceRepository — Trace 持久化契约
type TraceRepository interface {
    Create(ctx, *Trace) error
    GetByID(ctx, id string) (*Trace, error)
    Update(ctx, *Trace) error
    Delete(ctx, id string) error
}

// MomentRepository — Moment 持久化契约
type MomentRepository interface {
    Create(ctx, *Moment) error
    GetByID(ctx, id string) (*Moment, error)
    ListByTraceID(ctx, traceID string) ([]Moment, error)
    ListByUserID(ctx, userID string) ([]Moment, error)
}

// EchoRepository — Echo 持久化契约
type EchoRepository interface {
    Create(ctx, *Echo) error
    FindByMomentID(ctx, momentID string) (*Echo, error)
}

// InsightRepository — Insight 持久化契约
type InsightRepository interface {
    Create(ctx, *Insight) error
    FindByMomentID(ctx, momentID string) (*Insight, error)
}

// MomentReader — 对外只读契约（供 Timeline、Starmap、Conversation 使用）
type MomentReader interface {
    GetByID(ctx, id string) (*Moment, error)
    ListByUserID(ctx, userID string, cursor string, pageSize int32) ([]Moment, string, bool, error)
    RandomByUserID(ctx, userID string, count int32) ([]Moment, error)
}

// TraceReader — 对外只读契约（供 Starmap 使用 + ListTraces/GetTraceDetail RPC）
type TraceReader interface {
    GetTraceByID(ctx, id string) (*Trace, error)
    ListMomentsByTraceID(ctx, traceID string) ([]Moment, error)
    ListTracesByUserID(ctx, userID string, cursor string, pageSize int32) ([]Trace, string, bool, error)
}

// EmbeddingGenerator — 向量生成端口
type EmbeddingGenerator interface {
    Generate(ctx, content string) ([]EmbeddingEntry, error)
}

// EchoMatcher — 回声匹配领域服务
type EchoMatcher interface {
    Match(ctx, current *Moment, history []Moment) ([]MatchedMoment, error)
}

// InsightGenerator — 洞察生成端口
type InsightGenerator interface {
    Generate(ctx, momentID string, echoID string) (*Insight, error)
}
```

## 3. 应用层用例 (Application Use Cases)

### 3.1 CreateMoment

```
Input:  content (string), traceID (optional)
Output: saved Moment + Echo (可能为 nil)
Flow:
  1. 如果 traceID 为空 → 创建新 Trace（motivation='direct', stashed=false）
  2. 如果 traceID 不为空 → 加载已有 Trace（不存在返回错误）
  3. 创建 Moment
  4. 调用 EmbeddingGenerator 生成向量 → JSONB 写入 Moment.Embeddings
  5. 持久化 Moment
  6. 调用 EchoMatcher 匹配当前用户历史 Moment → 构建并持久化 Echo
  7. 返回 Moment + Echo
```

### 3.2 GenerateInsight

```
Input:  momentID (string), echoID (string), userID (string)
Output: persisted Insight
Flow:
  1. 调用 InsightGenerator.Generate(ctx, momentID, echoID) → 获取 Insight
  2. 填充 ID、MomentID、EchoID、UserID、CreatedAt
  3. 持久化到 insights 表
  4. 返回 Insight
```

## 4. 适配器 (Adapters)

### 4.1 adapter/postgres

| 文件 | 实现接口 |
|---|---|
| `trace_repo.go` | `TraceRepository` |
| `moment_repo.go` | `MomentRepository` |
| `echo_repo.go` | `EchoRepository` |
| `insight_repo.go` | `InsightRepository` |
| `reader.go` | `MomentReader` + `TraceReader`（组合实现） |

### 4.2 adapter/grpc

- `handler.go` — 实现 `CreateMoment`, `GenerateInsight`, `ListTraces`, `GetTraceDetail` 四个 RPC
- `mapper.go` — Proto DTO ↔ Domain Model 转换

## 5. 数据表

| 表 | 迁移文件 | 归属 |
|---|---|---|
| `traces` | `003_traces.sql` | Writing |
| `moments` | `002_moments.sql` | Writing |
| `echos` | `004_echos.sql` | Writing |
| `insights` | `005_insights.sql` | Writing |

## 6. 依赖方向

```
adapter/grpc → app → domain
adapter/postgres → platform/postgres (sqlc)
adapter/postgres → domain
app → domain
```

禁止：
- domain 层 import proto、sqlc、pgx、platform
- app 层 import adapter
- adapter 之间相互依赖
