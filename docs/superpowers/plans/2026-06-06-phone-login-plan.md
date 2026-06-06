# phone-login Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 ego login 从「账号+密码」升级为「手机号+密码」注册/登录，引入阿里云短信认证服务，拆分 Register/Login/SendVerificationCode 三个 RPC。

**Architecture:** Proto 新增 SendVerificationCode + Register RPC；后端 identity 领域新增 SendCode / Register 用例 + SmsService 接口 + 阿里云 adapter；前端 login page 重写为 3 step 流程（手机号输入 → 密码登录 or 验证码注册）。

**Tech Stack:** Flutter (Dart) + Riverpod + gRPC-web, Go + pgx + sqlc, PostgreSQL, 阿里云短信服务

---

## File Structure

```
proto/ego/api.proto                         # 新增 2 个 RPC + 消息定义

server/internal/platform/postgres/
├── migrations/010_phone_login.sql           # ALTER account → phone
├── queries/users.sql                        # account → phone
└── sqlc/users.sql.go                        # 重新生成

server/internal/identity/
├── domain/user.go                           # Account → Phone
├── domain/errors.go                         # 新增 4 个错误
├── app/ports.go                             # 新增 SmsService
├── app/login.go                             # account → phone，去注册逻辑
├── app/send_code.go                         # 新增：SendCode 用例
├── app/register.go                          # 新增：Register 用例
├── adapter/sms/aliyun.go                    # 新增：阿里云 SMS adapter
├── adapter/grpc/handler.go                  # 新增 2 个 handler 方法
├── adapter/postgres/user_repo.go            # Account → Phone
├── adapter/id/uuid.go                       # 不改
└── module.go                                # 新增 smsService 依赖

server/internal/bootstrap/
├── platform.go                              # 新增 SmsService 字段
└── identity.go                              # 注入 smsService

server/internal/config/config.go             # 新增阿里云 SMS 配置

client/lib/data/
├── generated/api.pb.dart                    # 重新生成
├── generated/api.pbgrpc.dart                # 重新生成
└── services/ego_client.dart                 # 新增 2 个方法，login 改签名

client/lib/features/login/login_page.dart    # 重写
```

---

### Task 1: Proto 契约变更

**Files:**
- Modify: `proto/ego/api.proto`

- [ ] **Step 1: 修改 LoginReq/LoginRes，新增消息和 RPC**

在现有 `service Ego` 中新增 2 个 rpc，修改 LoginReq/LoginRes，新增消息定义。

```proto
// === 找到 LoginReq 替换 ===
message LoginReq {
  string phone    = 1;
  string password = 2;
}

message LoginRes {
  string token = 1;
}

// === 新增消息（放在 LoginRes 之后）===
message SendVerificationCodeReq {
  string phone = 1;
}

message SendVerificationCodeRes {
  bool registered = 1;  // true = 已注册，前端展示密码登录；false = 未注册，展示验证码注册
}

message RegisterReq {
  string phone    = 1;
  string code     = 2;  // 短信验证码
  string password = 3;
}

message RegisterRes {
  string token = 1;
}

// === 在 service Ego 中，Login 之后新增 2 个 rpc ===
  rpc SendVerificationCode(SendVerificationCodeReq) returns (SendVerificationCodeRes);
  rpc Register(RegisterReq) returns (RegisterRes);
```

- [ ] **Step 2: 重新生成 Go proto 桩代码**

```bash
cd /home/sunlie/project/ego && make proto
```

预期：`server/proto/ego/api.pb.go` 和 `api_grpc.pb.go` 重新生成。

- [ ] **Step 3: 重新生成 Dart proto 桩代码**

```bash
cd /home/sunlie/project/ego/client && flutter pub global run protoc_plugin protoc-gen-dart
```

预期：`client/lib/data/generated/api.pb.dart` 和 `api.pbgrpc.dart` 重新生成。

- [ ] **Step 4: Commit**

```bash
git add proto/ server/proto/ client/lib/data/generated/ server/internal/
git commit -m "feat(proto): add SendVerificationCode and Register RPCs, update LoginReq/LoginRes for phone-based auth"
```

---

### Task 2: 数据库 Migration — account → phone

**Files:**
- Create: `server/internal/platform/postgres/migrations/010_phone_login.sql`
- Modify: `server/internal/platform/postgres/queries/users.sql`
- 重新生成: `server/internal/platform/postgres/sqlc/users.sql.go`

