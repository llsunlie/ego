本文件旨在为进入当前仓库的 AI 编程智能体（Agent）及人类开发者提供全局开发指南、核心约束与文档路由。

## 项目概览 (Project Overview)

ego 是一个基于个人语境的独白空间应用，核心逻辑是通过“已检索记忆”实现用户与过去自我的对话。项目采用前后端分离（Flutter + Go）架构，后端严格遵循领域驱动设计（DDD）模块化构建，全仓库实施分层 Harness 规范管理。


## 核心模式与硬约束 (Key Patterns & Hard Constraints)

- **Proto 唯一事实来源**：前后端数据结构与接口交互必须以 `proto/ego/api.proto` 为准，严禁在各端代码中私自伪造、猜测或硬编码字段。
- **分层 Harness 规范**：Agent 的具体开发任务、单元测试与进度记录仅限在模块目录的 `.harness/` 沙盒中进行。当前项目根目录的 `.harness/` 仅用于记录项目前后端跨端联调的状态和情况。

## 路由与专题文档索引 (Routing / Document Index)

请根据当前任务，按需阅读以下文档，不要盲目猜测：

- **了解全局目录结构与模块说明** ➔ 查阅 `docs/architecture-map.md`
- **了解项目 Harness 规范及开发工作流** ➔ 查阅 `docs/harness-system.md`
- **查看各端具体的开发工作流步骤** ➔ 查阅 `docs/workflow-rules.md`
- **修改接口或契约** ➔ 查阅 `proto/.harness/contract-rules.md`
- **处理跨模块集成与联调** ➔ 查阅 `.harness\integration-progress.md` 和 `.harness\integration_feature_list.json`
- **开发前端页面** ➔ 查阅 `client/AGENTS.md`
- **开发后端业务** ➔ 查阅 `server/AGENTS.md`


## 常用命令 (Commands)

```sh
# 注意：以下指令当前为演示占位，项目当前不可用
# 全局指令 (Global)
make setup              # 初始化本地开发环境与依赖
make proto              # 编译 protobuf 并生成前后端桩代码（修改接口后必执行）
make check              # 执行全局静态类型检查与 Lint 扫描
make build              # 构建后端 API Server 与 Migrate 迁移脚本
make up                 # 启动全栈本地开发环境（包含数据库与中间件）

# 后端 (Backend - Go)
cd server
go test ./...           # 运行后端所有单元测试
go run cmd/ego/main.go  # 本地启动后端服务

# 前端 (Frontend - Flutter)
cd client
flutter pub get         # 获取前端依赖
flutter test            # 运行前端测试
flutter run             # 本地运行前端应用
```