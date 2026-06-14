---
name: ego-feature
description: Use when implementing a new feature or making changes to the ego app. Dynamic context discovery from ego-* skills, DDD-aware design and planning, full-stack implementation with testing, and skill context update after completion.
---

# ego-feature

全栈功能开发工作流。输入 `/ego-feature <需求描述>` 自动执行 8 个阶段。

## 架构速查

```
proto/ego/api.proto          ← 唯一事实来源
client/lib/features/<page>/  ← Flutter Riverpod + GoRouter
server/internal/<domain>/    ← Go DDD: domain/app/adapter
server/internal/platform/    ← 基础设施: postgres, ai, auth, logging
server/internal/bootstrap/   ← 依赖注入, composite handler
```

**Proto 是前后端契约的唯一来源。** 接口变更必须从 proto 开始，生成两端桩代码：
- `make proto-go` → `server/proto/ego/`
- `make proto-dart` → `client/lib/data/generated/`

---

## Phase 0: 上下文发现

### 步骤

1. **解析需求中的关键词**，匹配对应的 ego-* skill：

| 关键词 | Client Page Skill | Server Domain Skill | 涉及文件 |
|--------|-------------------|---------------------|----------|
| 登录/注册/auth/鉴权 | `ego-login` | identity | `client/lib/features/login/`, `server/internal/identity/` |
| 首页/此刻/写/echo/moment | `ego-now` | writing | `client/lib/features/now/`, `server/internal/writing/` |
| 过往/history/past/trace | `ego-past` | timeline | `client/lib/features/past/`, `server/internal/timeline/` |
| 星图/starmap/星座/对话/chat | `ego-starmap` | starmap + conversation | `client/lib/features/starmap/`, `server/internal/starmap/` + `conversation/` |
| 引导/onboarding | `ego-onboarding` | - | `client/lib/features/onboarding/` |
| 设置/账号信息/登出/profile | `ego-setting` | setting | `client/lib/features/setting/`, `server/internal/setting/` |
| 基础设施/DB/AI/auth/配置 | `ego-platform` | platform + config + bootstrap | `server/internal/{platform,config,bootstrap}/` |

2. **读取对应 skill 的 `SKILL.md`**，然后引导 agent Read `client.md` 和 `server.md`

3. **向用户汇报**发现的 context 覆盖范围，确认是否遗漏

### 如果需求跨多个领域

按优先级分别加载上下文，每个领域单独过 Phase 1-2 的需求澄清。

---

## Phase 1: 需求澄清

**REQUIRED SUB-SKILL:** Use superpowers:brainstorming

### 要点

- 如果涉及前后端数据交互，优先讨论 proto 契约变更
- 遵循 ego 的 auth 模型：Login / CheckPhone / SendVerificationCode / Register 免 JWT 认证，其余 RPC 需 token
- 短信/邮件等外部服务通过 adapter 层抽象

---

## Phase 2: 设计定稿

### 设计文档必须覆盖

1. **Proto 变更** — 新/改 rpc + message 定义
2. **后端 DDD 变更** — 按三层列出：
   - `domain/`: 实体、Repository 接口、领域错误
   - `app/`: 用例逻辑、应用层接口
   - `adapter/`: gRPC handler、postgres repo、外部服务 adapter
3. **前端变更** — page、widget、provider、router、ego_client
4. **数据库** — migration SQL（如有）
5. **错误映射** — 领域错误 → gRPC status → 前端提示

### 设计逐节确认

每节展示后等待用户确认，不要一次性展示全部。

### 功能拆分评估

**设计定稿后，必须与用户一同评估是否要拆分 feature。** 改动越多，影响面越广，bug 更多。

如需拆分，给出拆分后的多条独立指令，格式：

```
建议拆分为 N 个子 feature，按顺序执行：

1. /ego-feature <子需求1>
2. /ego-feature <子需求2>
...
```

提示用户 clear 会话后重新使用 `/ego-feature <指令>` 逐个执行。

