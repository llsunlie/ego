# ego 后端 DDD 架构设计

> 语言：Go  
> 形态：单体后端优先，按 DDD 模块化组织  
> 范围：架构设计文档，不包含代码实现

本文基于 `overview/intro/input/ego4.0.txt`、`overview/intro/flows.md`、当前 `proto/ego/api.proto`、数据库草案和后端草案整理。

ego 的核心不是“记日记”，而是：

> 用户写下此刻，系统从过去的话里找回声，进一步沉淀为星图，并允许用户和过去的自己对话。

所以后端架构不应该只围绕表或接口拆分，而应该围绕这些稳定的业务能力拆分。

---

## 1. 先用一句话理解 DDD

DDD，即 Domain-Driven Design，领域驱动设计。

它关注的不是“目录一定要叫 domain/application/infrastructure”，而是：

1. 先理解业务里有哪些核心概念。
2. 给这些概念划清边界。
3. 让代码结构反映业务语言和业务规则。
4. 不让数据库、gRPC、AI SDK 这些技术细节反过来污染业务模型。

在 ego 里，真正重要的业务词不是 `handler`、`service`、`db`，而是：

- 用户
- 此刻
- 原话
- 回声
- 我发现
- Trace，一次完整思考痕迹
- 收进星图
- 星星
- 星座
- 过去的自己
- 话题引子
- 对话

DDD 的目标，就是让后端代码也围绕这些词组织。

---

## 2. DDD 常用概念，用 ego 解释

### 2.1 领域 Domain

领域就是业务问题本身。

ego 的大领域是“个人表达与自我回声系统”。它要解决的是：

- 如何保存用户当下的表达。
- 如何从过去的话里找相似回声。
- 如何把多轮思考沉淀成一颗星。
- 如何把相关的星连成星座。
- 如何让用户和过去的自己对话。

### 2.2 子领域 Subdomain

一个产品里通常有多个子领域。

ego 可以分成：

- 核心子领域：写下此刻与回声、星图形成、过去自己对话。
- 支撑子领域：登录认证、AI 能力封装、时间线查询。

核心子领域是产品差异化所在，要花最多设计精力。

### 2.3 限界上下文 Bounded Context

限界上下文是 DDD 最重要的概念之一。

它表示：某些业务词只在一个边界里有明确含义。

例如：

- 在“此刻写作”上下文里，`Moment` 是用户刚写下的一段表达。
- 在“星图”上下文里，`Star` 是一个被寄存的 Trace，不是单条 Moment。
- 在“对话”上下文里，`PastSelfCard` 是构建过去自己人格上下文的入口。

这些概念可以互相引用 ID，但不要互相越界修改内部状态。

### 2.4 聚合 Aggregate

聚合是一组必须保持一致的领域对象。

例如：

- 一次 Trace 下可以有多条 Moment。
- 一个 ChatSession 下有多条 ChatMessage。
- 一个 Constellation 下有若干 StarRef、Insight、PastSelfCard、TopicPrompt。

聚合的意义是：事务边界通常围绕聚合来定。

### 2.5 实体 Entity

实体有唯一身份，即使内容变化，它还是同一个东西。

ego 里的实体：

- User
- Moment
- Trace
- Star
- Constellation
- ChatSession
- ChatMessage

### 2.6 值对象 Value Object

值对象没有独立身份，只用值表达含义。

ego 里的值对象：

- TraceID
- MomentID
- UserID
- NormalizedPosition：星图坐标 0-1
- Echo：一次实时匹配结果
- SimilarityScore：相似度
- ReferencedMoment：AI 回复引用来源
- TopicQuestion：话题引子的问题文本

### 2.7 领域服务 Domain Service

当一条业务规则不自然属于某个实体时，可以放到领域服务里。

ego 里的领域服务：

- EchoMatcher：根据当前 Moment 找历史回声。
- TraceStashPolicy：判断一个 Trace 是否可以被收进星图。
- ConstellationClusterer：判断新 Star 应该成为孤星、加入正在成型星座，还是形成星座。
- PastSelfPersonaBuilder：根据原话构建“过去的自己”的约束。

### 2.8 应用服务 Application Service

应用服务负责完成一个用户用例，编排领域对象、仓储、外部服务。

例如：

- CreateMomentUseCase
- GenerateInsightUseCase
- StashTraceUseCase
- StartChatUseCase
- SendMessageUseCase

应用服务可以调数据库仓储、AI Port、事务管理，但不应该塞满业务判断。业务判断尽量回到领域对象或领域服务。

---

## 3. ego 的领域划分依据

领域不是按页面机械拆，也不是按数据库表拆。

ego 的划分依据是五件事：

### 3.1 按业务能力拆

每个领域回答一个稳定的业务问题：