- [ ] **Step 1: 创建 migration SQL**

```sql
-- server/internal/platform/postgres/migrations/010_phone_login.sql
ALTER TABLE users RENAME COLUMN account TO phone;
ALTER INDEX idx_users_account RENAME TO idx_users_phone;
```

- [ ] **Step 2: 更新 sqlc query 文件**

```sql
-- server/internal/platform/postgres/queries/users.sql
-- name: GetUserByPhone :one
SELECT id, password_hash FROM users WHERE phone = $1;

-- name: CreateUser :exec
INSERT INTO users (id, phone, password_hash, created_at) VALUES ($1, $2, $3, $4);
```

- [ ] **Step 3: 重新生成 sqlc Go 代码**

```bash
cd /home/sunlie/project/ego/server && sqlc generate
```

预期：`server/internal/platform/postgres/sqlc/users.sql.go` 中 `GetUserByAccount` → `GetUserByPhone`，`CreateUserParams.Account` → `Phone`。

- [ ] **Step 4: Commit**

```bash
git add server/internal/platform/postgres/migrations/010_phone_login.sql \
        server/internal/platform/postgres/queries/users.sql \
        server/internal/platform/postgres/sqlc/users.sql.go
git commit -m "feat(db): migrate account column to phone"
```

---

### Task 3: 后端 — domain 层改造

**Files:**
- Modify: `server/internal/identity/domain/user.go`
- Modify: `server/internal/identity/domain/errors.go`

- [ ] **Step 1: 修改 User 模型和 Repository 接口**

```go
// server/internal/identity/domain/user.go
package domain

import (
	"context"
	"time"
)

type User struct {
	ID           string
	Phone        string
	PasswordHash string
	CreatedAt    time.Time
}

type UserRepository interface {
	FindByPhone(ctx context.Context, phone string) (*User, error)
	Create(ctx context.Context, user *User) error
}
```

- [ ] **Step 2: 新增领域错误**

```go
// server/internal/identity/domain/errors.go
package domain

import "errors"

var ErrUserNotFound = errors.New("user not found")
var ErrInvalidPassword = errors.New("invalid password")
var ErrPhoneAlreadyRegistered = errors.New("phone already registered")
var ErrInvalidVerificationCode = errors.New("invalid verification code")
var ErrCodeExpired = errors.New("verification code expired")
var ErrInvalidPhone = errors.New("invalid phone number")
```

- [ ] **Step 3: Commit**

```bash
git add server/internal/identity/domain/
git commit -m "feat(identity): update domain model account→phone, add new error types"
```

---

### Task 4: 后端 — app 层 port 定义和 Login 改造

**Files:**
- Modify: `server/internal/identity/app/ports.go`
- Modify: `server/internal/identity/app/login.go`

- [ ] **Step 1: ports.go 新增 SmsService 接口**

```go
// server/internal/identity/app/ports.go
package app

import "context"

type PasswordHasher interface {
	Hash(plaintext string) (string, error)
	Verify(hash, plaintext string) error
}

type TokenIssuer interface {
	Issue(userID string) (string, error)
}

type IDGenerator interface {
	New() string
}

type SmsService interface {
	Send(ctx context.Context, phone string) error
	Verify(ctx context.Context, phone, code string) (bool, error)
}
```

- [ ] **Step 2: 重写 login.go — account → phone，去注册逻辑**

```go
// server/internal/identity/app/login.go
package app

import (
	"context"
	"errors"
	"fmt"

	"ego-server/internal/identity/domain"
)

type LoginUseCase struct {
	users  domain.UserRepository
	hasher PasswordHasher
	tokens TokenIssuer
}

func NewLoginUseCase(users domain.UserRepository, hasher PasswordHasher, tokens TokenIssuer) *LoginUseCase {
	return &LoginUseCase{users: users, hasher: hasher, tokens: tokens}
}

type LoginResult struct {
	Token string
}

func (uc *LoginUseCase) Login(ctx context.Context, phone, password string) (*LoginResult, error) {
	user, err := uc.users.FindByPhone(ctx, phone)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by phone: %w", err)
	}

	if err := uc.hasher.Verify(user.PasswordHash, password); err != nil {
		return nil, domain.ErrInvalidPassword
	}

	token, err := uc.tokens.Issue(user.ID)
	if err != nil {
		return nil, fmt.Errorf("issue token: %w", err)
	}

	return &LoginResult{Token: token}, nil
}
```