**用户确认拆分方案或决定不拆分后，才进入 Phase 3。**

---

## Phase 3: 实现计划

**REQUIRED SUB-SKILL:** Use superpowers:writing-plans

### ego 特殊约束

- **Task 顺序必须是**: proto → migration → domain → app → adapter → handler → wiring → client
- **Proto 生成**：
  - Go: `make proto-go`（产物在 `server/proto/ego/`）
  - Dart: `make proto-dart`（产物在 `client/lib/data/generated/`，使用 `--dart_out=grpc:` 生成 gRPC client）
- **sqlc**: `make sqlc` 重新生成数据访问代码
- **DDD 模块文件**：
  - `module.go` 是依赖注入入口，接收 `Deps` 返回 `Handler`
  - `bootstrap/<domain>.go` 创建 handler 并注入 `Platform` 资源
  - `composite.go` 按 RPC 方法分发到对应 handler
- **测试**：每个 domain 的 `adapter/grpc/` 下有单元测试（mock）和集成测试（真实 DB）
- **每个 Task 都是 2-5 分钟的小步骤，精确到文件路径和代码**

### Plan 结构

```markdown
### Task N: [Component Name]
**Files:** Create/Modify/Test: exact paths
- [ ] Step 1: ...
- [ ] Step N: Commit
```

---

## Phase 4: 编码实现

按 plan task 顺序执行：

- 开始前先创建新分支（如 `feat/xxx`、`fix/xxx`）
- 优先 subagent-driven（独立 task 并行）
- task 间 review
- proto 变更后必须重新 `make proto-go proto-dart`

### 提交策略

**一次 ego-feature 只有一次 commit。** 中间 task 的变更会导致代码无法编译或测试失败，产生 broken commit。

- 所有代码变更累积在 working tree 中
- **commit 时机**：Phase 4 编码 + Phase 5 全部检查通过 + 真机测试通过 + Phase 6 skill 回写完成 → 以上全部完成后才执行 **一次** `git commit`
- 特殊情况（如 proto 生成、sqlc 生成等纯无副作用步骤）可在确认生成正确后单独 commit

### 提交前必检清单

**所有变更 commit 前必须依次通过以下检查，任何一项未通过都不得提交：**

1. **Go 测试**: `go test ./internal/<domain>/... -v -count=1`（agent 执行）
2. **Go 静态检查**: `go vet ./internal/<domain>/...`（agent 执行）
3. **Flutter 静态分析**: `cd client && flutter analyze`（agent 执行，必须零 issue）
4. **Smoke 测试**: `bash smoke.sh`（agent 执行，端到端 grpcurl 测试；默认静默模式仅显示 pass/fail，出错时告知用户加 `--verbose` 查看完整响应并调整 `LOG_LEVEL=debug` 排查）
5. **真机测试**: 运行 `bash clean-start.sh`，按手动测试清单逐项验证（用户执行）
6. **sqlc 副作用检查**: `make sqlc` 后检查 `git diff --stat`，如果 `server/internal/platform/postgres/sqlc/` 下出现 features 无关的变更，需 `git checkout` 还原（agent 执行）

### 核心改动逻辑简述

**全部 task 编码完成后，必须向用户简述本次 feature 的核心改动逻辑**，包括：

1. **数据流**：请求/数据从头到尾经过哪些关键节点（如 interceptor → handler → repo → DB）
2. **关键设计决策**：为什么这样设计（如拦截器顺序、key 格式、error 映射策略）
3. **影响面**：改动了哪些文件、影响了哪些现有功能

此简述放在提测之前，帮助用户在真机测试时理解预期行为。

### 真机测试硬阻断规则

**Agent 严禁在用户确认真机测试通过前执行 commit。** 此规则无例外：

- 前 5 项自动化检查（Go test/vet、Flutter analyze、smoke.sh、sqlc）全部通过后，agent **必须停止并等待**
- Agent 输出手动测试清单，明确告知用户：「请在真机上完成测试后通知我 commit」
- **唯有用户明确通知测试通过后**，agent 才能执行 `git commit`
- 即使用户说「可以提交了」，agent 也应二次确认：「真机测试已通过？」
- 如果当前环境无可用设备：agent 禁止自行 commit，必须告知用户需自行在设备上测试并通知结果