- Identity：谁在使用系统？
- Writing：用户刚写下的话如何成为一个 Moment，并产生回声？
- Timeline：用户如何回看过去所有表达？
- Starmap：一个 Trace 如何被寄存、成为星星、聚成星座？
- Conversation：用户如何和某段过去的自己对话？

### 3.2 按生命周期拆

不同对象的生命周期不同：

- Moment 一旦写下，基本不可变。
- Trace 可以随着“顺着再想想”继续追加 Moment。
- Star 在“收进星图”后出现。
- Constellation 会随着更多 Star 加入而变化。
- ChatSession 会随着用户多轮发送消息而增长。

生命周期不同，通常应该属于不同边界。

### 3.3 按一致性要求拆

哪些状态必须一次事务内保证？

- 创建 Moment：必须保证 Moment 属于当前用户、属于某个 Trace。
- 发送聊天消息：用户消息和 AI 回复的持久化要有清晰顺序。
- 收进星图：创建 Star 必须成功；聚类和 AI 生成星座资产可以异步。

不需要强一致的部分，不要放进同一个大事务。

### 3.4 按变化原因拆

以后最可能变化的点：

- 回声匹配算法可能变。
- 星座聚类算法可能变。
- AI prompt 和模型供应商可能变。
- 页面展示方式可能变。
- 登录方式可能从 account/password 变成手机号或 OAuth。

变化原因不同，就应该隔离。

### 3.5 按团队语言拆

产品和设计在说“此刻、过往、星图、过去的自己”。

后端代码里也应该尽量保留这些语言，而不是所有东西都叫 `Record`、`Item`、`Data`。

---

## 4. 推荐的限界上下文

### 总览

| 上下文 | 类型 | 负责什么 | 不负责什么 |
| --- | --- | --- | --- |
| Identity | 支撑域 | 登录、自动注册、JWT 身份 | 用户画像、业务数据 |
| Writing | 核心域 | 写下此刻、Trace、Moment、回声、单次观察 | 星座聚类、时间线展示、聊天 |
| Timeline | 查询域 | 过往列表、记忆光点盲盒 | 创建或修改 Moment |
| Starmap | 核心域 | 收进星图、Star、Constellation、星座详情资产 | Moment 创建、聊天消息 |
| Conversation | 核心域 | 和过去自己对话、会话历史、引用标注 | 星座聚类、Moment 写入 |
| Intelligence | 支撑能力 | AI/Embedding/Prompt 外部能力适配 | 持有业务状态 |

说明：

- `Intelligence` 更像一个支撑能力或反腐层，不一定是独立业务上下文。
- 早期可以是 `internal/platform/ai`，由各领域通过 Port 调用。
- 不建议让业务代码直接调用 OpenAI、Gemini 或其他 SDK。

---

## 5. 上下文一：Identity

### 5.1 业务职责

Identity 只回答一个问题：

> 当前请求是谁发出的？

它负责：

- account/password 登录。
- account 不存在时自动注册。
- bcrypt 密码校验。
- JWT 签发。
- gRPC interceptor 解析 JWT，并把 UserID 注入上下文。

### 5.2 领域对象

聚合：

- User

实体：

- User

值对象：

- UserID
- Account
- PasswordHash

### 5.3 边界

Identity 可以写：

- `users`

Identity 不应该写：

- `moments`
- `stars`
- `constellations`
- `chat_sessions`

其他上下文只能拿到 `UserID`，不应该知道密码、密码 hash 或登录细节。

### 5.4 对应接口

- `Login`

### 5.5 案例

用户第一次输入账号密码：

1. LoginUseCase 查询 account。
2. 不存在则创建 User。
3. 生成 PasswordHash。
4. 签发 JWT。
5. 返回 `created=true`。

用户第二次登录：

1. LoginUseCase 查询 account。
2. 校验 bcrypt。
3. 签发 JWT。
4. 返回 `created=false`。

---

## 6. 上下文二：Writing

### 6.1 业务职责

Writing 是 ego 最核心的领域之一。

它回答：

> 用户写下此刻后，如何保存这句话，并让它和过去产生回声？

它负责：

- 创建 Moment。
- 创建或延续 Trace。
- 生成 embedding。
- 匹配历史 Echo。
- 生成当前这次体验里的“我发现”。
- 保证 Moment 只属于当前用户。

### 6.2 领域对象

推荐聚合：

- Trace

实体：

- Trace
- Moment

值对象：

- TraceID
- MomentID
- MomentContent
- Embedding
- Echo
- EchoCandidate
- SimilarityScore
- MomentInsight
- TopicContext

领域服务：

- EchoMatcher
- MomentInsightGeneratorPort

### 6.3 关于 Trace 是否需要单独表

当前数据库草案用 `moments.trace_id` 表示 Trace，没有单独 `traces` 表。