注意：LoginUseCase 不再需要 `ids IDGenerator` 字段。

- [ ] **Step 3: Commit**

```bash
git add server/internal/identity/app/ports.go server/internal/identity/app/login.go
git commit -m "feat(identity): add SmsService port, simplify LoginUseCase to phone-only"
```

---

### Task 5: 后端 — SendCode 用例

**Files:**
- Create: `server/internal/identity/app/send_code.go`

- [ ] **Step 1: 创建 send_code.go**

```go
// server/internal/identity/app/send_code.go
package app

import (
	"context"
	"regexp"
	"strings"

	"ego-server/internal/identity/domain"
)

type SendCodeUseCase struct {
	users     domain.UserRepository
	smsSender SmsService
}

func NewSendCodeUseCase(users domain.UserRepository, smsSender SmsService) *SendCodeUseCase {
	return &SendCodeUseCase{users: users, smsSender: smsSender}
}

type SendCodeResult struct {
	Registered bool
}

var phonePattern = regexp.MustCompile(`^1[3-9]\d{9}$`)

func (uc *SendCodeUseCase) SendCode(ctx context.Context, phone string) (*SendCodeResult, error) {
	phone = strings.TrimSpace(phone)
	if !phonePattern.MatchString(phone) {
		return nil, domain.ErrInvalidPhone
	}

	registered := true
	_, err := uc.users.FindByPhone(ctx, phone)
	if err != nil {
		if isUserNotFound(err) {
			registered = false
		} else {
			return nil, err
		}
	}

	if err := uc.smsSender.Send(ctx, phone); err != nil {
		return nil, err
	}

	return &SendCodeResult{Registered: registered}, nil
}

func isUserNotFound(err error) bool {
	return err != nil && err.Error() == "user not found"
}
```

- [ ] **Step 2: Commit**

```bash
git add server/internal/identity/app/send_code.go
git commit -m "feat(identity): add SendCode use case with phone validation"
```

---

### Task 6: 后端 — Register 用例

**Files:**
- Create: `server/internal/identity/app/register.go`

- [ ] **Step 1: 创建 register.go**

```go
// server/internal/identity/app/register.go
package app

import (
	"context"
	"fmt"

	"ego-server/internal/identity/domain"
)

type RegisterUseCase struct {
	users     domain.UserRepository
	hasher    PasswordHasher
	tokens    TokenIssuer
	ids       IDGenerator
	smsSender SmsService
}

func NewRegisterUseCase(
	users domain.UserRepository,
	hasher PasswordHasher,
	tokens TokenIssuer,
	ids IDGenerator,
	smsSender SmsService,
) *RegisterUseCase {
	return &RegisterUseCase{
		users:     users,
		hasher:    hasher,
		tokens:    tokens,
		ids:       ids,
		smsSender: smsSender,
	}
}

type RegisterResult struct {
	Token string
}

func (uc *RegisterUseCase) Register(ctx context.Context, phone, code, password string) (*RegisterResult, error) {
	if !phonePattern.MatchString(phone) {
		return nil, domain.ErrInvalidPhone
	}
	if len(password) < 6 {
		return nil, fmt.Errorf("password too short")
	}

	// 校验验证码
	ok, err := uc.smsSender.Verify(ctx, phone, code)
	if err != nil {
		return nil, fmt.Errorf("verify code: %w", err)
	}
	if !ok {
		return nil, domain.ErrInvalidVerificationCode
	}

	// 查重
	_, err = uc.users.FindByPhone(ctx, phone)
	if err == nil {
		return nil, domain.ErrPhoneAlreadyRegistered
	}
	if !isUserNotFound(err) {
		return nil, fmt.Errorf("find user by phone: %w", err)
	}

	// 创建用户
	hash, err := uc.hasher.Hash(password)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user := &domain.User{
		ID:           uc.ids.New(),
		Phone:        phone,
		PasswordHash: hash,
	}

	if err := uc.users.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}

	// 签发 JWT
	token, err := uc.tokens.Issue(user.ID)
	if err != nil {
		return nil, fmt.Errorf("issue token: %w", err)
	}

	return &RegisterResult{Token: token}, nil
}
```

- [ ] **Step 2: Commit**

```bash
git add server/internal/identity/app/register.go
git commit -m "feat(identity): add Register use case with SMS verification"
```

---

### Task 7: 后端 — 阿里云 SMS adapter

