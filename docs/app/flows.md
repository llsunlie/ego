# ego · 核心流程文档

> 对照 `api.proto` 验证接口覆盖。

## 目录

1. [F0 登录/自动注册](#f0-登录自动注册)
2. [F1 主流程：写字 → 回声 → 观察](#f1-主流程写字--回声--观察)
3. [F2 顺着再想想：深度探索循环](#f2-顺着再想想深度探索循环)
4. [F3 收进星图：保存 + 仪式感](#f3-收进星图保存--仪式感)
5. [F4 记忆光点：盲盒](#f4-记忆光点盲盒)
6. [F5 过往浏览：时间线](#f5-过往浏览时间线)
7. [F6 星图探索：星座 + 详情](#f6-星图探索星座--详情)
8. [F7 和那时的自己说说话：对话模式](#f7-和那时的自己说说话对话模式)
9. [F8 话题引子：从星图回到此刻](#f8-话题引子从星图回到此刻)
10. [F9 冷启动：用户首次使用](#f9-冷启动用户首次使用)

---

## F0 登录/自动注册

**路径：** LoginPage → NowPage

```
┌─────────────────────────────────────────────────────────┐
│ 1. App 启动                                               │
│    ├─ 本地无 JWT token → 显示 LoginPage                   │
│    ├─ 本地有 token → 直接进入 NowPage                      │
│    └─ token 过期 → 后端返回 UNAUTHENTICATED → 回 LoginPage │
│                                                         │
│ 2. 用户输入 account + password，点登录                      │
│    ├─ [API] Ego.Login(account, password)                 │
│    │     → 后端查 users 表                                │
│    │     → 不存在: INSERT 新用户 + 返回 JWT (created=true) │
│    │     → 存在: bcrypt 验证 + 返回 JWT (created=false)  │
│    └─ 前端存 token 到本地 → 跳转 NowPage                   │
│                                                         │
│ 3. 后续所有请求携带 JWT                                    │
│    └─ gRPC metadata: Authorization: Bearer <token>       │
└─────────────────────────────────────────────────────────┘
```

**覆盖状态：** `Login` ✓（新增）

---

## F1 主流程：写字 → 回声 → 观察

**路径：** NowPage（idle → writing → echoing → idle）

```
┌─────────────────────────────────────────────────────────┐
│ 1. 用户进入 App，看到主屏                                 │
│    ├─ 星空背景 + 呼吸光团                                 │
│    ├─ 3 颗记忆光点漂浮（→ F4）                            │
│    └─ "有什么想说的吗" + "写下此刻"                        │
│                                                         │
│ 2. 点击"写下此刻"                                         │
│    ├─ 光团缩小上移，WritingInput 从底部滑出                │
│    └─ placeholder: "随便说点什么，这里听着……"               │
│                                                         │
│ 3. 用户输入内容，点"先到这"                                │
│    ├─ [API] Ego.CreateMoment(content, trace_id)          │
│    │     → 后端保存 Moment + embedding                    │
│    │     → 后端语义匹配历史 → 持久化 Echo                  │
│    │       (matched_moment_ids[] + similarities[])       │
│    │     → 前端渲染 EchoCard                             │
│    ├─ [API] Ego.GenerateInsight(moment_id, echo_id)      │
│    │     → 后端 AI 生成第二人称观察                        │
│    │     → 持久化 Insight                                │
│    │     → 前端渲染 InsightCard                          │
│    └─ 两个接口顺序调用，EchoCard 先出，InsightCard 紧随    │
│                                                         │
│ 4. 用户看到三出口按钮                                     │
│    ├─ "顺着再想想"（金色主按钮）→ F2                       │
│    ├─ "✦ 收进星图" → F3                                  │
│    └─ "嗯，先这样" → 回到 idle                            │
└─────────────────────────────────────────────────────────┘
```

**接口调用序列：**
```
CreateMoment ──┬── 成功 → 持久化 Moment + Echo → 渲染 EchoCard
               │
               ├── GenerateInsight(moment_id, echo_id)
               │    └── 成功 → 持久化 Insight → 渲染 InsightCard
               │
               └── 失败/超时 → 隐藏 InsightCard（无声降级）
```

**覆盖状态：** `CreateMoment` ✓ `GenerateInsight` ✓

---

## F2 顺着再想想：深度探索循环

**路径：** NowPage（echoing → writing → echoing → writing → ... 可无限循环）

```
┌─────────────────────────────────────────────────────────┐
│ 1. 用户在 EchoCard 下方点"顺着再想想"                      │
│    ├─ 输入区重新滑出                                      │
│    ├─ 保留之前的 trace_id（不新建）                        │
│    └─ placeholder 可变为更轻的语气                         │
│                                                         │
│ 2. 用户继续写，点"先到这"                                  │
│    ├─ [API] Ego.CreateMoment(content, trace_id)          │
│    │     → 后端追加 Moment 到已有 Trace                   │
│    │     → 后端重新匹配回声 → 返回新 Echo                  │
│    │                                                     │
│    ├─ [API] Ego.GenerateInsight(moment_id, echo_id)      │
│    │     → AI 基于新 Moment + 新 Echo 生成观察             │
│    │                                                     │
│    └─ 继续循环，直到用户点"收进星图"或"先这样"             │
└─────────────────────────────────────────────────────────┘
```

**Trace 结构示意：**
```
Trace
├── Moment 1: "今天和同事起了冲突"      ← 第一次写
│   └── Echo: "你之前也说过类似的..."
│   └── Insight: "两次里，你用词不一样..."
├── Moment 2: "其实是我害怕被否定"      ← 顺着再想想 第 1 轮
│   └── Echo: "..."
│   └── Insight: "..."
└── Moment 3: "小时候也是这样"          ← 顺着再想想 第 2 轮
    └── Echo: "..."
    └── Insight: "..."
```

**接口调用序列：**
```
第 1 轮: CreateMoment("今天和同事...") → moment + echo
        GenerateInsight(moment_id, echo_id) → insight
第 2 轮: CreateMoment("其实是我害怕...", trace_id=T1) → moment + echo
        GenerateInsight(moment_id, echo_id) → insight
第 3 轮: CreateMoment("小时候...", trace_id=T1) → moment + echo
        GenerateInsight(moment_id, echo_id) → insight
...
收进星图: StashTrace(trace_id=T1)
```

**覆盖状态：** `CreateMoment`（含 trace_id 复用）✓ `GenerateInsight` ✓

---

## F3 收进星图：保存 + 仪式感

**路径：** NowPage（echoing → stashing → idle）

```
┌─────────────────────────────────────────────────────────┐
│ 1. 用户在 EchoCard 下方点"✦ 收进星图"                     │
│                                                         │
│ 2. [API] Ego.StashTrace(trace_id)                       │
│    ├─ 后端：                                              │
│    │   ├─ 标记 traces.stashed = true                    │
│    │   ├─ 创建 Star（trace_id, topic）                   │
│    │   ├─ 异步聚类：                                      │
│    │   │   ├─ 与已有 Star 合并/新建 Constellation         │
│    │   │   ├─ 更新 Constellation.star_ids               │
│    │   │   ├─ 生成/更新 Constellation 级 Insight          │
│    │   │   ├─ 生成 TopicPrompts                         │
│    │   │   └─ 孤星也是 Constellation（star_count=1）     │
│    │   └─ 返回 Star                                     │
│    │                                                     │
│ 3. 前端播放仪式感动效                                     │
│    ├─ 卡片发光 → 三圈金色涟漪                              │
│    ├─ 光点沿贝塞尔曲线飞向星图 Tab                          │
│    ├─ 拖尾 → 星屑炸开 → Tab 金色脉冲                       │
│    └─ 动效期间 StashTrace 已异步完成                       │
│                                                         │
│ 4. 重置 NowPage 状态到 idle                              │
│    └─ 如果用户切换到星图 Tab：                             │
│       [API] Ego.ListConstellations → 获取最新星座列表     │
└─────────────────────────────────────────────────────────┘
```

**覆盖状态：** `StashTrace` ✓ `ListConstellations`（stash 后刷新用）✓

---

## F4 记忆光点：盲盒

**路径：** NowPage（首页常驻 + 进入时刷新）

```
┌─────────────────────────────────────────────────────────┐
│ 1. 进入 NowPage 时                                        │
│    ├─ [API] Ego.GetRandomMoments(count=3)                │
│    │     → 后端从当前用户的 Moment 中随机抽取 3 条         │
│    └─ 前端渲染 3 颗光点（暖金/冷蓝/柔紫），漂浮在光团周围  │
│                                                         │
│ 2. 用户点击一颗光点                                       │
│    ├─ 光点展开为 EnvelopeCard 信封                        │
│    ├─ 展示：日期 + 原话全文                                │
│    └─ 纯前端动画，数据已在 GetRandomMoments 中返回         │
│                                                         │
│ 3. 关闭信封                                              │
│    └─ 点击外部 / 关闭按钮 → 信封收回为光点                  │
│                                                         │
│ 4. 下拉刷新 → 重新 GetRandomMoments(count=3)              │
│                                                         │
│ 特殊情况：无历史数据                                      │
│    └─ GetRandomMoments 返回空 → 不显示光点，只有光团       │
└─────────────────────────────────────────────────────────┘
```

**覆盖状态：** `GetRandomMoments` ✓

---

## F5 过往浏览：时间线

**路径：** Tab 切换到 PastPage

```
┌─────────────────────────────────────────────────────────┐
│ 1. 切换到"过往"Tab                                        │
│    ├─ [API] Ego.ListTraces(cursor="", page_size=20)      │
│    │     → 返回 Trace 列表（轻量，不含 Items）              │
│    │     → 前端按 created_at 按月分组                     │
│    │     → 每条 Trace 含 stashed 字段（是否已寄存）         │
│    │     → stashed = true 显示 ✦ 已联结 标记              │
│    └─ 前端渲染时间线                                      │
│                                                         │
│ 2. 滚动到底部（加载更多）                                  │
│    ├─ [API] Ego.ListTraces(cursor=last_id, page_size=20) │
│    └─ 追加到列表尾部                                      │
│                                                         │
│ 3. 点击某条 Trace                                        │
│    ├─ [API] Ego.GetTraceDetail(trace_id)                 │
│    │     → 返回 Trace + TraceItem[]                      │
│    │     → TraceItem = <Moment, Echo[], Insight>         │
│    │     → 展开后用户看到：                                │
│    │        ├─ 原话列表（Moment[]）                       │
│    │        ├─ 每条 Moment 对应的回声（Echo[]）            │
│    │        └─ 每条 Moment 的 AI 观察（Insight）          │
│    └─ 纯前端展开/折叠                                    │
│                                                         │
│ 4. ✦ 已联结 标记                                         │
│    └─ traces.stashed == true 时显示                      │
│       此字段由 StashTrace 后更新                          │
└─────────────────────────────────────────────────────────┘
```

**覆盖状态：** `ListTraces` ✓ `GetTraceDetail` ✓ （游标分页，前端按月分组）

---

## F6 星图探索：星座 + 详情

**路径：** Tab 切换到 StarmapPage → push ConstellationDetailPage

```
┌─────────────────────────────────────────────────────────┐
│ 1. 切换到"星图"Tab                                        │
│    ├─ [API] Ego.ListConstellations()                     │
│    │     → 返回所有 Constellation（含 star_ids, star_count）│
│    │     → 返回 total_star_count（"已有 N 颗星"）          │
│    └─ 前端渲染星图：                                       │
│        ├─ 前端按 star_count 判断状态：                     │
│        │   star_count=1 → 孤星（独立弱光）                 │
│        │   star_count=2 → 正在成型（弱光脉冲+"隐约有什么…"）│
│        │   star_count≥3 → 已成型（亮星+连线+标签名）       │
│        └─ 坐标、闪烁节奏由前端自行分配                     │
│                                                         │
│ 2. 用户在星图上缩放/平移                                    │
│    └─ InteractiveViewer，纯前端手势                       │
│                                                         │
│ 3. 点击一颗星座 → push 详情页                              │
│    ├─ [API] Ego.GetConstellation(constellation_id)       │
│    │     → 返回全量详情数据：                              │
│    │        ├─ ① Constellation（含 constellation_insight） │
│    │        ├─ ② Moment[]（主题内所有原话）                 │
│    │        ├─ ③ Star[]（前端组装为 PastSelfCard）         │
│    │        └─ ④ topic_prompts 在 Constellation 消息中    │
│    └─ 前端渲染详情页                                      │
│                                                         │
│ 4. 展开"主题里说过的话"                                    │
│    └─ ExpansionTile → 显示原话列表                         │
│       （纯前端，数据已在 GetConstellation 中返回）           │
└─────────────────────────────────────────────────────────┘
```

**覆盖状态：** `ListConstellations` ✓ `GetConstellation` ✓

---

## F7 和那时的自己说说话：对话模式

**路径：** ConstellationDetailPage → ChatDialog

```
┌─────────────────────────────────────────────────────────┐
│ 1. 用户在详情页看到 PastSelfCard × N                      │
│    ├─ 每张卡片 = 一个 Star（前端从 GetConstellation.stars[] │
│    │   组装为 PastSelfCard）                              │
│    ├─ 展示：Star.topic 作为标题                            │
│    │   该 Star 对应 Trace 最后一个 Moment 作为开场白         │
│    └─ 数据来自 Ego.GetConstellation 返回的 stars[]        │
│                                                         │
│ 2. 点击一张卡片 → 拉起对话浮层                              │
│    ├─ [API] Ego.StartChat(star_id,                        │
│    │                       context_moment_ids)            │
│    │     → 后端构建 past-self 上下文                      │
│    │        ├─ 原话（思想灵魂）— 不可伪造                   │
│    │        ├─ 延伸（语气/风格）— 灵动回应当下               │
│    │        └─ 标注（引用来源）— 每条回复明确出处            │
│    │     → 返回 chat_session_id + AI 开场白               │
│    │                                                     │
│    └─ 前端渲染对话界面，开场白显示                           │
│                                                         │
│ 3. 用户输入消息，点发送                                    │
│    ├─ [API] Ego.SendMessage(chat_session_id, content)    │
│    │     → 后端 AI 以 past-self 第一人称回复              │
│    │     → 回复底部标注 "参考了 X月X日 前后"                │
│    │                                                     │
│    └─ 前端追加 ChatBubble                                │
│                                                         │
│ 4. 多轮对话                                              │
│    ├─ 用户 → 发送 → Ego.SendMessage → AI 回复 → 循环      │
│    │                                                     │
│    ├─ AI 行为约束（后端保证）：                             │
│    │   ├─ 可反问用户："你呢，现在怎么想的？"                │
│    │   ├─ 可自省："我那时候还没想明白"                     │
│    │   ├─ 不可捏造话题（超出范围的回复：                     │
│    │   │   "这个我那时候没想过，要不要你说给我听？")        │
│    │   └─ 每条回复标注引用来源                             │
│    │                                                     │
│    └─ 5. 关闭浮层 → 对话结束，chatSession 保留可恢复        │
│                                                         │
│ 特殊情况：                                               │
│    离线/超时 → Chat 不可用，按钮置灰                        │
│    超过 N 张 PastSelfCard → 折叠 + "…还有 N 条"           │
└─────────────────────────────────────────────────────────┘
```

**覆盖状态：** `StartChat` ✓ `SendMessage` ✓

> StartChat 支持传入旧 `chat_session_id` 恢复历史对话，无需额外的 `GetChatHistory` 接口。

---

## F8 话题引子：从星图回到此刻

**路径：** ConstellationDetailPage → pop → NowPage（带话题）

```
┌─────────────────────────────────────────────────────────┐
│ 1. 用户在详情页看到 TopicPromptCard × N                   │
│    ├─ 每个引子为一条文本（Constellation.topic_prompts[]）   │
│    └─ 数据来自 Ego.GetConstellation 返回的                  │
│       Constellation.topic_prompts                        │
│                                                         │
│ 2. 点击话题引子卡片                                       │
│    ├─ Navigator.pop() 回到星图 Tab                        │
│    ├─ 切换到 NowPage Tab                                 │
│    └─ 携带 motivation: "constellation:<id>" 进入写字区     │
│       （纯前端路由传参）                                   │
│                                                         │
│ 3. 用户看到输入区，placeholder 携带话题上下文               │
│    └─ 用户写完后走完整 F1 主流程                          │
│                                                         │
│ 新建 Trace 时带上 motivation                             │
│ └─ Traces.motivation = "constellation:<id>"              │
│    后端持久化，PastPage 可追溯灵感来源                     │
└─────────────────────────────────────────────────────────┘
```

**覆盖状态：** `GetConstellation`（已含 topic_prompts）✓

---

## F9 冷启动：用户首次使用

**路径：** 当前用户无任何历史数据时的体验

```
┌─────────────────────────────────────────────────────────┐
│ 1. 用户登录后首次进入 NowPage                               │
│                                                         │
│ 2. 进入 NowPage                                          │
│    ├─ [API] Ego.GetRandomMoments(count=3)                │
│    │     → 当前用户无历史数据，返回空                       │
│    │     → 不显示记忆光点                                 │
│    └─ 只显示光团 + "有什么想说的吗" + "写下此刻"            │
│                                                         │
│ 3. 用户写第一句话，提交                                    │
│    ├─ [API] Ego.CreateMoment(content)                    │
│    │     → 保存 Moment                                   │
│    │     → echo: null（无历史可匹配）                      │
│    │     → 不显示 EchoCard，只显示操作确认                 │
│    │                                                     │
│    ├─ Echo 匹配为空 → 前端不调用 GenerateInsight          │
│    │     → 不显示 InsightCard                             │
│    └─ 只显示三个出口按钮                                   │
│                                                         │
│ 4. "收进星图" → 第一颗星                                   │
│    ├─ [API] Ego.StashTrace(trace_id)                      │
│    │     → 创建第一颗 Star                                │
│    │     → status: lone（孤星，等待同伴）                   │
│    └─ 星图 Tab → 已有 1 颗星                              │
│                                                         │
│ 5. 用户写第二句话 → 可能有回声了                            │
│    └─ 产品体验正式开始                                    │
└─────────────────────────────────────────────────────────┘
```

**覆盖状态：** 所有接口需处理当前用户空数据返回：
- `Login` → 首次登录自动注册
- `GetRandomMoments` → 返回空列表
- `CreateMoment` → echo 字段为 null
- `GenerateInsight` → insight 字段为 null（echo 为 null 时前端不调用）
- `ListTraces` → traces 为空
- `ListConstellations` → constellations 为空

