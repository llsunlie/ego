# AGENTS.md - Writing 模块开发指南

## 0. 全局上下文溯源

如果你是在 `server/internal/writing/` 目录下被唤醒的 Agent，请在执行任何代码前先向上读取：

1. `../../../AGENTS.md`
2. `../../AGENTS.md`


## 1. 模块定位

Writing 是“此刻写作上下文”，回答：

> 用户刚写下的话如何成为 Moment，并与过去产生回声？

本模块负责 Trace、Moment、Echo，以及当前体验中的实时 Insight。

## 2. 负责的接口范围

本模块负责实现以下前后端 RPC：

| RPC | 责任 |
| --- | --- |
| `CreateMoment` | 创建或延续 Trace，保存 Moment，返回 Echo 和候选回声 |
| `GenerateInsight` | 基于当前内容与 Echo Moment 生成当前体验的“我发现” |

## 3. 模块边界

### 3.1 拥有的业务能力

- 创建 Trace，或延续已有 Trace。
- 追加 Moment。
- 生成或接收 Moment embedding。
- 匹配当前用户历史 Moment 作为 Echo。
- 生成当前会话级 Insight。
- 为其他模块提供 Trace/Moment 的只读契约。

### 3.2 数据归属

Writing 拥有唯一写入权：

```text
moments
traces（未来显式建表时）
```

其他模块如 Starmap、Timeline、Conversation 只能通过明确的只读契约读取 Moment/Trace，不允许直接创建或更新。


## 5. 常用开发命令

从 `server/` 目录运行：

```text
go test ./internal/writing/...
go test ./...
go build ./...
```
