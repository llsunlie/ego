# ego-login server

后端 identity 领域 context — 手机号认证。

## 所属 gRPC 方法（4 个，均为免认证）

| RPC | 功能 | 副作用 |
|-----|------|--------|
| `CheckPhone` | 查询手机号是否已注册 | 无 |
| `SendVerificationCode` | 发送短信验证码（阿里云） | SMS |
| `Register` | 验证码校验 + 创建用户 + 签发 JWT | 写库 |
| `Login` | 手机号+密码登录 + 签发 JWT | 无 |

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
│   └── ports.go                       # PasswordHasher, TokenIssuer, IDGenerator, SmsService
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
7. Issue JWT
```

### Login (`app/login.go`)

```
1. FindByPhone → ErrUserNotFound
2. hasher.Verify(passwordHash, password) → ErrInvalidPassword
3. Issue JWT
```

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
var preAuthMethods = map[string]bool{
    "Login":                 true,
    "CheckPhone":            true,
    "SendVerificationCode":  true,
    "Register":              true,
}
```

这 4 个 RPC 不校验 JWT，直接放行。

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
```

## 数据库

```sql
users (id UUID PK, phone VARCHAR(100) UNIQUE, password_hash VARCHAR(255), created_at TIMESTAMPTZ)
```

sqlc queries: `GetUserByPhone`, `CreateUser`
