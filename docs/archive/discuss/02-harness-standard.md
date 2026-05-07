# Harness 仓库规范总结

Harness 的目标是把仓库本身变成 agent 可以稳定工作的系统。

它不只是“给 AI 写一段提示词”，而是通过一组固定文件，让 agent 能够知道：

- 从哪里开始。
- 项目怎么初始化。
- 当前功能做到哪里。
- 哪些规则不能违反。
- 怎样验证一个功能完成。
- 会话结束时怎样留下可接手的状态。

---

## 1. Harness 核心文件结构

推荐结构：

```text
project/
├── AGENTS.md / CLAUDE.md
├── init.sh
├── Makefile
├── feature_list.json
├── claude-progress.md / PROGRESS.md
├── DECISIONS.md
├── session-handoff.md
├── clean-state-checklist.md
├── evaluator-rubric.md
├── quality-document.md
├── docs/
│   ├── api-patterns.md
│   ├── database-rules.md
│   └── testing-standards.md
└── src/
    ├── api/
    │   └── ARCHITECTURE.md
    └── db/
        └── CONSTRAINTS.md
```

最小可用结构：

```text
project/
├── AGENTS.md / CLAUDE.md
├── init.sh
├── feature_list.json
└── claude-progress.md / PROGRESS.md
```

这四个文件分别解决：

- agent 进入仓库后先读什么。
- 项目如何初始化和验证。
- 功能清单与完成状态如何记录。
- 跨会话进度如何交接。

---

## 2. 核心文件作用

### 2.1 AGENTS.md / CLAUDE.md

根指令文件，是 agent 每次进入仓库后的第一入口。

它应该包含：

- 项目一句话说明。
- 快速开始命令。
- 必须遵守的硬约束。
- 常用验证命令。
- 当前最重要的文档入口。
- 专题文档的索引。

它不应该包含：

- 所有架构细节。
- 所有 API 规则。
- 所有数据库规范。
- 长篇背景材料。
- 过期的手工操作记录。

更好的做法是让它成为“路由文件”：

```text
AGENTS.md
├── 想改 API？读 docs/api-patterns.md
├── 想改数据库？读 docs/database-rules.md
├── 想跑测试？读 docs/testing-standards.md
└── 想看当前进度？读 PROGRESS.md
```

根指令文件要短、稳定、可扫描。它负责告诉 agent 去哪里找信息，而不是把所有信息塞进去。

### 2.2 init.sh

初始化脚本，负责把仓库从“刚拉下来”带到“可以工作”的状态。

它通常应该做：

- 检查运行环境。
- 安装依赖。
- 生成必要代码。
- 执行最基础的健康检查。
- 打印下一步常用命令。

它的价值是减少 agent 猜测。

没有 `init.sh` 时，agent 往往需要自己判断：

- 用 npm、pnpm、yarn 还是 go test？
- 需要先跑 migration 吗？
- proto 是否需要生成？
- 测试前是否要启动服务？

有了 `init.sh`，仓库明确告诉 agent：先跑这一条。

### 2.3 Makefile

标准命令入口。

推荐至少提供：

```text
make setup
make test
make lint
make check
make dev
make build
```

`Makefile` 的作用不是炫技，而是统一命令语言。

当 agent 需要验证时，不应该每次重新探索项目脚本，而是优先执行：

```text
make check
```

如果项目不适合 Makefile，也可以用等价脚本，但仓库必须提供稳定入口。

### 2.4 feature_list.json

功能清单，是 Harness 的核心调度文件。

它描述：

- 仓库有哪些功能。
- 每个功能当前状态是什么。
- 每个功能怎么算完成。
- 需要执行什么验证。
- 完成证据是什么。

推荐字段：

```json
{
  "id": "auth.login",
  "title": "用户可以登录或自动注册",
  "status": "passing",
  "description": "用户输入 account/password 后，系统登录或自动创建账号。",
  "acceptance": [
    "账号不存在时自动注册并返回 token",
    "账号存在时校验密码并返回 token",
    "密码错误时返回 unauthenticated"
  ],
  "verification": [
    "go test ./...",
    "手动调用 Login RPC"
  ],
  "evidence": [
    "server/internal/login/handler_test.go 覆盖登录场景"
  ],
  "notes": "JWT 有效期为 30 天"
}
```

