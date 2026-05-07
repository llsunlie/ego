## 目录结构 (Repository Structure)

- `proto/` — 前后端 API 通信契约定义目录
- `client/` — 项目前端代码目录
- `server/` — Go 模块化单体后端服务沙盒
  - `cmd/` — 进程入口（`ego` 主服务与 `migrate` 数据迁移工具）
  - `internal/bootstrap/` — 依赖注入与组件装配层
  - `internal/config/` — 环境变量与配置读取层
  - `internal/shared/` — 极少量全局共享领域类型与接口（对业务模块只读）
  - `internal/platform/` — 纯技术基础设施层（PostgreSQL、gRPC、AI 防腐层、EventBus 等）
  - `internal/identity/` — 身份与鉴权限界上下文 (Login, User 等)
  - `internal/writing/` — 核心写作与回声匹配限界上下文 (Trace, Moment, Echo 等)
  - `internal/timeline/` — 过往查询限界上下文 (列表展示，纯查询无写入权)
  - `internal/starmap/` — 星图沉淀与聚类限界上下文 (Star, Constellation, Insight 等)
  - `internal/conversation/` — 跨时空对话管理限界上下文 (ChatSession, ChatMessage 等)
- `.harness/` — 全局跨模块集成进度与任务索引配置

> **附：后端业务模块内部标准结构** (以 `internal/{module}/` 为例)
> - `domain/` — 业务核心 (Entity, Value Object, Repo Interface)，**严禁**依赖具体的基础设施技术或 proto。
> - `app/` — 用例编排层 (UseCase)，调度 Domain Port，不写具体 SQL 或底层 API 请求。
> - `adapter/` — 端口适配器层，通常包含 `grpc/` (处理输入请求) 与 `postgres/` (实现持久化仓储)。