**Files:**
- Create: `server/internal/identity/adapter/sms/aliyun.go`

- [ ] **Step 1: 创建阿里云 SMS adapter**

```go
// server/internal/identity/adapter/sms/aliyun.go
package sms

import (
	"context"
	"fmt"
)

// AliyunSmsService 通过阿里云短信服务发送和校验验证码。
// 阿里云短信服务内部管理验证码的生成、生命周期和安全校验，
// 本 adapter 仅做转发。
type AliyunSmsService struct {
	// TODO: 根据实际阿里云 SDK 添加 client 字段
	// client *dysmsapi.Client
}

func NewAliyunSmsService() *AliyunSmsService {
	return &AliyunSmsService{}
}

// Send 调用阿里云 API 发送短信验证码。
// 阿里云会自动生成验证码、管理过期时间。
func (s *AliyunSmsService) Send(ctx context.Context, phone string) error {
	// TODO: 集成阿里云 SDK
	// req := &dysmsapi.SendSmsRequest{
	//     PhoneNumbers:  phone,
	//     SignName:      "ego",
	//     TemplateCode:  "SMS_XXXXXXXXX",
	//     TemplateParam: `{"code":"123456"}`,
	// }
	// _, err := s.client.SendSms(req)
	// return err

	// 占位实现：阿里云 SDK 接入时替换
	return fmt.Errorf("aliyun SMS not configured")
}

// Verify 调用阿里云 API 校验验证码。
func (s *AliyunSmsService) Verify(ctx context.Context, phone, code string) (bool, error) {
	// TODO: 集成阿里云 SDK 校验
	// req := &dysmsapi.CheckSmsCodeRequest{...}
	// res, err := s.client.CheckSmsCode(req)
	// return res.Success, err

	return false, fmt.Errorf("aliyun SMS not configured")
}
```

- [ ] **Step 2: Commit**

```bash
git add server/internal/identity/adapter/sms/
git commit -m "feat(identity): add Aliyun SMS adapter skeleton"
```

---

### Task 8: 后端 — postgres adapter 更新

**Files:**
- Modify: `server/internal/identity/adapter/postgres/user_repo.go`

- [ ] **Step 1: Account → Phone，Update method names**

```go
// server/internal/identity/adapter/postgres/user_repo.go
package postgres

import (
	"context"
	"errors"
	"time"

	"ego-server/internal/identity/domain"
	"ego-server/internal/platform/postgres/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserRepository struct {
	queries *sqlc.Queries
}

func NewUserRepository(queries *sqlc.Queries) *UserRepository {
	return &UserRepository{queries: queries}
}

func (r *UserRepository) FindByPhone(ctx context.Context, phone string) (*domain.User, error) {
	row, err := r.queries.GetUserByPhone(ctx, phone)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}

	id, err := uuid.FromBytes(row.ID.Bytes[:])
	if err != nil {
		return nil, err
	}

	return &domain.User{
		ID:           id.String(),
		Phone:        phone,
		PasswordHash: row.PasswordHash,
	}, nil
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	uid, err := uuid.Parse(user.ID)
	if err != nil {
		return err
	}

	now := time.Now()
	user.CreatedAt = now

	return r.queries.CreateUser(ctx, sqlc.CreateUserParams{
		ID:           pgtype.UUID{Bytes: [16]byte(uid), Valid: true},
		Phone:        user.Phone,
		PasswordHash: user.PasswordHash,
		CreatedAt:    pgtype.Timestamptz{Time: now, Valid: true},
	})
}
```

- [ ] **Step 2: Commit**

```bash
git add server/internal/identity/adapter/postgres/user_repo.go
git commit -m "refactor(identity): update user_repo Account→Phone"
```

---

### Task 9: 后端 — gRPC Handler 更新

**Files:**
- Modify: `server/internal/identity/adapter/grpc/handler.go`

- [ ] **Step 1: 重写 handler.go — 新增 3 个 handler 方法**