从 DDD 角度，Trace 是一个真实领域概念：

- 它表示一次完整思考痕迹。
- 它可以包含多轮 Moment。
- 它是“收进星图”的最小业务单位。

因此推荐后续显式增加 `traces` 表：

```text
traces
- id
- user_id
- status: active / stashed / abandoned
- created_at
- updated_at
```

MVP 阶段可以继续使用 `moments.trace_id` 作为隐式 Trace，但领域代码里仍然应该把它建模为 `Trace`，不要把它只当成一个普通字段。

### 6.4 边界

Writing 可以写：

- `moments`
- 推荐未来写 `traces`

Writing 可以读：

- 当前用户自己的历史 `moments`，用于 Echo 匹配。

Writing 不应该写：

- `stars`
- `constellations`
- `chat_sessions`
- `chat_messages`

“收进星图”看起来是在此刻页触发，但业务上属于 Starmap，而不是 Writing。

### 6.5 对应接口

- `CreateMoment`
- `GenerateInsight`

### 6.6 案例：第一次写下此刻

用户输入：“今天和同事起了冲突。”

Writing 的用例流程：

1. 校验用户已登录。
2. 如果请求没有 `trace_id`，创建新的 TraceID。
3. 创建 Moment。
4. 调用 EmbeddingPort 得到向量。
5. 保存 Moment。
6. EchoMatcher 在当前用户历史 Moment 里找相似内容。
7. 如果没有历史，返回 `echo=nil`。
8. 如果有历史，返回 Echo target 和 candidates。

注意：

- Echo 是实时结果，不一定要入库。
- 当前这次的 “我发现” 也可以是实时结果，不一定入库。
- Moment 保存成功不应该依赖 AI insight 必须成功。

### 6.7 案例：顺着再想想

用户点击“顺着再想想”，继续输入：“其实是我害怕被否定。”

Writing 的关键规则：

1. 新 Moment 继续使用同一个 TraceID。
2. Echo 匹配应排除同一个 Trace 内的 Moment，避免自己回自己。
3. 新 Moment 和新 Echo 形成 Trace 的下一段。

这说明 Trace 是业务聚合，而不只是页面状态。

---

## 7. 上下文三：Timeline

### 7.1 业务职责

Timeline 回答：

> 用户如何回看自己过去说过的话？

它负责：

- 过往时间线。
- 按游标分页读取 Moment。
- 记忆光点盲盒随机读取 3 条 Moment。
- 返回 `connected` 状态供前端显示 `✦ 已联结`。

### 7.2 性质

Timeline 更像查询上下文，不一定需要复杂领域模型。

它不拥有 Moment 的创建权，只负责读取。

### 7.3 边界

Timeline 可以读：

- `moments`

Timeline 不应该写：

- `moments`
- `stars`
- `constellations`

### 7.4 对应接口

- `ListMoments`
- `GetRandomMoments`

### 7.5 案例：过往页加载

用户切到“过往”：

1. TimelineQueryService 调用 MomentReadRepository。
2. 按 `created_at desc` 返回扁平列表。
3. 前端按月分组。
4. 每条带 `connected`。

为什么不把这个放到 Writing？

因为 Writing 的核心是“创建和回声”，Timeline 的核心是“浏览和读取”。它们用到同一张表，但变化原因不同。

---

## 8. 上下文四：Starmap

### 8.1 业务职责

Starmap 是 ego 另一个核心领域。

它回答：

> 用户的一次完整 Trace，如何被寄存为星星，并逐渐形成星座？

它负责：

- 收进星图。
- 创建 Star。
- 根据 Trace/Moment 的语义进行聚类。
- 维护 Constellation 状态：lone / forming / formed。
- 生成星座名称。
- 生成星座级 Insight。
- 生成 PastSelfCard。
- 生成 TopicPrompt。
- 组装星图列表和星座详情。

### 8.2 领域对象

聚合：

- Star
- Constellation

实体：

- Star
- Constellation
- ConstellationInsight
- PastSelfCard
- TopicPrompt

值对象：

- StarID
- ConstellationID
- TraceRef
- StarPosition
- VisualRhythm
- ConstellationStatus
- StarRef

领域服务：

- TraceStashPolicy
- ConstellationClusterer
- ConstellationAssetGeneratorPort

### 8.3 边界

Starmap 可以写：

- `stars`
- `constellations`
- `constellation_stars`
- `insights`
- `past_self_cards`
- `topic_prompts`

Starmap 可以读：

- Writing 提供的 Trace/Moment 只读视图。

Starmap 不应该直接创建 Moment。

如果需要判断 Trace 是否存在、是否属于当前用户，应该通过 Writing 暴露的 `TraceReader` 或 `MomentReader` Port 完成，而不是绕过领域边界随意写 Writing 的表。