常见状态：

```text
not_started
in_progress
blocked
passing
```

关键规则：

- 同一时间只应有一个 `in_progress`。
- `passing` 必须有 evidence。
- `blocked` 必须写清楚 blocker 和解除条件。
- 功能清单应机器可读，不要只写成自然语言备忘录。

### 2.5 claude-progress.md / PROGRESS.md

跨会话进度文件。

它记录的是“当前仓库状态”，不是流水账。

建议包含：

- 当前总体状态。
- 最近完成了什么。
- 当前正在做什么。
- 下一个最高优先级任务。
- 已知 blocker。
- 已执行过的验证命令。
- 最近一次验证结果。
- 下个 agent 接手时应该先看哪里。

示例结构：

```text
# Progress

## Current State

## Last Verified

## Active Work

## Blockers

## Next Best Step

## Verification Log
```

它的价值是避免每次新会话都从零开始理解仓库。

### 2.6 DECISIONS.md

架构和产品决策日志。

它记录：

- 做了什么决定。
- 为什么这么决定。
- 当时有哪些备选方案。
- 放弃了什么方案。
- 这个决定带来哪些后续约束。

推荐格式：

```text
## 2026-05-07 使用模块化单体而不是微服务

Decision:
采用 Go 模块化单体，按 DDD bounded context 拆包。

Context:
当前团队和产品阶段更需要快速迭代，微服务会增加部署和调试成本。

Alternatives:
- 微服务
- 按技术层分包

Consequences:
- 领域边界通过代码包和写入归属约束保证
- 后续可按 bounded context 拆服务
```

它的作用是防止 agent 重复争论已经决定过的事情。

### 2.7 session-handoff.md

会话交接文件。

适合在长任务、复杂重构、未完全收尾时使用。

它应该回答：

- 本轮做了什么。
- 哪些文件被修改。
- 哪些验证已经通过。
- 哪些地方还没验证。
- 哪些地方不要随便改。
- 下一轮应该怎么继续。

它和 `PROGRESS.md` 的区别：

- `PROGRESS.md` 是长期状态。
- `session-handoff.md` 是本轮交接。

### 2.8 clean-state-checklist.md

干净收尾检查清单。

它定义一次 agent 会话结束前要确认的事项。

典型检查项：

```text
- 代码能构建
- 测试已运行或明确说明未运行原因
- feature_list.json 状态已更新
- PROGRESS.md 已更新
- 没有临时调试代码
- 没有无意义日志输出
- 没有未解释的大范围重构
- 下一个 agent 能从文档接手
```

它的意义是让“完成”不只等于“写了代码”，而是仓库处在可继续工作的状态。

### 2.9 evaluator-rubric.md

评估标准文件。

它用于评价一次 agent 输出是否达标。

常见评分维度：

- Correctness：功能是否正确。
- Verification：是否真实验证。
- Scope Control：是否控制范围。
- Maintainability：是否可维护。
- Safety：是否避免破坏已有功能。
- Handoff Quality：是否留下清晰交接信息。

它适合用于复盘、评审或训练 agent。

### 2.10 quality-document.md

代码库质量文档。

它不是单次任务的验收，而是整个仓库的健康度快照。

可记录：

- 当前核心功能质量。
- 测试覆盖薄弱区。
- 架构债务。
- 文档缺口。
- 运行稳定性。
- agent 可读性。

它回答的问题是：

> 这个仓库整体是否适合持续交给 agent 工作？

### 2.11 docs/*.md

专题规则文档。

例如：

```text
docs/api-patterns.md
docs/database-rules.md
docs/testing-standards.md
docs/deployment.md
```

这些文档承载细节规则。

根 `AGENTS.md` 只需要链接它们。

这样可以避免一个巨型根指令文件膨胀到难以扫描。

### 2.12 src/**/ARCHITECTURE.md

模块就近架构说明。

适合放在具体模块旁边，例如：

```text
src/api/ARCHITECTURE.md
src/billing/ARCHITECTURE.md
server/internal/starmap/ARCHITECTURE.md
```

