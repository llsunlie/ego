# ego-setting server

后端 setting 领域 context — 账号信息只读查询。

## 所属 gRPC 方法

| RPC | 功能 | 认证 |
|-----|------|------|
| `GetProfile` | 查询当前用户手机号 + 注册时间 | 需 JWT（从 context 提取 user_id） |

## 模块结构 (`server/internal/setting/`)

```
setting/
├── module.go                          # 依赖注入
├── domain/
│   ├── ports.go                       # UserInfo{Phone, CreatedAt} + UserReader interface
│   └── errors.go                      # ErrUserNotFound
├── app/
│   └── profile.go                     # GetProfileUseCase
└── adapter/
    ├── grpc/handler.go                # gRPC handler + mapError
    └── postgres/user_reader.go        # UserReader 实现（sqlc GetUserByID）
```

## 用例详解

### GetProfile (`app/profile.go`)

```
1. 接收 userID（由 handler 从 context.Value("user_id") 提取）
2. userReader.FindByID(userID) → ErrUserNotFound
3. 返回 ProfileResult{Phone, CreatedAt}
```

## 依赖

```go
type Deps struct {
    DB sqlc.DBTX  // 复用 identity 的 users 表查询
}
```

## 错误映射

| 领域错误 | gRPC Status | Message |
|----------|-------------|---------|
| `ErrUserNotFound` | `NotFound` | "用户不存在" |
| 未登录（无 user_id） | `Unauthenticated` | "未登录" |

## 与 identity 的关系

- setting 模块**只读** users 表，不写入
- 通过 sqlc `GetUserByID` 查询（与 identity 的 `GetUserByPhone` 共享 users.sql 查询文件）
- 独立的 domain/app/adapter 层，不与 identity 相互依赖