---

## Phase 5: 提测验证

**全部 task 完成后的提测流程。必须先通过所有检查，再执行 commit。**

### 5a. 静态检查（必须）

```bash
# Go 单元测试
go test ./internal/<domain>/adapter/grpc/ -run '^Test[^I]' -v

# Go 静态检查
go vet ./internal/<domain>/...

# Flutter 静态分析
cd client && flutter analyze
```

### 5b. Smoke 端到端测试（必须）

```bash
# 从零启动 PostgreSQL + 迁移 + 编译 + 启动服务 + grpcurl 测试全部 RPC
# 默认静默模式：仅显示 pass/fail
bash smoke.sh

# 调试模式：显示完整 RPC 响应 + 服务端日志
VERBOSE=1 bash smoke.sh
```

**如果 smoke 失败**，agent 应告知用户：
1. 使用 `bash smoke.sh --verbose` 查看完整 RPC 响应定位失败断言
2. 修改 smoke.sh 中 `LOG_LEVEL` 为 `debug` 可查看服务端日志排查 backend 错误
3. 使用 `--keep-db` 保留数据库（含 seed 数据）避免每次重建

**新增 RPC 必须在 smoke.sh 中添加对应的测试断言**：
- 正常调用：带 token 验证返回数据正确
- 鉴权验证：不带 token 验证返回 UNAUTHENTICATED

### 5c. 真机测试（如有设备）

```bash
# 连接设备后运行
cd client && flutter run

# 或指定设备
flutter run -d <device_id>
```

### 5d. 手动测试清单

生成 checklist 覆盖：

| # | 场景 | 步骤 | 预期 |
|---|------|------|------|
| 1 | 正常流程 | ... | ... |
| 2-N | 错误/边界 | ... | ... |

---

## Phase 6: Context 回写

**实现完成后，更新受影响的 ego-* skill 文件：**

1. 确认哪些 ego-* skill 被本次变更影响
2. 读取对应的 `client.md` / `server.md`
3. 更新：路由、新 RPC、新文件、改动后的数据流、架构变更
4. 如果 `SKILL.md` 中的快速文件索引有变化，一并更新

**skill 回写与实现代码一起提交，不单独 commit。** 

---

## Phase 7: Push & PR

**全部 commit 完成后，推送到远程并创建 Pull Request。**

### 步骤

1. **推送分支**：

```bash
git push origin <current-branch>
```

2. **创建 PR**：

```bash
gh pr create \
  --base test \
  --head <current-branch> \
  --title "<PR title>" \
  --body "<PR body>"
```

### 约定

- **Base branch**: `test`（开发集成分支），最终合入 `main`
- **PR body** 末尾附带 `🤖 Generated with [Claude Code](https://claude.com/claude-code)`
- 如果 `gh` CLI 未认证，提示用户自行创建 PR

---

## 文件路径速查

```
proto/ego/api.proto
server/cmd/ego/main.go
server/internal/
  bootstrap/composite.go
  bootstrap/platform.go
  bootstrap/<domain>.go
  config/config.go
  <domain>/
    module.go
    domain/{types,ports,errors}.go
    app/{<usecase>,ports}.go
    adapter/grpc/{handler,mapper}.go
    adapter/postgres/<repo>.go
    adapter/sms/
  platform/
    postgres/{postgres,queries,migrations,sqlc}/
    ai/client.go
    auth/{bcrypt,jwt,jwt_issuer,interceptor}.go
    logging/
client/lib/
  main.dart / app.dart
  core/router/router.dart
  core/providers/
  core/theme/
  core/version.dart
  data/services/ego_client.dart
  data/generated/ (proto 生成，勿手动编辑)
  features/<page>/
    <page>_page.dart
    providers/<page>_provider.dart
    widgets/
```