它应该说明：

- 当前模块负责什么。
- 不负责什么。
- 对外暴露什么接口。
- 依赖哪些模块。
- 哪些规则不能破坏。

讲义中的核心思想是：知识应靠近代码。

### 2.13 src/**/CONSTRAINTS.md

模块硬约束文件。

它记录不可违反的规则。

例如数据库模块：

```text
- 所有查询必须带 user_id
- 不允许跨用户读取数据
- 不允许直接拼接 SQL
- migration 不允许修改历史文件，只能新增
```

例如 API 模块：

```text
- 所有非 Login RPC 必须鉴权
- handler 不能直接写 SQL
- proto 字段只能追加，不能复用旧 tag
```

这类约束放在模块附近，agent 修改相关代码时更容易看到。

---

## 3. 关键原则

### 3.1 仓库是唯一事实来源

agent 只能稳定依赖仓库里的信息。

如果规则只存在于聊天记录、人的记忆或口头说明里，下次会话很容易丢失。

因此这些信息应该进入仓库：

- 项目如何启动。
- 如何测试。
- 当前功能进度。
- 架构决策。
- 已知问题。
- 模块约束。
- 完成标准。

### 3.2 根指令文件应该短而清晰

`AGENTS.md` / `CLAUDE.md` 不应该变成百科全书。

它的职责是：

- 给方向。
- 给入口。
- 给硬约束。

细节应该拆到专题文档和模块文档。

### 3.3 功能清单是 Harness 的脊梁

`feature_list.json` 不只是 todo list。

它是：

- 任务调度依据。
- 验收依据。
- 进度依据。
- 交接依据。
- 自动化评估依据。

没有功能清单，agent 容易只完成“看起来该做的事”，而不是产品真正需要的事。

### 3.4 完成必须有证据

`passing` 不能只靠自然语言声明。

它应该有证据，例如：

- 测试命令。
- 测试文件。
- 构建结果。
- 手动验证步骤。
- 截图。
- 日志。
- commit。

证据越具体，下一个 agent 越容易相信并继续推进。

### 3.5 初始化必须独立成阶段

agent 在写代码前需要先知道仓库是否可运行。

`init.sh` 或 `make setup` 的价值是把“准备环境”变成显式阶段。

否则 agent 可能在一个本来就无法启动的仓库里误判问题。

### 3.6 验证要覆盖真实路径

单元测试很重要，但不等于完整完成。

Harness 应鼓励分层验证：

- 静态检查。
- 单元测试。
- 集成测试。
- 端到端测试。
- 手动验证步骤。

尤其是用户可见功能，最好有端到端验证或明确的人工验证路径。

### 3.7 每次会话都要留下干净状态

好的 agent 会话结束后，下一个 agent 应该能快速接手。

这要求：

- 状态已记录。
- 测试结果已记录。
- 未完成事项已记录。
- 临时修改已清理。
- 不确定点已说明。

### 3.8 知识应该靠近代码

全局规则放根目录。

模块规则放模块旁边。

例如：

- 全局开发流程放 `AGENTS.md`。
- API 规范放 `docs/api-patterns.md`。
- Starmap 架构放 `server/internal/starmap/ARCHITECTURE.md`。
- 数据库硬约束放 `server/internal/db/CONSTRAINTS.md`。

这样 agent 修改某块代码时，更容易读到对应规则。

### 3.9 状态要机器可读

自然语言文档适合解释。

机器可读文件适合调度和检查。

因此：

- 功能状态放 `feature_list.json`。
- 长期说明放 Markdown。
- 验证命令放 Makefile 或脚本。

### 3.10 Harness 是演进出来的

不需要一开始就创建所有文件。

推荐顺序：

1. `AGENTS.md`
2. `init.sh`
3. `feature_list.json`
4. `PROGRESS.md`
5. `DECISIONS.md`
6. `clean-state-checklist.md`
7. 专题 docs
8. 模块级 `ARCHITECTURE.md` / `CONSTRAINTS.md`

先保证 agent 能正确开始、正确验证、正确交接，再逐步补充更细的规范。