### 8.4 对应接口

- `StashTrace`
- `ListConstellations`
- `GetConstellation`

### 8.5 一致性策略

`StashTrace` 应该拆成两段：

1. 强一致部分：验证 Trace，创建 Star，返回成功。
2. 最终一致部分：聚类、更新星座状态、生成星座资产。

原因：

- AI 聚类和资产生成可能慢、可能失败。
- 用户点击“收进星图”时，最重要的是仪式感和 Star 被保存。
- 星座详情可以稍后刷新出来。

MVP 可使用 goroutine。

生产级建议使用：

- Outbox 表
- 后台 worker
- 重试机制
- 任务状态字段

### 8.6 案例：收进星图

用户点击“✦ 收进星图”。

Starmap 用例流程：

1. StashTraceUseCase 接收 `trace_id`、坐标 x/y。
2. 通过 TraceReader 校验 Trace 属于当前用户。
3. 检查该 Trace 是否已经被寄存过。
4. 创建 Star。
5. 发布 `StarCreated` 或 `TraceStashed` 领域事件。
6. 返回 Star 给前端，前端播放飞星动效。
7. 后台 Clusterer 异步处理：
   - 根据 Trace 的 Moment 向量计算 Star 表征。
   - 与已有 Constellation 比较。
   - 无匹配：成为 lone。
   - 有弱匹配：成为 forming。
   - 达到阈值：成为 formed。
   - 生成或刷新 Insight、PastSelfCard、TopicPrompt。

### 8.7 案例：星座详情

用户点击星座。

GetConstellationUseCase：

1. 校验 Constellation 属于当前用户。
2. 读取星座基本信息。
3. 读取星座内所有 Star。
4. 通过 Trace/Moment 只读端口拿原话列表。
5. 读取星座级 Insight。
6. 读取 PastSelfCard。
7. 读取 TopicPrompt。
8. 组装详情 DTO。

这里的 DTO 是展示模型，不等于领域模型。

---

## 9. 上下文五：Conversation

### 9.1 业务职责

Conversation 回答：

> 用户如何和某段过去的自己说话？

它负责：

- 启动 ChatSession。
- 恢复 ChatSession。
- 保存用户消息。
- 调用 AI 生成“过去的自己”的第一人称回复。
- 保存 AI 回复和引用来源。
- 保证 AI 回复不越界、不伪造过去没想过的东西。

### 9.2 领域对象

聚合：

- ChatSession

实体：

- ChatSession
- ChatMessage

值对象：

- ChatSessionID
- MessageRole：user / past_self
- PastSelfContext
- ReferencedMoment
- CitationLabel，例如 “参考了 2月11日 前后”

领域服务：

- PastSelfPersonaBuilder
- PastSelfReplyPolicy

外部 Port：

- PastSelfResponderPort

### 9.3 边界

Conversation 可以写：

- `chat_sessions`
- `chat_messages`

Conversation 可以读：

- Starmap 提供的 PastSelfCard 只读视图。
- Writing 提供的 Moment 只读视图。

Conversation 不应该写：

- `moments`
- `stars`
- `constellations`
- `past_self_cards`

### 9.4 对应接口

- `StartChat`
- `SendMessage`

### 9.5 案例：开始对话

用户点击某张 PastSelfCard。

Conversation 用例流程：

1. 读取 PastSelfCard。
2. 读取 context_moment_ids 对应的原话。
3. 创建 ChatSession。
4. 构建 PastSelfContext。
5. 调用 PastSelfResponder 生成开场白。
6. 保存 AI 开场白为 ChatMessage。
7. 返回 chat_session_id 和 opening。

### 9.6 案例：发送消息

用户问：“那时候的我为什么一直等对方先开口？”

SendMessageUseCase：

1. 校验 ChatSession 属于当前用户。
2. 追加用户消息。
3. 加载会话历史。
4. 加载 PastSelfContext 原话。
5. 调 AI 生成回复。
6. 校验回复必须包含引用来源或明确说明“那时候没想过”。
7. 保存 AI 回复。
8. 返回 ChatMessage。

关键原则：

- AI 可以表达，但不能拥有最终领域权威。
- AI 输出必须被解析、校验、降级。
- 引用来源是业务规则，不是 UI 装饰。

---

## 10. Intelligence：AI 反腐层

AI 能力会贯穿多个上下文，但不建议把 AI SDK 到处注入。

推荐把 AI 作为外部系统，通过 Port/Adapter 使用。

### 10.1 为什么叫反腐层

外部 AI 返回的是不稳定文本或 JSON：

- 字段可能缺失。
- 格式可能错。
- 内容可能越界。
- 模型供应商可能替换。

如果业务代码直接依赖这些返回，会被污染。