```go
// server/internal/identity/adapter/grpc/handler.go
package identitygrpc

import (
	"context"
	"errors"

	"ego-server/internal/identity/app"
	"ego-server/internal/identity/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "ego-server/proto/ego"
)

type Handler struct {
	pb.UnimplementedEgoServer
	login    *app.LoginUseCase
	register *app.RegisterUseCase
	sendCode *app.SendCodeUseCase
}

func NewHandler(
	login *app.LoginUseCase,
	register *app.RegisterUseCase,
	sendCode *app.SendCodeUseCase,
) *Handler {
	return &Handler{login: login, register: register, sendCode: sendCode}
}

func (h *Handler) SendVerificationCode(ctx context.Context, req *pb.SendVerificationCodeReq) (*pb.SendVerificationCodeRes, error) {
	result, err := h.sendCode.SendCode(ctx, req.Phone)
	if err != nil {
		return nil, mapError(err)
	}
	return &pb.SendVerificationCodeRes{Registered: result.Registered}, nil
}

func (h *Handler) Register(ctx context.Context, req *pb.RegisterReq) (*pb.RegisterRes, error) {
	result, err := h.register.Register(ctx, req.Phone, req.Code, req.Password)
	if err != nil {
		return nil, mapError(err)
	}
	return &pb.RegisterRes{Token: result.Token}, nil
}

func (h *Handler) Login(ctx context.Context, req *pb.LoginReq) (*pb.LoginRes, error) {
	result, err := h.login.Login(ctx, req.Phone, req.Password)
	if err != nil {
		return nil, mapError(err)
	}
	return &pb.LoginRes{Token: result.Token}, nil
}

func mapError(err error) error {
	if errors.Is(err, domain.ErrInvalidPassword) {
		return status.Error(codes.Unauthenticated, "密码错误")
	}
	if errors.Is(err, domain.ErrUserNotFound) {
		return status.Error(codes.NotFound, "用户不存在")
	}
	if errors.Is(err, domain.ErrInvalidPhone) {
		return status.Error(codes.InvalidArgument, "请输入正确的手机号")
	}
	if errors.Is(err, domain.ErrInvalidVerificationCode) {
		return status.Error(codes.Unauthenticated, "验证码错误")
	}
	if errors.Is(err, domain.ErrCodeExpired) {
		return status.Error(codes.Unauthenticated, "验证码已过期")
	}
	if errors.Is(err, domain.ErrPhoneAlreadyRegistered) {
		return status.Error(codes.AlreadyExists, "该手机号已注册")
	}
	return status.Error(codes.Internal, err.Error())
}
```

- [ ] **Step 2: Commit**

```bash
git add server/internal/identity/adapter/grpc/handler.go
git commit -m "feat(identity): update gRPC handler with SendCode/Register/Login methods"
```

---

### Task 10: 后端 — module.go + bootstrap wiring

**Files:**
- Modify: `server/internal/identity/module.go`
- Modify: `server/internal/bootstrap/identity.go`
- Modify: `server/internal/bootstrap/platform.go`
- Modify: `server/internal/config/config.go`

- [ ] **Step 1: 更新 module.go**

```go
// server/internal/identity/module.go
package identity

import (
	identitygrpc "ego-server/internal/identity/adapter/grpc"
	identityid "ego-server/internal/identity/adapter/id"
	identitypostgres "ego-server/internal/identity/adapter/postgres"
	identityapp "ego-server/internal/identity/app"
	"ego-server/internal/platform/postgres/sqlc"
)

type Deps struct {
	DB        sqlc.DBTX
	Hasher    identityapp.PasswordHasher
	Tokens    identityapp.TokenIssuer
	SmsSender identityapp.SmsService
}

func NewHandler(deps Deps) *identitygrpc.Handler {
	queries := sqlc.New(deps.DB)

	userRepo := identitypostgres.NewUserRepository(queries)
	ids := identityid.NewUUIDGenerator()

	loginUseCase := identityapp.NewLoginUseCase(userRepo, deps.Hasher, deps.Tokens)
	registerUseCase := identityapp.NewRegisterUseCase(userRepo, deps.Hasher, deps.Tokens, ids, deps.SmsSender)
	sendCodeUseCase := identityapp.NewSendCodeUseCase(userRepo, deps.SmsSender)

	return identitygrpc.NewHandler(loginUseCase, registerUseCase, sendCodeUseCase)
}
```

- [ ] **Step 2: 更新 bootstrap/identity.go**

```go
// server/internal/bootstrap/identity.go
package bootstrap

import (
	"ego-server/internal/identity"

	pb "ego-server/proto/ego"
)

func NewIdentityHandler(p *Platform) pb.EgoServer {
	return identity.NewHandler(identity.Deps{
		DB:        p.Pool,
		Hasher:    p.Hasher,
		Tokens:    p.Tokens,
		SmsSender: p.SmsService,
	})
}
```

