---
name: ego-feature
description: Use when implementing a new feature or making changes to the ego app. Dynamic context discovery from ego-* skills, DDD-aware design and planning, full-stack implementation with testing, and skill context update after completion.
---

# ego-feature

全栈功能开发工作流。输入 `/ego-feature <需求描述>` 自动执行 7 个阶段。

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

- 优先 subagent-driven（独立 task 并行）
- task 间 review
- proto 变更后必须重新 `make proto-go proto-dart`
- 在 feat/<feature-name> 分支上工作

### 提交策略

**严禁每个 task commit 一次。** 中间 task 的变更会导致代码无法编译或测试失败，产生 broken commit。

- 所有代码变更累积在 working tree 中
- 仅在 **全部 task 完成 + 测试通过** 后进行 **一次** commit
- 特殊情况（如 proto 生成、sqlc 生成等纯无副作用步骤）可在确认生成正确后单独 commit

---

## Phase 5: 测试

### 5a. Go 测试

```bash
# 单元测试（mock）
go test ./internal/<domain>/adapter/grpc/ -run '^Test[^I]' -v

# 集成测试（需 postgres 运行）
go test ./internal/<domain>/adapter/grpc/ -run '^TestIntegration' -v
```

### 5b. SMS/外部服务测试（如涉及）

```bash
TEST_PHONE_NUMBER=138xxxxxxx go test ./internal/<domain>/adapter/sms/ -run TestSend -v
```

### 5c. Flutter 静态检查

```bash
cd client && flutter analyze
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
5. 提交 skill 更新

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
  data/services/ego_client.dart
  data/generated/ (proto 生成，勿手动编辑)
  features/<page>/
    <page>_page.dart
    providers/<page>_provider.dart
    widgets/
```