反腐层负责把 AI 输出转换成稳定的领域对象。

### 10.2 推荐 Port

Writing 需要：

- `EmbeddingProvider`
- `MomentInsightGenerator`

Starmap 需要：

- `ConstellationNameGenerator`
- `ConstellationInsightGenerator`
- `PastSelfCardGenerator`
- `TopicPromptGenerator`

Conversation 需要：

- `PastSelfResponder`

### 10.3 规则

- 领域层不知道 OpenAI/Gemini/本地模型。
- 应用层只依赖接口。
- Adapter 层实现具体模型调用。
- prompt 模板和版本号应集中管理。
- AI 输出进入领域前必须校验。

---

## 11. 上下文关系图

```text
                 ┌──────────────────┐
                 │     Identity     │
                 │  User / JWT      │
                 └────────┬─────────┘
                          │ UserID
                          ▼
┌──────────────────────────────────────────────────────────┐
│                        Writing                           │
│ Trace / Moment / Echo / current Insight                  │
└──────────────┬─────────────────────────────┬─────────────┘
               │ read moments                 │ read trace
               ▼                              ▼
       ┌──────────────┐              ┌──────────────────┐
       │   Timeline   │              │     Starmap      │
       │ Past / Blind │              │ Star /           │
       │ Memories     │              │ Constellation    │
       └──────────────┘              └────────┬─────────┘
                                              │ past-self cards
                                              ▼
                                      ┌──────────────────┐
                                      │  Conversation    │
                                      │ ChatSession      │
                                      └──────────────────┘

                 ┌────────────────────────────────────┐
                 │ Intelligence / AI Anti-corruption  │
                 │ Embedding / Insight / Chat Reply   │
                 └────────────────────────────────────┘
```

---

## 12. 推荐 Go 目录结构

采用“模块化单体 + DDD + 六边形架构”。

含义：

- 还是一个 Go 服务，一个进程，部署简单。
- 代码按业务上下文分包。
- 每个上下文内部区分 domain、app、adapter。
- 技术细节放 adapter 或 platform，不进入 domain。

```text
server/
├── cmd/
│   ├── ego/
│   │   └── main.go
│   └── migrate/
│       └── main.go
│
├── internal/
│   ├── bootstrap/
│   │   ├── wire.go
│   │   └── server.go
│   │
│   ├── config/
│   │   └── config.go
│   │
│   ├── shared/
│   │   ├── domain/
│   │   │   ├── ids.go
│   │   │   ├── errors.go
│   │   │   ├── event.go
│   │   │   └── clock.go
│   │   └── app/
│   │       └── transaction.go
│   │
│   ├── platform/
│   │   ├── postgres/
│   │   │   ├── pool.go
│   │   │   ├── tx.go
│   │   │   └── migrations/
│   │   ├── grpc/
│   │   │   ├── server.go
│   │   │   ├── auth_interceptor.go
│   │   │   └── error_mapper.go
│   │   ├── auth/
│   │   │   ├── jwt.go
│   │   │   └── password.go
│   │   ├── ai/
│   │   │   ├── client.go
│   │   │   ├── embeddings.go
│   │   │   ├── prompts.go
│   │   │   └── validators.go
│   │   └── eventbus/
│   │       ├── in_memory.go
│   │       └── outbox.go
│   │
│   ├── identity/
│   │   ├── domain/
│   │   │   ├── user.go
│   │   │   └── repository.go
│   │   ├── app/
│   │   │   └── login.go
│   │   └── adapter/
│   │       ├── grpc/
│   │       │   └── handler.go
│   │       └── postgres/
│   │           └── user_repository.go
│   │
│   ├── writing/
│   │   ├── domain/
│   │   │   ├── trace.go
│   │   │   ├── moment.go
│   │   │   ├── echo.go
│   │   │   ├── insight.go
│   │   │   ├── repository.go
│   │   │   ├── ports.go
│   │   │   └── events.go
│   │   ├── app/
│   │   │   ├── create_moment.go
│   │   │   ├── generate_insight.go
│   │   │   └── queries.go
│   │   └── adapter/
│   │       ├── grpc/
│   │       │   ├── handler.go
│   │       │   └── mapper.go
│   │       └── postgres/
│   │           ├── moment_repository.go
│   │           └── echo_query.go
│   │
│   ├── timeline/
│   │   ├── app/
│   │   │   ├── list_moments.go
│   │   │   └── random_moments.go
│   │   └── adapter/
│   │       ├── grpc/
│   │       │   ├── handler.go
│   │       │   └── mapper.go
│   │       └── postgres/
│   │           └── moment_read_model.go
│   │
│   ├── starmap/
│   │   ├── domain/
│   │   │   ├── star.go
│   │   │   ├── constellation.go
│   │   │   ├── insight.go
│   │   │   ├── past_self_card.go
│   │   │   ├── topic_prompt.go
│   │   │   ├── clustering_policy.go
│   │   │   ├── repository.go
│   │   │   ├── ports.go
│   │   │   └── events.go
│   │   ├── app/
│   │   │   ├── stash_trace.go
│   │   │   ├── cluster_star.go
│   │   │   ├── list_constellations.go
│   │   │   ├── get_constellation.go
│   │   │   └── rebuild_assets.go
│   │   └── adapter/
│   │       ├── grpc/
│   │       │   ├── handler.go
│   │       │   └── mapper.go
│   │       └── postgres/
│   │           ├── star_repository.go
│   │           ├── constellation_repository.go
│   │           └── constellation_read_model.go
│   │
│   └── conversation/
│       ├── domain/
│       │   ├── chat_session.go
│       │   ├── chat_message.go
│       │   ├── past_self_context.go
│       │   ├── repository.go
│       │   └── ports.go
│       ├── app/
│       │   ├── start_chat.go
│       │   └── send_message.go
│       └── adapter/
│           ├── grpc/
│           │   ├── handler.go
│           │   └── mapper.go
│           └── postgres/
│               ├── chat_repository.go
│               └── past_self_context_query.go
│
├── proto/
│   └── ego/
│       ├── api.proto
│       ├── api.pb.go
│       └── api_grpc.pb.go
│
├── go.mod
├── go.sum
└── Makefile
```

