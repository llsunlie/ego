# User Feedback Channel — Design Spec

**Date:** 2026-06-09
**Status:** Approved

## Overview

在 setting 模块「关于」区新增用户反馈渠道。用户点击入口进入独立反馈页面，填写纯文字反馈内容后提交到后端存储。

---

## 1. Proto 变更

**文件:** `proto/ego/api.proto`

### 新增 RPC（Setting 段）

```protobuf
rpc SubmitFeedback(SubmitFeedbackReq) returns (SubmitFeedbackRes);
```

### 新增 Message

```protobuf
message SubmitFeedbackReq {
  string content = 1;  // 反馈文本内容
}

message SubmitFeedbackRes {
  string id         = 1;  // 反馈记录 ID
  int64  created_at = 2;  // 提交时间 unix timestamp ms
}
```

### 认证

需 JWT，从 context 提取 user_id。

---

## 2. 后端 DDD 变更

### 2a. domain/ — 实体 + 接口 + 错误

**新文件:** `server/internal/setting/domain/feedback.go`

- `Feedback` 实体：ID, UserID, Content, CreatedAt

**修改:** `server/internal/setting/domain/ports.go`

- 新增 `FeedbackWriter` 接口：`Save(ctx, *Feedback) error`

**修改:** `server/internal/setting/domain/errors.go`

- 新增 `ErrFeedbackEmpty`

### 2b. app/ — 用例

**新文件:** `server/internal/setting/app/feedback.go`

- `SubmitFeedbackUseCase`：校验非空 → 生成 UUID → 保存 → 返回结果

### 2c. adapter/grpc/ — Handler

**修改:** `server/internal/setting/adapter/grpc/handler.go`

- 新增 `SubmitFeedback` 方法：提取 user_id → 委托 use case → mapError

### 2d. adapter/postgres/ — 仓储

**新文件:** `server/internal/setting/adapter/postgres/feedback_writer.go`

- `FeedbackWriter` 实现，写入 `feedbacks` 表

---

## 3. 数据库

### Migration SQL

```sql
CREATE TABLE feedbacks (
    id         VARCHAR(64)  PRIMARY KEY,
    user_id    VARCHAR(64)  NOT NULL REFERENCES users(id),
    content    TEXT         NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);
```

---

## 4. 前端变更

### 4a. 设置页入口

**修改:** `client/lib/features/setting/setting_page.dart`

- 「关于」区隐私政策行下方新增「用户反馈」行
- 图标：`Icons.favorite_border` 或 `Icons.feedback_outlined`
- 标签：「用户反馈」
- 右箭头 + `context.push('/feedback')`

### 4b. 反馈页面

**新文件:** `client/lib/features/setting/feedback_page.dart`

- `FeedbackPage` — `ConsumerStatefulWidget`
- Scaffold + StarryBackground + AppBar（金色标题「用户反馈」）
- 引导文字 + 多行 TextField + 提交按钮
- 状态：idle → submitting → success（SnackBar + 返回）/ error（SnackBar 提示）

### 4c. 路由

**修改:** `client/lib/core/router/router.dart`

- 新增 `/feedback` 路由（GoRoute，需登录）

### 4d. EgoClient

**修改:** `client/lib/data/services/ego_client.dart`

- 新增 `submitFeedback(Ref ref, String content)` 方法

---

## 5. 错误映射

| 领域错误 | gRPC Status | Message |
|----------|-------------|---------|
| `ErrFeedbackEmpty` | `InvalidArgument` | "反馈内容不能为空" |
| 未登录（无 user_id） | `Unauthenticated` | "未登录" |

---

## 6. 涉及文件汇总

| 操作 | 文件 |
|------|------|
| 修改 | `proto/ego/api.proto` |
| 生成 | `server/proto/ego/`（make proto-go） |
| 生成 | `client/lib/data/generated/`（make proto-dart） |
| 修改 | `server/internal/setting/domain/ports.go` |
| 修改 | `server/internal/setting/domain/errors.go` |
| 新增 | `server/internal/setting/domain/feedback.go` |
| 新增 | `server/internal/setting/app/feedback.go` |
| 修改 | `server/internal/setting/adapter/grpc/handler.go` |
| 新增 | `server/internal/setting/adapter/postgres/feedback_writer.go` |
| 修改 | `server/internal/setting/module.go` |
| 新增 | `server/internal/platform/postgres/migrations/` (SQL) |
| 修改 | `server/internal/bootstrap/composite.go` |
| 修改 | `client/lib/features/setting/setting_page.dart` |
| 新增 | `client/lib/features/setting/feedback_page.dart` |
| 修改 | `client/lib/core/router/router.dart` |
| 修改 | `client/lib/data/services/ego_client.dart` |
