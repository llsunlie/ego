# AGENTS.md - Conversation 模块开发指南

## 0. 全局上下文溯源

如果你是在 `server/internal/conversation/` 目录下被唤醒的 Agent，请在执行任何代码前先向上读取：

1. `../../../AGENTS.md`
2. `../../AGENTS.md`

Conversation 是"和过去的自己对话"的上下文。它只管理聊天会话与消息，不生成星座资产。

## 1. 模块定位

Conversation 回答：

> 用户如何和某段过去的自己说话？

本模块负责 ChatSession、ChatMessage、PastSelfContext 构建、AI past-self 回复编排，以及引用来源校验。

## 2. 负责的接口范围

本模块负责实现以下前后端 RPC：

| RPC | 责任 |
| --- | --- |
| `StartChat` | 基于星图和上下文 Moment 创建或恢复 ChatSession，并返回开场白/历史 |
| `SendMessage` | 保存用户消息，生成并保存 past-self 回复，返回带引用来源的 ChatMessage |

## 3. 模块边界

### 3.1 拥有的业务能力

- 创建或恢复 ChatSession。
- 追加用户 ChatMessage。
- 构建 PastSelfContext。
- 调用 ChatGenerator 生成第一人称回复。
- 校验 AI 回复的引用来源和越界行为。
- 保存 AI 回复及 `referenced_moments`。

### 3.2 数据归属

Conversation 拥有唯一写入权：

```text
chat_sessions
chat_messages
```

允许读取：

```text
stars（通过 Starmap 契约）
moments（通过 Writing 契约）
```

### 3.3 禁止事项

- 禁止生成或修改 PastSelfCard。
- 禁止修改 Constellation、Star、TopicPrompt。
- 禁止创建或修改 Moment。
- 禁止在 `domain/` 中直接调用 AI SDK、pgx、sqlc 或 proto。

## 4. 架构与装配

- **两级装配**：`conversation/module.go` 负责组装本模块的 adapter、app use case、gRPC handler；`bootstrap/chat.go` 只注入 DB 等进程级资源。
- **业务策略归位 app**：DefaultChatGenerator（开场白和回复生成策略）属于 Conversation 业务逻辑，位于 `app/`。
- **Domain ports**：Conversation 定义自己的端口（ChatSessionRepository、ChatMessageRepository、StarReader、MomentReader、ChatGenerator），由 adapter/postgres 实现。
- **Mapper 在 adapter/grpc**：proto 映射逻辑在 `mapper.go`，handler 纯粹做 ctx→input→output→pb 转换。

## 5. 常用开发命令

从 `server/` 目录运行：

```text
go test ./internal/conversation/...
go test ./...
go build ./...
```