---

## 13. 每层负责什么

### 13.1 domain

domain 是业务核心。

可以包含：

- Entity
- Value Object
- Aggregate
- Domain Service
- Repository interface
- Domain Event
- 业务错误

不应该包含：

- gRPC pb 类型
- SQL
- pgx
- sqlc
- HTTP/gRPC status code
- AI SDK
- 环境变量读取

### 13.2 app

app 是用例编排层。

可以包含：

- CreateMomentUseCase
- StashTraceUseCase
- SendMessageUseCase
- 事务编排
- 调用 Repository interface
- 调用 AI Port
- 发布领域事件

不应该包含：

- SQL 字符串
- 具体 AI SDK 调用
- gRPC 请求响应结构

### 13.3 adapter

adapter 是输入输出适配层。

可以包含：

- gRPC handler
- proto mapper
- postgres repository implementation
- read model query

### 13.4 platform

platform 是通用技术设施。

包括：

- PostgreSQL 连接
- JWT
- bcrypt
- AI SDK client
- event bus
- 日志
- 配置

platform 不表达 ego 业务规则。

---

## 14. 模型转换规则

DDD 后端里至少有三类模型：

### 14.1 Proto DTO

位置：

```text
proto/ego/*.pb.go
```

用途：

- gRPC 入参/出参。
- 面向客户端。

规则：

- 不进入 domain。
- handler 层负责转成 Command 或 Query。

### 14.2 Domain Model

位置：

```text
internal/{context}/domain
```

用途：

- 表达业务概念和业务规则。

规则：

- 不带 db tag。
- 不依赖 proto。
- 不依赖 pgx/sqlc。

### 14.3 Persistence Model

位置：

```text
internal/{context}/adapter/postgres
```

用途：

- 数据库读写。
- 可使用 sqlc 生成类型。

规则：

- 只在 adapter 内部使用。
- 进入 app/domain 前要转换。

---

## 15. 接口与领域的映射

| RPC | 所属上下文 | 应用服务 | 说明 |
| --- | --- | --- | --- |
| Login | Identity | LoginUseCase | 登录或自动注册 |
| CreateMoment | Writing | CreateMomentUseCase | 创建 Moment，返回 Echo |
| GenerateInsight | Writing | GenerateMomentInsightUseCase | 当前体验的实时观察 |
| GetRandomMoments | Timeline | GetRandomMomentsQuery | 记忆光点盲盒 |
| ListMoments | Timeline | ListMomentsQuery | 过往时间线 |
| StashTrace | Starmap | StashTraceUseCase | Trace 变 Star |
| ListConstellations | Starmap | ListConstellationsQuery | 星图总览 |
| GetConstellation | Starmap | GetConstellationQuery | 星座详情 |
| StartChat | Conversation | StartChatUseCase | 开始或恢复对话 |
| SendMessage | Conversation | SendMessageUseCase | 发送消息并得到 past-self 回复 |

---

## 16. 数据表归属

