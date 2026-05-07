## 标准工作流规范 (Standard Workflow)

**强制要求：禁止直接在项目根目录下进行业务逻辑开发。Agent 必须进入对应的子沙盒目录作业。**

### 1 前端开发流程
1. **进入沙盒**：执行 `cd client/`。
2. **状态同步**：阅读 `client/AGENTS.md` 与 `.harness/progress.md` 确定当前进度。
3. **任务执行**：依据 `.harness/feature_list.json` 执行任务，**数据结构必须严格对齐 `proto/` 定义**。

### 2 后端模块开发流程 (以 writing 为例)
1. **进入沙盒**：执行 `cd server/internal/writing/`。
2. **定义端口 (Port)**：根据业务需求在 `domain/` 或 `app/` 定义所需的抽象接口（如 `EmbeddingProvider`）。
3. **依赖倒置执行**：进行业务逻辑开发，业务代码仅面向自身定义的接口编程，严禁直接引用 `platform` 中的具体实现类。
4. **交接存档**：会话结束前更新 `.harness/progress.md`。

### 3 全栈功能开发流程
1. **契约阶段**：修改 `proto/` 中的定义，更新其 `.harness/feature_list.json` 并运行 `make proto` 同步桩代码。
2. **后端实现**：在对应后端模块沙盒内完成业务逻辑，确保通过局部单元测试。
3. **前端接入**：在前端沙盒内完成 UI 绑定与联调。
4. **集成登记**：更新根目录 `.harness/integration-progress.md`，在全局视角确认功能闭环。