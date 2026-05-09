# AGENTS.md - ego 后端开发枢纽

欢迎进入后端沙盒。本项目基于 **Go 模块化单体 + DDD (领域驱动设计) + 六边形架构** 构建。

## 后端核心模式与硬约束 (Key Patterns & Hard Constraints)
- **开发流程必须符合项目 Harness 规范**
- **代码实现中保持必要的注释以及日志追踪**
- **日志使用规范**：基于 slog + zap 的结构化日志，通过 context propagation 自动携带 request trace 信息。logger 不作为 struct 成员，统一通过 `logging.FromContext(ctx)` 获取，详细设计见 `platform/logging/ARCHITECTURE.md`。
- **后端 DDD 隔离与跨模块协作**：各业务模块拥有独立的表写入权，严禁越权修改其他模块的内部实现。模块间协作必须遵循以下路径：
  - **同步调用**：仅限参考并调用目标模块 `CONTRACT.md` 中描述的接口和公开接口定义。
  - **异步解耦**：对于非强一致性需求，优先通过 `internal/platform/eventbus` 发布领域事件（Domain Event）进行通信。
- **后端依赖倒置与 Platform 规范**：
  - **端口与适配器 (Ports & Adapters)**：业务领域层（`domain`）仅定义需求接口（Port）。
    - **业务专属适配实现**：如涉及到业务实体转换的仓储实现（Repository Impl），必须放在各模块自身的 `adapter/` 目录下。
    - **纯技术底座实现**：`platform` 仅负责提供无业务语境的技术适配（如 `PasswordHasher`）或技术原子（如 `sqlc.Queries`）。业务层（`domain` / `app`）严禁直接引入外部技术 SDK。
  - **依赖装配与注入**：`internal/bootstrap/` 是全后端唯一允许进行依赖组装（Wiring）的地方。
    - 只有在这里，才能将 `platform` 的技术实例注入到各业务模块的 `adapter` 中。
    - 只有在这里，才能将装配好的 `adapter` 注入给具体的业务用例（UseCase）。

- **禁止在此目录下直接开发。必须根据业务归属，进入相应的限界上下文沙盒进行具体业务的开发任务**。


## 文档导航 (Module Routing)

- **server 后端架构设计和目录结构文档** -> `docs/server-structure.md`
- **项目 Harness 规范及开发工作流** ➔ 查阅 `docs/harness-system.md`
- **模块内部设计细节** ➔ 查阅对应模块下的 `ARCHITECTURE.md`
- **模块对外暴露接口** ➔ 查阅对应模块下的 `CONTRACT.md`
- **模块的开发功能列表及状态** ➔ 查阅对应模块下的 `.harness/feature_list.json`
- **模块当前的开发进度** ➔ 查阅对应模块下的 `.harness/progress.md`

## 后端核心命令
```sh
cd ego/server
go test ./...           # 运行后端所有单元测试
make build              # 构建后端 API Server 与 Migrate 迁移脚本
go run cmd/ego/main.go  # 本地启动后端服务
```