| 表 | 写入归属 | 读取归属 | 说明 |
| --- | --- | --- | --- |
| users | Identity | Identity | 账号与密码 |
| traces | Writing | Writing / Starmap | 推荐新增，MVP 可隐式 |
| moments | Writing | Writing / Timeline / Starmap / Conversation | 原话是多个上下文的只读材料 |
| stars | Starmap | Starmap | Trace 被寄存后的星 |
| constellations | Starmap | Starmap | 星座聚合 |
| constellation_stars | Starmap | Starmap | 星星与星座关系 |
| insights | Starmap | Starmap | 星座级观察 |
| past_self_cards | Starmap | Starmap / Conversation | 对话入口 |
| topic_prompts | Starmap | Starmap | 话题引子 |
| chat_sessions | Conversation | Conversation | 对话会话 |
| chat_messages | Conversation | Conversation | 对话消息 |

重要规则：

> 一张表只能有一个上下文负责写入，其他上下文最多只读。

这样可以避免“哪里都能改状态”的混乱。

---

## 17. 事务边界建议

### 17.1 Login

事务范围：

- 创建 User。

JWT 签发可以在事务后进行。

### 17.2 CreateMoment

事务范围：

- 创建 Trace，如果需要。
- 创建 Moment。

不建议把 AI insight 放进同一个事务。

Embedding 有两种策略：

1. 同步生成 embedding，再保存 Moment。
2. 先保存 Moment，再异步补 embedding。

MVP 推荐同步生成，失败则保存 Moment 且 `embedding=nil`，Echo 返回 nil。

### 17.3 GenerateInsight

不需要事务。

它是实时 AI 读操作，失败前端无声降级。

### 17.4 StashTrace

事务范围：

- 校验 Trace。
- 创建 Star。
- 标记 Trace 已寄存，如果有 traces 表。

事务外：

- 聚类。
- 生成星座资产。
- 更新 Moment connected。

这些用领域事件异步完成。

### 17.5 SendMessage

推荐顺序：

1. 事务一：保存用户消息。
2. 事务外：调用 AI。
3. 事务二：保存 AI 回复。

如果 AI 失败，用户消息仍可保留，并返回可恢复错误。

---

## 18. 领域事件

事件用于跨上下文协作，避免直接互相调用内部实现。

推荐事件：

### MomentCreated

由 Writing 发布。

可用于未来：

- 异步补 embedding。
- 更新搜索索引。
- 行为分析。

### TraceStashed

由 Starmap 发布。

包含：

- user_id
- trace_id
- star_id

可触发：

- 聚类。
- 更新 connected。
- 生成星座资产。

### ConstellationFormed

由 Starmap 发布。

可触发：

- 生成 PastSelfCard。
- 生成 TopicPrompt。
- 通知前端刷新。

### ChatMessageAppended

由 Conversation 发布。

可用于未来：

- 对话摘要。
- 安全审计。

MVP 可以用进程内 event bus。

后续建议升级为 outbox：

```text
业务事务写表 + 写 outbox_events
后台 worker 扫描 outbox_events
执行聚类/AI 任务
成功后标记 processed
```

---

## 19. 典型用例的跨领域协作

### 19.1 主流程：写字 → 回声 → 观察

```text
gRPC CreateMoment
  → Writing adapter/grpc
  → Writing app/CreateMomentUseCase
  → Writing domain/Trace append Moment
  → Writing repository save Moment
  → Intelligence EmbeddingProvider
  → Writing domain/EchoMatcher
  → 返回 Echo DTO

gRPC GenerateInsight
  → Writing app/GenerateMomentInsightUseCase
  → Writing repository read echo Moment
  → Intelligence MomentInsightGenerator
  → 返回 Insight DTO
```

边界说明：

- Starmap 不参与。
- Timeline 不参与。
- Echo 不持久化。

### 19.2 顺着再想想

```text
gRPC CreateMoment(trace_id=T1)
  → Writing 确认 Trace 属于当前用户
  → Trace 追加 Moment
  → EchoMatcher 排除 trace_id=T1 的已有 Moment
  → 返回新 Echo
```

边界说明：

- 仍属于 Writing。
- 不是新建星星。
- 不是创建新星座。

### 19.3 收进星图

```text
gRPC StashTrace(trace_id=T1, x, y)
  → Starmap app/StashTraceUseCase
  → 通过 Writing TraceReader 校验 T1
  → Starmap domain 创建 Star
  → Starmap repository 保存 Star
  → 发布 TraceStashed
  → 返回 Star

后台处理 TraceStashed
  → Starmap ClusterStarUseCase
  → 读取 Trace 的 Moment 向量
  → ConstellationClusterer 判断归属
  → 更新 Constellation
  → 调 Intelligence 生成星座资产
```

边界说明：

- StashTrace 属于 Starmap，不属于 Writing。
- 聚类是最终一致。
- 前端的飞星动画不依赖聚类完成。

### 19.4 星座详情

```text
gRPC GetConstellation
  → Starmap query
  → 读取 Constellation / Star / Insight / PastSelfCard / TopicPrompt
  → 通过 Writing MomentReader 读取原话
  → 组装详情 DTO
```

