# AGENTS.md - Timeline 模块开发指南

## 0. 全局上下文溯源

如果你是在 `server/internal/timeline/` 目录下被唤醒的 Agent，请在执行任何代码前先向上读取：

1. `../../../AGENTS.md`
2. `../../AGENTS.md`


Timeline 是查询上下文，核心边界是“只读”。

## 1. 模块定位

Timeline 负责“过往”页面和记忆光点盲盒。

它回答：

> 用户如何回看自己过去说过的话？

本模块不拥有 Moment 的创建权，只负责面向前端的 Moment 读取视图。

## 2. 负责的接口范围

本模块负责实现以下前后端 RPC：

| RPC | 责任 |
| --- | --- |
| `ListMoments` | 返回当前用户的过往 Moment 列表，支持游标分页 |
| `GetRandomMoments` | 返回当前用户随机历史 Moment，用于记忆光点盲盒 |

## 3. 模块边界

### 3.1 拥有的业务能力

- 按时间倒序读取 Moment。
- 支持 cursor/page_size 分页。
- 随机读取 N 条历史 Moment。
- 返回前端需要的 `connected` 状态。

### 3.2 数据归属

Timeline 不拥有任何业务表写入权。

允许读取：

```text
moments
```

Moment 写入权属于 Writing。

### 3.3 禁止事项

- 禁止创建、更新或删除 Moment。
- 禁止更新 `connected` 状态。
- 禁止写入星图、星座、聊天相关表。
- 禁止在 Timeline 中实现 Echo 匹配、StashTrace 或 Chat。
- 禁止绕过用户隔离读取其他用户数据。

## 4. 依赖规则

- Timeline 可以使用只读 read model。
- 如果读取字段不足，应通过 Writing 的契约或数据库 read model 扩展，而不是获得 Moment 写入权。
- Timeline 不需要复杂 domain 聚合；优先保持查询模型简单清晰。

## 5. 常用开发命令

从 `server/` 目录运行：

```text
go test ./internal/timeline/...
go test ./...
go build ./...
```

