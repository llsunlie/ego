# ego-login server

后端 identity 领域 context — 手机号认证。

## 所属 gRPC 方法（6 个，均为免认证）

| RPC | 功能 | 副作用 |
|-----|------|--------|
| `CheckPhone` | 查询手机号是否已注册 | 无 |
| `SendVerificationCode` | 发送短信验证码（阿里云） | SMS |
| `Register` | 验证码校验 + 创建用户 + 签发双 JWT | 写库 |
| `Login` | 手机号+密码登录 + 签发双 JWT | 无 |
| `ResetPassword` | 验证码校验 + 更新密码 + 签发双 JWT | 写库 |
| `RefreshToken` | 验证 refresh token + 签发新 access token | 无 |

## 模块结构 (`server/internal/identity/`)

```
identity/
├── module.go                          # 依赖注入
├── domain/
│   ├── user.go                        # User{ID, Phone, PasswordHash}
│   └── errors.go                      # ErrUserNotFound, ErrInvalidPassword, ErrPhoneAlreadyRegistered, etc.
├── app/
│   ├── check_phone.go                 # CheckPhone 用例：校验格式 + 查库
│   ├── send_code.go                   # SendCode 用例：校验格式 + 调 SMS
│   ├── register.go                    # Register 用例：验码 + 查重 + bcrypt + 创建 + JWT
│   ├── login.go                       # Login 用例：查库 + 验密 + JWT
│   ├── reset_password.go              # ResetPassword 用例：验码 + 查库 + 更新密码 + 双 JWT
│   ├── refresh_token.go               # RefreshToken 用例：验证 refresh token + 签发新 access token
│   └── ports.go                       # PasswordHasher, TokenIssuer, IDGenerator, SmsService, RefreshTokenVerifier
└── adapter/
    ├── grpc/handler.go                # 4 个 handler 方法 + mapError
    ├── postgres/user_repo.go          # FindByPhone, Create
    ├── sms/aliyun.go                  # AliyunSmsService: SendSmsVerifyCode / CheckSmsVerifyCode
    └── id/uuid.go                     # UUID 生成
```

## 用例详解

### CheckPhone (`app/check_phone.go`)

```
1. trim + 正则校验 (1[3-9]xxxxxxxxx)
2. FindByPhone → 返回 registered=true/false
3. 无副作用
```

### SendCode (`app/send_code.go`)

```
1. trim + 正则校验
2. smsSender.Send(phone) → 阿里云 SendSmsVerifyCode API
   - CodeLength=6, ValidTime=300s, Interval=60s
```

### Register (`app/register.go`)

```
1. 正则校验手机号
2. 密码 >= 6 位
3. smsSender.Verify(phone, code) → 阿里云 CheckSmsVerifyCode API
4. FindByPhone 查重 → ErrPhoneAlreadyRegistered
5. bcrypt 哈希密码
6. Create User
7. Issue access JWT (1h) + refresh JWT (30d)
```

### Login (`app/login.go`)

```
1. FindByPhone → ErrUserNotFound
2. hasher.Verify(passwordHash, password) → ErrInvalidPassword
3. Issue access JWT (1h) + refresh JWT (30d)
```

### ResetPassword (`app/reset_password.go`)

```
1. 正则校验手机号 (1[3-9]xxxxxxxxx) → ErrInvalidPhone
2. 新密码 >= 6 位
3. smsSender.Verify(phone, code) → ErrInvalidVerificationCode
4. userRepo.FindByPhone(phone) → ErrUserNotFound
5. hasher.Hash(newPassword) → bcrypt hash
6. userRepo.UpdatePassword(userID, hash)
7. Issue access JWT (1h) + refresh JWT (30d)
```

### RefreshToken (`app/refresh_token.go`)

```
1. Verifier.Verify(refreshToken, "refresh") → 验证 token_type + 过期
2. 失败 → ErrInvalidRefreshToken
3. tokens.Issue(userID) → 签发新 access JWT (1h)
```

> **注意**：阿里云 SDK 对验证码错误/过期/已使用返回 SDKError（isv.*），
> `adapter/sms/aliyun.go` 的 Verify 方法已将其转为 `(false, nil)`，确保
> `ErrInvalidVerificationCode` 能被正确触发。

## 阿里云短信 (`adapter/sms/aliyun.go`)

```go
type AliyunSmsService struct {
    client       *dypnsapi.Client
    signName     string  // ALIYUN_SMS_SIGN_NAME
    templateCode string  // ALIYUN_SMS_TEMPLATE_CODE
    codeLength   int     // ALIYUN_SMS_CODE_LENGTH (default 6)
    validTime    int     // ALIYUN_SMS_VALID_TIME (default 300)
    interval     int     // ALIYUN_SMS_INTERVAL (default 60)
}

Send(ctx, phone)   → SendSmsVerifyCode API
Verify(ctx, phone, code) → CheckSmsVerifyCode API → "PASS" / "UNKNOWN"
```

## 认证拦截器 (`platform/auth/interceptor.go`)

```go
var PreAuthMethods = map[string]bool{
    "Login":                 true,
    "CheckPhone":            true,
    "SendVerificationCode":  true,
    "Register":              true,
    "ResetPassword":         true,
    "RefreshToken":          true,
}
```

这 6 个 RPC 不校验 access token。其中 RefreshToken 以 refresh token 自身作为凭证。所有鉴权 RPC 会验证 `token_type: "access"` 防止 refresh token 被滥用。

## 模块组装 (`module.go`)

```go
type Deps struct {
    DB        sqlc.DBTX
    Hasher    identityapp.PasswordHasher    // platform/auth/bcrypt
    Tokens    identityapp.TokenIssuer       // platform/auth/jwt
    SmsSender identityapp.SmsService        // adapter/sms/aliyun
}

NewHandler → CheckPhone + SendCode + Register + Login → Handler
```

## Composite 路由 (`bootstrap/composite.go`)

```go
EgoHandler.CheckPhone            → identity.CheckPhone
EgoHandler.SendVerificationCode  → identity.SendVerificationCode
EgoHandler.Register              → identity.Register
EgoHandler.Login                 → identity.Login
EgoHandler.ResetPassword          → identity.ResetPassword
EgoHandler.RefreshToken           → identity.RefreshToken
```

## 数据库

```sql
users (id UUID PK, phone VARCHAR(100) UNIQUE, password_hash VARCHAR(255), created_at TIMESTAMPTZ)
```

sqlc queries: `GetUserByPhone`, `CreateUser`, `UpdateUserPassword`
