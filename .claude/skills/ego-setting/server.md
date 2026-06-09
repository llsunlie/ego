# ego-setting server

后端 setting 领域 context — 账号信息只读查询。

## 所属 gRPC 方法

| RPC | 功能 | 认证 |
|-----|------|------|
| `GetProfile` | 查询当前用户手机号 + 注册时间 | 需 JWT（从 context 提取 user_id） |
| `SubmitFeedback` | 提交用户文字反馈 | 需 JWT（从 context 提取 user_id） |

## 模块结构 (`server/internal/setting/`)

```
setting/
├── module.go                          # 依赖注入
├── domain/
│   ├── ports.go                       # UserInfo{Phone, CreatedAt} + UserReader + FeedbackWriter interface
│   ├── feedback.go                    # Feedback 实体
│   └── errors.go                      # ErrUserNotFound, ErrFeedbackEmpty
├── app/
│   ├── profile.go                     # GetProfileUseCase
│   └── feedback.go                    # SubmitFeedbackUseCase + IDGenerator interface
└── adapter/
    ├── id/uuid.go                     # UUID ID 生成器
    ├── grpc/handler.go                # gRPC handler + mapError
    ├── postgres/user_reader.go        # UserReader 实现（sqlc GetUserByID）
    └── postgres/feedback_writer.go    # FeedbackWriter 实现（sqlc InsertFeedback）
```

## 用例详解

### GetProfile (`app/profile.go`)

```
1. 接收 userID（由 handler 从 context.Value("user_id") 提取）
2. userReader.FindByID(userID) → ErrUserNotFound
3. 返回 ProfileResult{Phone, CreatedAt}
```

### SubmitFeedback (`app/feedback.go`)

```
1. 接收 userID + content（由 handler 从 context 提取 user_id）
2. 校验 content 非空 → ErrFeedbackEmpty
3. IDGenerator.New() 生成 UUID
4. feedbackWriter.Save(feedback) → 写库
5. 返回 FeedbackResult{ID, CreatedAt}
```

## 依赖

```go
type Deps struct {
    DB sqlc.DBTX  // 复用 identity 的 users 表查询 + feedbacks 表写入
}
```

## 错误映射

| 领域错误 | gRPC Status | Message |
|----------|-------------|---------|
| `ErrUserNotFound` | `NotFound` | "用户不存在" |
| `ErrFeedbackEmpty` | `InvalidArgument` | "反馈内容不能为空" |
| 未登录（无 user_id） | `Unauthenticated` | "未登录" |

## 与 identity 的关系

- setting 模块**只读** users 表，不写入 users 表
- 通过 sqlc `GetUserByID` 查询（与 identity 的 `GetUserByPhone` 共享 users.sql 查询文件）
- 独立拥有 `feedbacks` 表的写入权（migration 011）
- 独立的 domain/app/adapter 层，不与 identity 相互依赖