边界说明：

- 详情页 DTO 可以很丰富。
- DTO 丰富不代表一个巨大的领域对象。

### 19.5 和过去自己聊天

```text
gRPC StartChat(past_self_card_id)
  → Conversation app/StartChatUseCase
  → Starmap PastSelfCardReader
  → Writing MomentReader
  → PastSelfPersonaBuilder 构建上下文
  → Intelligence PastSelfResponder 生成开场白
  → Conversation repository 保存 ChatSession/Message

gRPC SendMessage(chat_session_id, content)
  → Conversation app/SendMessageUseCase
  → 追加用户消息
  → 加载历史和 PastSelfContext
  → Intelligence PastSelfResponder 生成回复
  → 校验引用来源
  → 保存 AI 回复
```

边界说明：

- Conversation 只保存聊天，不改星座。
- PastSelfCard 来自 Starmap，但 Conversation 不生成它。
- Moment 是原话材料，只读。

---

## 20. 代码依赖方向

推荐依赖方向：

```text
adapter → app → domain
platform → 被 adapter/app 通过接口使用
domain → 只依赖 shared/domain 或标准库
```

禁止：

```text
domain import proto
domain import pgx/sqlc
domain import grpc/status
domain import platform/ai
domain import config
```

允许：

```text
adapter/grpc import proto
adapter/postgres import pgx/sqlc
app import domain
app depend on domain-defined ports
platform/ai implement domain/app ports
```

---

## 21. Repository 设计原则

不要做过度通用的 `Repository[T]`。

DDD 里仓储应围绕业务意图命名。

### Writing

推荐方法语义：

```text
SaveTrace
AppendMoment
FindMomentByID
FindTraceByID
FindEchoCandidates
ListMomentsByTrace
```

### Starmap

推荐方法语义：

```text
SaveStar
FindStarByTrace
SaveConstellation
AddStarToConstellation
FindConstellationDetail
ReplaceConstellationAssets
```

### Conversation

推荐方法语义：

```text
CreateSession
AppendMessage
FindSession
ListMessages
```

这些名字比 `Insert`、`Update`、`Query` 更能表达业务。

---

## 22. 当前项目的迁移建议

当前 server 已有：

```text
internal/auth
internal/config
internal/db
internal/login
internal/db/sqlc
```

可以渐进演进，不需要一次性大重构。

### 阶段 1：保持现有 Login，重命名边界

将：

```text
internal/login
```

逐步演进为：

```text
internal/identity
```

Login 仍然先跑通。

### 阶段 2：新增 Writing 上下文

实现接口前先建立包边界：

```text
internal/writing/domain
internal/writing/app
internal/writing/adapter/grpc
internal/writing/adapter/postgres
```

先实现：

- CreateMoment
- GenerateInsight

### 阶段 3：新增 Timeline 查询上下文

Timeline 可以很薄：

```text
internal/timeline/app
internal/timeline/adapter/grpc
internal/timeline/adapter/postgres
```

实现：

- ListMoments
- GetRandomMoments

### 阶段 4：新增 Starmap

先实现同步版本：

- StashTrace 创建 Star。
- ListConstellations 返回简单数据。

再实现：

- 异步聚类。
- 星座资产生成。
- GetConstellation。

### 阶段 5：新增 Conversation

实现：

- StartChat
- SendMessage

AI prompt 和引用校验一定要放到 Conversation 的用例规则里，不要只靠前端。

---

## 23. MVP 可以简化的地方

DDD 不等于一上来写很多抽象。

MVP 可以简化：

- 先不用 outbox，用 goroutine。
- 先不用独立 read model，Timeline 直接读 moments。
- 先不新增 traces 表，但领域里保留 Trace 概念。
- 先用 sqlc 生成查询，再在 adapter 中转换成 domain。
- 先把 prompt 模板放 `platform/ai/prompts.go`，后续再版本化。

但不建议简化：

- 不要让 domain 依赖 proto。
- 不要让所有业务都堆进一个 `service.go`。
- 不要让 AI SDK 调用散落在各个 handler。
- 不要让多个上下文随意写同一张表。
- 不要把 Echo、Star、Constellation、Chat 都做成贫血的 `model.Model`。

---

## 24. 最终推荐架构形态

ego 后端建议采用：

```text
模块化单体
+ DDD 限界上下文
+ 六边形端口适配器
+ PostgreSQL/pgvector
+ gRPC API
+ AI 反腐层
+ 领域事件驱动的最终一致
```

一句话总结：

> Writing 负责“此刻如何产生回声”，Starmap 负责“痕迹如何形成星座”，Conversation 负责“如何和过去的自己说话”，Timeline 负责“如何回看过去”，Identity 负责“谁在使用”。AI 和数据库都只是支撑这些领域的工具。

