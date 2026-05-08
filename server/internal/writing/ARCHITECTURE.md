# writing Architecture

Bounded context: Writing — 此刻写作上下文。

> 用户刚写下的话如何成为 Moment，并与过去产生回声？

## 1. 领域模型 (Domain Model)

### 1.1 聚合根: Trace

一条 Trace 代表一次连续的思考会话。用户首次写下 Moment 时自动创建 Trace，后续写下的 Moment 可以选择延续已有 Trace（"顺着再想想"）。

```go
type Trace struct {
    ID        string    // UUID
    UserID    string    // 所属用户
    Topic     string    // 可选，来自 TopicPrompt 的话题引子
    CreatedAt time.Time
    UpdatedAt time.Time
}
```

生命周期：`created` → `active`（持续追加 Moment）→ `stashed`（用户寄存进星图后，由 Starmap 更新）

### 1.2 实体: Moment

用户写下的单条话语，归属于某个 Trace。

```go
type Moment struct {
    ID        string    // UUID
    TraceID   string    // 所属 Trace
    UserID    string    // 所属用户（冗余，便于跨模块只读查询）
    Content   string    // 用户写的原话
    Embedding []float32 // 内部向量，不对外暴露
    Connected bool      // 是否已联结进星座（由 Starmap 更新）
    CreatedAt time.Time
}
```

- `Embedding` 由 `EmbeddingGenerator` 端口生成，持久化后用于 Echo 匹配。
- `Connected` 字段由 Writing 模块维护默认值 `false`，由 Starmap 在 `StashTrace` 后通过领域事件或契约更新（不在 Writing 内部修改）。

### 1.3 值对象: Echo

Echo 是匹配结果，不独立持久化。

```go
type Echo struct {
    ID          string    // 匹配到的历史 Moment ID
    TargetMoment Moment   // 匹配到的历史 Moment
    Candidates   []Moment // 2-3 条候选回声
    Similarity   float64  // 0-1 相似度
}
```

### 1.4 实体: Insight

Writing 模块负责生成"此刻会话级"Insight（区别于 Starmap 的星座级 Insight）。

```go
type Insight struct {
    ID              string   // UUID
    Text            string   // 第二人称观察文本
    RelatedMomentIDs []string // 关联的 Moment ID 列表
}
```

Writing 生成的 Insight 实时返回给前端，不在此模块内持久化。星座级 Insight 由 Starmap 拥有并持久化到 `insights` 表。

## 2. 领域端口 (Domain Ports)

端口定义在 `domain/` 层，由 `adapter/` 或 `platform/` 实现。

```go
// TraceRepository — Trace 持久化契约
type TraceRepository interface {
    Create(ctx context.Context, trace *Trace) error
    GetByID(ctx context.Context, id string) (*Trace, error)
    Update(ctx context.Context, trace *Trace) error
}

// MomentRepository — Moment 持久化契约
type MomentRepository interface {
    Create(ctx context.Context, moment *Moment) error
    GetByID(ctx context.Context, id string) (*Moment, error)
    ListByTraceID(ctx context.Context, traceID string) ([]Moment, error)
    // ListByUserID 用于 Echo 匹配时检索当前用户的所有历史 Moment
    ListByUserID(ctx context.Context, userID string) ([]Moment, error)
}

// MomentReader — 对外只读契约（供 Timeline、Starmap、Conversation 使用）
type MomentReader interface {
    GetByID(ctx context.Context, id string) (*Moment, error)
    ListByUserID(ctx context.Context, userID string, cursor string, pageSize int32) ([]Moment, string, bool, error)
    RandomByUserID(ctx context.Context, userID string, count int32) ([]Moment, error)
}

// TraceReader — 对外只读契约（供 Starmap 使用）
type TraceReader interface {
    GetByID(ctx context.Context, id string) (*Trace, error)
    ListMomentsByTraceID(ctx context.Context, traceID string) ([]Moment, error)
}

// EmbeddingGenerator — 向量生成端口
type EmbeddingGenerator interface {
    Generate(ctx context.Context, content string) ([]float32, error)
}

// EchoMatcher — 回声匹配领域服务
type EchoMatcher interface {
    Match(ctx context.Context, current *Moment, history []Moment) (*Echo, error)
}

// InsightGenerator — 洞察生成端口
type InsightGenerator interface {
    Generate(ctx context.Context, currentContent string, echoMomentID string) (*Insight, error)
}
```

## 3. 应用层用例 (Application Use Cases)

用例编排在 `app/` 层，负责事务边界、调用端口、发布领域事件。

### 3.1 CreateMoment

```
Input:  content (string), traceID (optional), topic (optional)
Output: saved Moment + Echo
Flow:
  1. 如果 traceID 为空 → 创建新 Trace
  2. 如果 traceID 不为空 → 加载已有 Trace（不存在返回错误）
  3. 创建 Moment（状态: connected=false）
  4. 调用 EmbeddingGenerator 生成向量并写入 Moment
  5. 持久化 Moment
  6. 调用 EchoMatcher 匹配历史 Moment → 生成 Echo
  7. 返回 Moment + Echo
```

### 3.2 GenerateInsight

```
Input:  currentContent (string), echoMomentID (string)
Output: Insight
Flow:
  1. 调用 InsightGenerator 基于当前内容 + 回声 Moment 生成观察
  2. 返回 Insight
```

## 4. 适配器 (Adapters)

### 4.1 adapter/postgres

- `trace_repo.go` — TraceRepository 的 PostgreSQL 实现（sqlc 适配）
- `moment_repo.go` — MomentRepository 的 PostgreSQL 实现（sqlc 适配）
- `reader.go` — MomentReader + TraceReader 的组合实现

### 4.2 adapter/grpc

- `handler.go` — gRPC handler 实现 `pb.EgoServer` 接口中的 `CreateMoment` 和 `GenerateInsight`
- `mapper.go` — Proto DTO ↔ Domain Model 转换

## 5. 依赖方向

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