- [ ] **Step 3: 更新 bootstrap/platform.go — 新增 SmsService 字段**

在 `Platform` struct 中新增字段：

```go
type Platform struct {
	Pool       *pgxpool.Pool
	JWTKey     []byte
	JWTExp     time.Duration
	Hasher     auth.BcryptHasher
	Tokens     auth.JWTIssuer
	Logger     *slog.Logger
	AIClient   *ai.Client
	SmsService sms.AliyunSmsService  // 新增
}
```

在 `InitPlatform` 函数末尾 return 前初始化：

```go
// 阿里云 SMS 目前使用空实现，集成时替换
// smsClient := sms.NewAliyunSmsService(cfg.AliyunAccessKeyID, cfg.AliyunAccessKeySecret)
smsClient := sms.NewAliyunSmsService()
```

return Platform 中添加 `SmsService: smsClient`。

需要添加 import:
```go
import (
	// ... existing imports
	sms "ego-server/internal/identity/adapter/sms"
)
```

- [ ] **Step 4: 更新 config.go — 新增阿里云配置字段**

在 `Config` struct 中新增（暂不实际使用，为集成做准备）：

```go
AliyunAccessKeyID     string
AliyunAccessKeySecret string
AliyunSmsSignName     string
AliyunSmsTemplateCode string
```

在 `Load()` 中添加：

```go
AliyunAccessKeyID:     os.Getenv("ALIYUN_ACCESS_KEY_ID"),
AliyunAccessKeySecret: os.Getenv("ALIYUN_ACCESS_KEY_SECRET"),
AliyunSmsSignName:     os.Getenv("ALIYUN_SMS_SIGN_NAME"),
AliyunSmsTemplateCode: os.Getenv("ALIYUN_SMS_TEMPLATE_CODE"),
```

- [ ] **Step 5: 编译验证**

```bash
cd /home/sunlie/project/ego/server && go build ./...
```

- [ ] **Step 6: Commit**

```bash
git add server/internal/identity/module.go \
        server/internal/bootstrap/ \
        server/internal/config/config.go
git commit -m "feat(identity): wire SmsService into module and bootstrap"
```

---

### Task 11: 前端 — ego_client.dart 更新

**Files:**
- Modify: `client/lib/data/services/ego_client.dart`

- [ ] **Step 1: 新增 2 个方法，login 改签名**

```dart
// 替换 Auth 部分（login 方法）并新增 2 个方法

  // ─── Auth ─────────────────────────────────────

  Future<grpc.SendVerificationCodeRes> sendVerificationCode(String phone) async {
    final req = grpc.SendVerificationCodeReq(phone: phone);
    return _stub.sendVerificationCode(req);
  }

  Future<grpc.RegisterRes> register({
    required String phone,
    required String code,
    required String password,
  }) async {
    final req = grpc.RegisterReq(phone: phone, code: code, password: password);
    return _stub.register(req);
  }

  Future<grpc.LoginRes> login(String phone, String password) async {
    final req = grpc.LoginReq(phone: phone, password: password);
    return _stub.login(req);
  }
```

注意：`sendVerificationCode` 和 `register` 不需要 auth options（注册时还没有 token）。`login` 也不需要 auth options（登录获取 token）。

- [ ] **Step 2: Commit**

```bash
git add client/lib/data/services/ego_client.dart
git commit -m "feat(client): add sendVerificationCode/register methods, update login to phone"
```

---

### Task 12: 前端 — login_page.dart 重写

**Files:**
- Modify: `client/lib/features/login/login_page.dart`

- [ ] **Step 1: 重写 login_page.dart — 3 step 流程**

