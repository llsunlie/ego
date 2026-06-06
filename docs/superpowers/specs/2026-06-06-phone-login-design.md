# Login 升级：手机号注册/登录

日期: 2026-06-06

## 背景

当前 ego 的 login 是「账号 + 密码 → 登录/注册合一」。改为仅支持手机号，通过阿里云短信认证服务验证手机号，注册需短信验证码 + 设置密码，登录需手机号 + 密码。

## Proto 契约 (`proto/ego/api.proto`)

### 新增 RPC

```proto
service Ego {
  // ... 已有 rpc
  rpc SendVerificationCode(SendVerificationCodeReq) returns (SendVerificationCodeRes);
  rpc Register(RegisterReq) returns (RegisterRes);
}
```

### 新增消息

```proto
message SendVerificationCodeReq {
  string phone = 1;
}
message SendVerificationCodeRes {
  bool registered = 1;  // 手机号是否已注册（前端据此决定下一步展示密码登录还是验证码注册）
}
```

```proto
message RegisterReq {
  string phone    = 1;
  string code     = 2;  // 短信验证码
  string password = 3;
}
message RegisterRes {
  string token = 1;
}
```

### 修改消息

```proto
// LoginReq: account → phone
message LoginReq {
  string phone    = 1;
  string password = 2;
}
// LoginRes: 去掉 created
message LoginRes {
  string token = 1;
}
```

## 后端 — identity 领域 (`server/internal/identity/`)

### domain/user.go — User 模型

```go
type User struct {
    ID           string
    Phone        string   // 原 Account → Phone
    PasswordHash string
    CreatedAt    time.Time
}

type UserRepository interface {
    FindByPhone(ctx context.Context, phone string) (*User, error)
    Create(ctx context.Context, user *User) error
}
```

### app/ports.go — 新增接口

```go
type SmsService interface {
    Send(ctx context.Context, phone string) error
    Verify(ctx context.Context, phone, code string) (bool, error)
}
```

### 新增 app/send_code.go — SendCode 用例

- 校验手机号格式
- 调用 `SmsService.Send(phone)` 发送短信

### 新增 app/register.go — Register 用例

- 校验手机号 + 密码格式
- 调用 `SmsService.Verify(phone, code)` 校验验证码
- 查重（手机号已注册则返回 `ErrPhoneAlreadyRegistered`）
- bcrypt 哈希密码 → 创建 User → 签发 JWT

### 修改 app/login.go — Login 用例

- `account` → `phone`，去掉注册逻辑（注册已拆分）
- 流程：`FindByPhone` → `Verify` 密码 → 签发 JWT

### adapter/sms/aliyun.go — 阿里云 SMS

```go
type AliyunSmsService struct { ... }
func (s *AliyunSmsService) Send(ctx context.Context, phone string) error
func (s *AliyunSmsService) Verify(ctx context.Context, phone, code string) (bool, error)
```

### adapter/grpc/handler.go — Handler 变更

- 新增 `SendVerificationCode`、`Register` 方法
- `Login` 方法签名更新（account → phone）
- 错误新增：`ErrPhoneAlreadyRegistered`、`ErrInvalidVerificationCode`、`ErrCodeExpired`、`ErrInvalidPhone`

### domain/errors.go — 新增领域错误

```go
ErrPhoneAlreadyRegistered
ErrInvalidVerificationCode
ErrCodeExpired
ErrInvalidPhone
```

### module.go — 模块组装

新增 `SmsService smsService` 到 Deps，注入到 `SendCodeUseCase` 和 `RegisterUseCase`。

## 前端 — login page (`client/`)

### 流程（3 步）

```
Step 0: 输入手机号 → "下一步"
  → 调 SendVerificationCode
  → 手机号已注册 → Step 1a（密码登录）
  → 手机号未注册 → Step 1b（验证码注册）

Step 1a: 输入密码 → "登录"
  → 调 Login RPC

Step 1b: 输入验证码 + 设置密码 → "注册"
  → 调 Register RPC
```

### ego_client.dart 新增方法

```dart
Future<void> sendVerificationCode(String phone)
Future<RegisterRes> register(String phone, String code, String password)
// login 参数改为 phone
Future<LoginRes> login(String phone, String password)
```

### 页面状态 (ConsumerStatefulWidget)

```dart
_phoneCtrl, _passwordCtrl, _codeCtrl
_step           // 0=手机号, 1=密码, 2=验证码注册
_loading, _error
_phoneExists     // 后端 SendVerificationCode 返回（或通过独立 check）
```

### 文件改动

| 文件 | 改动 |
|------|------|
| `client/lib/features/login/login_page.dart` | 重写 3 step UI |
| `client/lib/data/services/ego_client.dart` | 新增方法，login 改签名 |
| `client/lib/data/generated/api.pb.dart` | proto 重新生成 |
| `client/lib/data/generated/api.pbgrpc.dart` | proto 重新生成 |

## 数据库 migration

新增 migration SQL：`ALTER TABLE users RENAME COLUMN account TO phone;`

## Error 映射

| 领域错误 | gRPC Status | 前端展示 |
|----------|-------------|----------|
| `ErrInvalidPhone` | `InvalidArgument` | "请输入正确的手机号" |
| `ErrInvalidPassword` | `Unauthenticated` | "密码错误" |
| `ErrPhoneAlreadyRegistered` | `AlreadyExists` | "该手机号已注册" |
| `ErrInvalidVerificationCode` | 映射为身份验证错误 | "验证码错误" |
| `ErrCodeExpired` | 映射为身份验证错误 | "验证码已过期" |