```dart
// client/lib/features/login/login_page.dart
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../core/providers/auth_provider.dart';
import '../../data/services/ego_client.dart';

class LoginPage extends ConsumerStatefulWidget {
  const LoginPage({super.key});

  @override
  ConsumerState<LoginPage> createState() => _LoginPageState();
}

class _LoginPageState extends ConsumerState<LoginPage> {
  final _phoneCtrl = TextEditingController();
  final _passwordCtrl = TextEditingController();
  final _codeCtrl = TextEditingController();
  int _step = 0; // 0=手机号, 1=密码登录, 2=验证码注册
  bool _loading = false;
  String? _error;
  int _countdown = 0;

  @override
  void initState() {
    super.initState();
  }

  @override
  void dispose() {
    _phoneCtrl.dispose();
    _passwordCtrl.dispose();
    _codeCtrl.dispose();
    super.dispose();
  }

  void _setError(String? msg) {
    if (mounted) setState(() => _error = msg);
  }

  Future<void> _sendCode() async {
    final phone = _phoneCtrl.text.trim();
    if (phone.isEmpty) {
      _setError('请输入手机号');
      return;
    }
    if (!RegExp(r'^1[3-9]\d{9}$').hasMatch(phone)) {
      _setError('请输入正确的手机号');
      return;
    }

    setState(() { _loading = true; _error = null; });

    try {
      final client = ref.read(EgoClient.provider);
      final res = await client.sendVerificationCode(phone);

      if (mounted) {
        setState(() {
          _loading = false;
          _step = res.registered ? 1 : 2;
          _countdown = 60;
        });
        _startCountdown();
      }
    } catch (e) {
      if (mounted) {
        setState(() => _loading = false);
        _setError('发送验证码失败，请稍后重试');
      }
    }
  }

  void _startCountdown() {
    Future.delayed(const Duration(seconds: 1), () {
      if (mounted && _countdown > 0) {
        setState(() => _countdown--);
        _startCountdown();
      }
    });
  }

  Future<void> _login() async {
    final phone = _phoneCtrl.text.trim();
    final password = _passwordCtrl.text;

    if (password.isEmpty) {
      _setError('请输入密码');
      return;
    }

    setState(() { _loading = true; _error = null; });

    try {
      final client = ref.read(EgoClient.provider);
      final res = await client.login(phone, password);
      if (mounted) {
        ref.read(authProvider.notifier).login(res.token);
      }
    } catch (e) {
      if (mounted) {
        setState(() => _loading = false);
        _setError(e.toString().contains('密码错误')
            ? '密码错误'
            : e.toString().contains('用户不存在')
                ? '用户不存在'
                : '登录失败，请稍后重试');
      }
    }
  }

  Future<void> _register() async {
    final phone = _phoneCtrl.text.trim();
    final code = _codeCtrl.text.trim();
    final password = _passwordCtrl.text;

    if (code.isEmpty) {
      _setError('请输入验证码');
      return;
    }
    if (password.length < 6) {
      _setError('密码至少 6 位');
      return;
    }

    setState(() { _loading = true; _error = null; });

    try {
      final client = ref.read(EgoClient.provider);
      final res = await client.register(phone: phone, code: code, password: password);
      if (mounted) {
        ref.read(authProvider.notifier).login(res.token);
      }
    } catch (e) {
      if (mounted) {
        setState(() => _loading = false);
        final msg = e.toString();
        if (msg.contains('验证码')) {
          _setError('验证码错误');
        } else if (msg.contains('已注册')) {
          _setError('该手机号已注册，请返回登录');
        } else {
          _setError('注册失败，请稍后重试');
        }
      }
    }
  }

  void _backToStep0() {
    setState(() {
      _step = 0;
      _error = null;
      _countdown = 0;
      _passwordCtrl.clear();
      _codeCtrl.clear();
    });
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: SafeArea(
        child: Center(
          child: SingleChildScrollView(
            padding: const EdgeInsets.symmetric(horizontal: 32),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                // Logo
                Container(
                  width: 120,
                  height: 120,
                  decoration: const BoxDecoration(shape: BoxShape.circle),
                  clipBehavior: Clip.antiAlias,
                  child: Image.asset('ego-logo.webp', fit: BoxFit.cover),
                ),
                const SizedBox(height: 48),

                // Step indicator
                if (_step != 0)
                  Padding(
                    padding: const EdgeInsets.only(bottom: 16),
                    child: Text(
                      _step == 1 ? '密码登录' : '创建账号',
                      style: const TextStyle(
                        fontSize: 16,
                        color: Color(0xFFCCA880),
                        fontWeight: FontWeight.w300,
                        letterSpacing: 2,
                      ),
                    ),
                  ),

                // Phone display (step 1/2)
                if (_step != 0) ...[
                  Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Text(
                        _phoneCtrl.text.trim(),
                        style: const TextStyle(fontSize: 14, color: Color(0xFFA0A0B8)),
                      ),
                      const SizedBox(width: 8),
                      GestureDetector(
                        onTap: _loading ? null : _backToStep0,
                        child: const Text(
                          '修改',
                          style: TextStyle(fontSize: 12, color: Color(0xFFCCA880)),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 24),
                ],

                // Step 0: Phone input
                if (_step == 0)
                  TextField(
                    controller: _phoneCtrl,
                    keyboardType: TextInputType.phone,
                    inputFormatters: [
                      FilteringTextInputFormatter.digitsOnly,
                      LengthLimitingTextInputFormatter(11),
                    ],
                    decoration: const InputDecoration(
                      hintText: '请输入手机号',
                      prefixIcon: Icon(Icons.phone_android_outlined),
                    ),
                    textInputAction: TextInputAction.done,
                    onSubmitted: (_) => _sendCode(),
                  ),

                // Step 1: Password
                if (_step == 1)
                  TextField(
                    controller: _passwordCtrl,
                    obscureText: true,
                    decoration: const InputDecoration(
                      hintText: '请输入密码',
                      prefixIcon: Icon(Icons.lock_outline),
                    ),
                    textInputAction: TextInputAction.done,
                    onSubmitted: (_) => _login(),
                  ),

                // Step 2: Code + Password
                if (_step == 2) ...[
                  Row(
                    children: [
                      Expanded(
                        child: TextField(
                          controller: _codeCtrl,
                          keyboardType: TextInputType.number,
                          inputFormatters: [
                            FilteringTextInputFormatter.digitsOnly,
                            LengthLimitingTextInputFormatter(6),
                          ],
                          decoration: const InputDecoration(
                            hintText: '验证码',
                            prefixIcon: Icon(Icons.sms_outlined),
                          ),
                        ),
                      ),
                      const SizedBox(width: 12),
                      SizedBox(
                        width: 100,
                        child: TextButton(
                          onPressed: _countdown > 0 || _loading ? null : _sendCode,
                          child: Text(
                            _countdown > 0 ? '${_countdown}s' : '重新发送',
                            style: const TextStyle(fontSize: 12),
                          ),
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 16),
                  TextField(
                    controller: _passwordCtrl,
                    obscureText: true,
                    decoration: const InputDecoration(
                      hintText: '设置密码（至少 6 位）',
                      prefixIcon: Icon(Icons.lock_outline),
                    ),
                    textInputAction: TextInputAction.done,
                    onSubmitted: (_) => _register(),
                  ),
                ],

                // Error
                if (_error != null) ...[
                  const SizedBox(height: 16),
                  Text(_error!, style: const TextStyle(color: Colors.redAccent)),
                ],

                const SizedBox(height: 32),

                // Action button
                if (_step == 0)
                  ElevatedButton(
                    onPressed: _loading ? null : _sendCode,
                    child: _loading
                        ? const SizedBox(
                            height: 20, width: 20,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : const Text('下一步', style: TextStyle(fontSize: 16)),
                  ),
                if (_step == 1)
                  ElevatedButton(
                    onPressed: _loading ? null : _login,
                    child: _loading
                        ? const SizedBox(
                            height: 20, width: 20,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : const Text('登录', style: TextStyle(fontSize: 16)),
                  ),
                if (_step == 2)
                  ElevatedButton(
                    onPressed: _loading ? null : _register,
                    child: _loading
                        ? const SizedBox(
                            height: 20, width: 20,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : const Text('注册', style: TextStyle(fontSize: 16)),
                  ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}
```

- [ ] **Step 2: Flutter analyze 验证**

```bash
cd /home/sunlie/project/ego/client && flutter analyze lib/features/login/
```

- [ ] **Step 3: Commit**

```bash
git add client/lib/features/login/login_page.dart
git commit -m "feat(client): rewrite login page with 3-step phone auth flow"
```

---

### Task 13: 联调验证

- [ ] **Step 1: 启动本地环境运行 clean-start.sh**

```bash
bash clean-start.sh
```

- [ ] **Step 2: 手动测试流程**

1. 打开 http://localhost:9081
2. 输入未注册手机号 → 点击下一步
3. 验证进入注册界面（验证码 + 密码）
4. 返回，输入已注册手机号 → 验证进入登录界面（密码）
5. 密码错误 → 验证错误提示

预期：前端 UI 流程和路由跳转正常（因阿里云 SMS 还未集成，发送验证码会报错，UI 会展示错误提示，这是预期行为）。

- [ ] **Step 3: Commit**

```bash
git add docs/superpowers/specs/2026-06-06-phone-login-design.md docs/superpowers/plans/
git commit -m "docs: add phone-login spec and implementation plan"
```
