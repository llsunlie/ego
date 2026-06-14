# JWT Refresh Token 机制实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 从单一 JWT token 升级为 access(1h) + refresh(30d) 双 token 体系，新增 RefreshToken RPC

**Architecture:** Proto 新增 RefreshToken RPC 并重命名 token→access_token+refresh_token；后端 identity domain 新增 RefreshToken 用例，jwt_issuer 支持签发双类型 token，interceptor 将 RefreshToken 加入免鉴权名单；前端 ego_client 捕获 UNAUTHENTICATED 自动刷新

**Tech Stack:** Go gRPC proto(golang-jwt) + Flutter Dart grpc-dart

---

## 文件变更清单

| 操作 | 文件 | 说明 |
|------|------|------|
| 修改 | `proto/ego/api.proto` | 新增 RefreshToken RPC，重命名 token→access_token+refresh_token |
| 修改 | `server/.env.example` | JWT_EXP_HOURS → JWT_ACCESS_EXP_HOURS + JWT_REFRESH_EXP_DAYS |
| 修改 | `server/internal/config/config.go` | Config 结构体 + Load() |
| 修改 | `server/internal/platform/auth/jwt.go` | token_type claim + ParseJWTWithType |
| 修改 | `server/internal/platform/auth/jwt_issuer.go` | 新增 RefreshExp + IssueRefresh 方法 |
| 修改 | `server/internal/platform/auth/interceptor.go` | RefreshToken 加入 PreAuthMethods |
| 修改 | `server/internal/identity/domain/errors.go` | 新增 ErrInvalidRefreshToken |
| 修改 | `server/internal/identity/app/ports.go` | TokenIssuer 新增 IssueRefresh |
| 修改 | `server/internal/identity/app/login.go` | LoginResult 改为双 token |
| 修改 | `server/internal/identity/app/register.go` | RegisterResult 改为双 token |
| 修改 | `server/internal/identity/app/reset_password.go` | ResetPasswordResult 改为双 token |
| 新增 | `server/internal/identity/app/refresh_token.go` | RefreshToken 用例 |
| 修改 | `server/internal/identity/adapter/grpc/handler.go` | 新增 RefreshToken handler + 双 token 返回 |
| 修改 | `server/internal/identity/module.go` | 组装 RefreshToken 用例 |
| 修改 | `server/internal/bootstrap/platform.go` | 解析新配置 + 双过期时间注入 |
| 修改 | `server/internal/bootstrap/composite.go` | 路由 RefreshToken RPC |
| 生成 | `server/proto/ego/` | `make proto-go` 产物 |
| 生成 | `client/lib/data/generated/` | `make proto-dart` 产物 |
| 修改 | `client/lib/data/repositories/local_store.dart` | 新增 refresh token 存取 |
| 修改 | `client/lib/core/providers/auth_provider.dart` | 双 token 状态 + refreshAccessToken |
| 修改 | `client/lib/data/services/ego_client.dart` | refreshToken() + 自动刷新包装器 |
| 修改 | `client/lib/features/login/login_page.dart` | 使用新的 accessToken/refreshToken |

---

### Task 1: Proto 契约变更

**Files:**
- Modify: `proto/ego/api.proto`

- [ ] **Step 1: 修改 LoginRes / RegisterRes / ResetPasswordRes**

将三个 response message 中的 `token` 字段改为 `access_token` 并新增 `refresh_token`：

```protobuf
message LoginRes {
  string access_token = 1;
  string refresh_token = 2;
}

message RegisterRes {
  string access_token = 1;
  string refresh_token = 2;
}

message ResetPasswordRes {
  string access_token = 1;
  string refresh_token = 2;
}
```

- [ ] **Step 2: 新增 RefreshToken RPC 及 message**

在 proto 的 Auth 区块（Login/Register 之后）新增：

```protobuf
rpc RefreshToken(RefreshTokenReq) returns (RefreshTokenRes);

message RefreshTokenReq {
  string refresh_token = 1;
}

message RefreshTokenRes {
  string access_token = 1;
}
```

- [ ] **Step 3: 生成代码**

```bash
make proto-go proto-dart
```

验证生成产物：
- `server/proto/ego/api.pb.go` — LoginRes 含 `AccessToken` + `RefreshToken` 字段
- `client/lib/data/generated/api.pb.dart` — LoginRes 含 `accessToken` + `refreshToken` getter
- `client/lib/data/generated/api.pbgrpc.dart` — EgoClient 含 `refreshToken` 方法

---

### Task 2: 配置变更

**Files:**
- Modify: `server/.env.example`
- Modify: `server/internal/config/config.go`

- [ ] **Step 1: 更新 Config 结构体**

```go
// config.go — Config struct 内
// 删除: JWTExpHours string
// 新增:
JwtAccessExpHours  string
JwtRefreshExpDays  string
```

- [ ] **Step 2: 更新 Load() 函数**

```go
// config.go — Load() 返回值内
// 删除: JWTExpHours: os.Getenv("JWT_EXP_HOURS"),
// 新增:
JwtAccessExpHours: os.Getenv("JWT_ACCESS_EXP_HOURS"),
JwtRefreshExpDays: os.Getenv("JWT_REFRESH_EXP_DAYS"),
```

- [ ] **Step 3: 更新 .env.example**

```bash
# 删除:
JWT_EXP_HOURS=720

# 新增:
JWT_ACCESS_EXP_HOURS=1     # Access token 过期时间（小时）
JWT_REFRESH_EXP_DAYS=30    # Refresh token 过期时间（天）
```

---

### Task 3: Auth 平台层 — JWT 签发/验证扩展

**Files:**
- Modify: `server/internal/platform/auth/jwt.go`
- Modify: `server/internal/platform/auth/jwt_issuer.go`
- Modify: `server/internal/platform/auth/interceptor.go`

- [ ] **Step 1: jwt.go — 新增 ParseJWTWithType**

在 `ParseJWT` 函数之后追加：

```go
// ParseJWTWithType parses and validates a JWT, additionally checking that the
// token_type claim matches expectedType. This prevents access tokens from
// being used as refresh tokens and vice versa.
func ParseJWTWithType(tokenStr string, secret []byte, expectedType string) (string, error) {
	userID, err := ParseJWT(tokenStr, secret)
	if err != nil {
		return "", err
	}

	token, _, err := new(jwt.Parser).ParseUnverified(tokenStr, jwt.MapClaims{})
	if err != nil {
		return "", fmt.Errorf("parse unverified: %w", err)
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid claims")
	}
	tokenType, _ := claims["token_type"].(string)
	if tokenType != expectedType {
		return "", fmt.Errorf("unexpected token type: %q, expected %q", tokenType, expectedType)
	}

	return userID, nil
}
```

- [ ] **Step 2: jwt.go — GenerateJWT 增加 token_type claim**

修改 `GenerateJWT` 函数签名，新增 `tokenType` 参数：

```go
func GenerateJWT(userID string, secret []byte, expiration time.Duration, tokenType string) (string, error) {
	claims := jwt.MapClaims{
		"user_id":    userID,
		"token_type": tokenType,
		"iat":        time.Now().Unix(),
		"exp":        time.Now().Add(expiration).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(secret)
}
```

- [ ] **Step 3: jwt_issuer.go — 新增 IssueRefresh**

修改 `JWTIssuer` 结构体，新增 `RefreshExp` 字段和 `IssueRefresh` 方法：

```go
type JWTIssuer struct {
	Secret     []byte
	Exp        time.Duration
	RefreshExp time.Duration
}

func (i JWTIssuer) Issue(userID string) (string, error) {
	return GenerateJWT(userID, i.Secret, i.Exp, "access")
}

func (i JWTIssuer) IssueRefresh(userID string) (string, error) {
	return GenerateJWT(userID, i.Secret, i.RefreshExp, "refresh")
}
```

- [ ] **Step 4: interceptor.go — 鉴权拦截器验证 token_type + RefreshToken 加入免鉴权名单**

将 `ParseJWT` 调用改为 `ParseJWTWithType`，确保 refresh token 不能直接用于 API 鉴权：

```go
// 修改 interceptor 中的 token 解析:
userID, err := ParseJWTWithType(tokenStr, jwtSecret, "access")
if err != nil {
    return nil, status.Error(codes.Unauthenticated, "invalid token")
}
```

同时将 `RefreshToken` 加入免鉴权名单：
```go
var PreAuthMethods = map[string]bool{
    "Login":                true,
    "CheckPhone":           true,
    "SendVerificationCode": true,
    "Register":             true,
    "ResetPassword":        true,
    "RefreshToken":         true,  // 新增
}
```

```go
var PreAuthMethods = map[string]bool{
	"Login":                true,
	"CheckPhone":           true,
	"SendVerificationCode": true,
	"Register":             true,
	"ResetPassword":        true,
	"RefreshToken":         true,  // 新增
}
```

---

### Task 4: Identity 领域 — domain 层

**Files:**
- Modify: `server/internal/identity/domain/errors.go`
- Modify: `server/internal/identity/app/ports.go`

- [ ] **Step 1: 新增 ErrInvalidRefreshToken**

```go
// domain/errors.go 末尾追加:
var ErrInvalidRefreshToken = errors.New("invalid refresh token")
```

- [ ] **Step 2: TokenIssuer 接口扩展**

```go
// app/ports.go — TokenIssuer interface 内新增:
IssueRefresh(userID string) (string, error)
```

---

### Task 5: Identity 领域 — app 层

**Files:**
- Modify: `server/internal/identity/app/login.go`
- Modify: `server/internal/identity/app/register.go`
- Modify: `server/internal/identity/app/reset_password.go`
- Create: `server/internal/identity/app/refresh_token.go`

- [ ] **Step 1: login.go — LoginResult 改为双 token**

修改 `LoginResult` 结构体，`Login` 方法同时签发两个 token：

```go
type LoginResult struct {
	AccessToken  string
	RefreshToken string
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

	accessToken, err := uc.tokens.Issue(user.ID)
	if err != nil {
		return nil, fmt.Errorf("issue access token: %w", err)
	}

	refreshToken, err := uc.tokens.IssueRefresh(user.ID)
	if err != nil {
		return nil, fmt.Errorf("issue refresh token: %w", err)
	}

	return &LoginResult{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}
```

- [ ] **Step 2: register.go — RegisterResult 改为双 token**

```go
type RegisterResult struct {
	AccessToken  string
	RefreshToken string
}

// Register 方法末尾改为:
accessToken, err := uc.tokens.Issue(user.ID)
if err != nil {
	return nil, fmt.Errorf("issue access token: %w", err)
}

refreshToken, err := uc.tokens.IssueRefresh(user.ID)
if err != nil {
	return nil, fmt.Errorf("issue refresh token: %w", err)
}

return &RegisterResult{AccessToken: accessToken, RefreshToken: refreshToken}, nil
```

- [ ] **Step 3: reset_password.go — ResetPasswordResult 改为双 token**

```go
type ResetPasswordResult struct {
	AccessToken  string
	RefreshToken string
}

// ResetPassword 方法末尾改为:
accessToken, err := uc.tokens.Issue(user.ID)
if err != nil {
	return nil, fmt.Errorf("issue access token: %w", err)
}

refreshToken, err := uc.tokens.IssueRefresh(user.ID)
if err != nil {
	return nil, fmt.Errorf("issue refresh token: %w", err)
}

return &ResetPasswordResult{AccessToken: accessToken, RefreshToken: refreshToken}, nil
```

- [ ] **Step 4: 新增 refresh_token.go**

```go
package app

import (
	"context"
	"errors"
	"ego-server/internal/identity/domain"
)

type RefreshTokenUseCase struct {
	tokens  TokenIssuer
	secret  []byte
}

func NewRefreshTokenUseCase(tokens TokenIssuer, secret []byte) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{tokens: tokens, secret: secret}
}

func (uc *RefreshTokenUseCase) Refresh(ctx context.Context, refreshToken string) (string, error) {
	// Verify it's a valid refresh token (not access token)
	userID, err := ParseTokenWithType(refreshToken, uc.secret, "refresh")
	if err != nil {
		return "", domain.ErrInvalidRefreshToken
	}

	// Issue new access token
	accessToken, err := uc.tokens.Issue(userID)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}
```

注意 `ParseTokenWithType` 需要从 `platform/auth` 包暴露到 `identity/app`。当前 `identity/app` 不依赖 `platform/auth`（通过接口注入）。最干净的方案是在 `app/ports.go` 或新建的 `refresh_token.go` 内部定义解析函数签名，通过依赖注入传入。

更好的方案：在 `app/ports.go` 新增一个 `RefreshTokenVerifier` 接口：

```go
// app/ports.go 新增:
type RefreshTokenVerifier interface {
	Verify(tokenStr, expectedType string) (userID string, err error)
}
```

并在 `refresh_token.go` 使用此接口：

```go
type RefreshTokenUseCase struct {
	tokens   TokenIssuer
	verifier RefreshTokenVerifier
}

func NewRefreshTokenUseCase(tokens TokenIssuer, verifier RefreshTokenVerifier) *RefreshTokenUseCase {
	return &RefreshTokenUseCase{tokens: tokens, verifier: verifier}
}

func (uc *RefreshTokenUseCase) Refresh(ctx context.Context, refreshToken string) (string, error) {
	userID, err := uc.verifier.Verify(refreshToken, "refresh")
	if err != nil {
		return "", domain.ErrInvalidRefreshToken
	}

	accessToken, err := uc.tokens.Issue(userID)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}
```

同时在 `platform/auth/jwt.go` 中实现该接口（`JWTIssuer` 已实现 `TokenIssuer`，这里新增一个 adapter 实现 `RefreshTokenVerifier`）：

在 `platform/auth/jwt.go` 末尾追加：

```go
// RefreshTokenVerifier 实现 identity app 层的 RefreshTokenVerifier 接口
type RefreshTokenVerifier struct {
	Secret []byte
}

func (v RefreshTokenVerifier) Verify(tokenStr, expectedType string) (string, error) {
	return ParseJWTWithType(tokenStr, v.Secret, expectedType)
}
```

---

### Task 6: Identity 领域 — adapter/grpc handler

**Files:**
- Modify: `server/internal/identity/adapter/grpc/handler.go`

- [ ] **Step 1: Handler 结构体新增 RefreshTokenUseCase 字段**

```go
type Handler struct {
	pb.UnimplementedEgoServer
	login         *app.LoginUseCase
	register      *app.RegisterUseCase
	sendCode      *app.SendCodeUseCase
	checkPhone    *app.CheckPhoneUseCase
	resetPassword *app.ResetPasswordUseCase
	refreshToken  *app.RefreshTokenUseCase  // 新增
}
```

- [ ] **Step 2: NewHandler 接受 refreshToken 参数**

```go
func NewHandler(
	login *app.LoginUseCase,
	register *app.RegisterUseCase,
	sendCode *app.SendCodeUseCase,
	checkPhone *app.CheckPhoneUseCase,
	resetPassword *app.ResetPasswordUseCase,
	refreshToken *app.RefreshTokenUseCase,  // 新增
) *Handler {
	return &Handler{
		login: login, register: register,
		sendCode: sendCode, checkPhone: checkPhone,
		resetPassword: resetPassword,
		refreshToken: refreshToken,
	}
}
```

- [ ] **Step 3: 更新 Login handler 返回双 token**

```go
func (h *Handler) Login(ctx context.Context, req *pb.LoginReq) (*pb.LoginRes, error) {
	result, err := h.login.Login(ctx, req.Phone, req.Password)
	if err != nil {
		return nil, mapError(err)
	}
	return &pb.LoginRes{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	}, nil
}
```

- [ ] **Step 4: 更新 Register handler 返回双 token**

```go
func (h *Handler) Register(ctx context.Context, req *pb.RegisterReq) (*pb.RegisterRes, error) {
	result, err := h.register.Register(ctx, req.Phone, req.Code, req.Password)
	if err != nil {
		return nil, mapError(err)
	}
	return &pb.RegisterRes{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	}, nil
}
```

- [ ] **Step 5: 更新 ResetPassword handler 返回双 token**

```go
func (h *Handler) ResetPassword(ctx context.Context, req *pb.ResetPasswordReq) (*pb.ResetPasswordRes, error) {
	result, err := h.resetPassword.ResetPassword(ctx, req.Phone, req.Code, req.NewPassword)
	if err != nil {
		return nil, mapError(err)
	}
	return &pb.ResetPasswordRes{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	}, nil
}
```

- [ ] **Step 6: 新增 RefreshToken handler**

```go
func (h *Handler) RefreshToken(ctx context.Context, req *pb.RefreshTokenReq) (*pb.RefreshTokenRes, error) {
	accessToken, err := h.refreshToken.Refresh(ctx, req.RefreshToken)
	if err != nil {
		return nil, mapError(err)
	}
	return &pb.RefreshTokenRes{AccessToken: accessToken}, nil
}
```

- [ ] **Step 7: mapError 新增 ErrInvalidRefreshToken**

```go
// 在 mapError 函数中新增:
if errors.Is(err, domain.ErrInvalidRefreshToken) {
	return status.Error(codes.Unauthenticated, "登录已过期，请重新登录")
}
```

---

### Task 7: Identity 领域 — module.go 组装

**Files:**
- Modify: `server/internal/identity/module.go`

- [ ] **Step 1: Deps 新增 Verifier**

```go
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
	Verifier  identityapp.RefreshTokenVerifier  // 新增
}
```

- [ ] **Step 2: NewHandler 组装 RefreshTokenUseCase**

```go
func NewHandler(deps Deps) *identitygrpc.Handler {
	queries := sqlc.New(deps.DB)

	userRepo := identitypostgres.NewUserRepository(queries)
	ids := identityid.NewUUIDGenerator()

	loginUseCase := identityapp.NewLoginUseCase(userRepo, deps.Hasher, deps.Tokens)
	registerUseCase := identityapp.NewRegisterUseCase(userRepo, deps.Hasher, deps.Tokens, ids, deps.SmsSender)
	sendCodeUseCase := identityapp.NewSendCodeUseCase(deps.SmsSender)
	checkPhoneUseCase := identityapp.NewCheckPhoneUseCase(userRepo)
	resetPasswordUseCase := identityapp.NewResetPasswordUseCase(userRepo, deps.Hasher, deps.Tokens, deps.SmsSender)
	refreshTokenUseCase := identityapp.NewRefreshTokenUseCase(deps.Tokens, deps.Verifier)  // 新增

	return identitygrpc.NewHandler(
		loginUseCase, registerUseCase, sendCodeUseCase,
		checkPhoneUseCase, resetPasswordUseCase,
		refreshTokenUseCase,  // 新增
	)
}
```

---

### Task 8: Bootstrap 层

**Files:**
- Modify: `server/internal/bootstrap/platform.go`
- Modify: `server/internal/bootstrap/identity.go`
- Modify: `server/internal/bootstrap/composite.go`

- [ ] **Step 1: platform.go — 新配置解析 + 双过期时间**

```go
// Platform struct 中删除 JWTExp，新增 JWTRefreshExp
type Platform struct {
	Pool          *pgxpool.Pool
	JWTKey        []byte
	JWTRefreshExp time.Duration   // 改名: JWTExp → JWTRefreshExp? 不，需要两个
	JWTExp        time.Duration   // access token 过期
	// ...
	Tokens        auth.JWTIssuer
	// ...
}
```

实际上 Platform struct 不需要改 — `JWTExp` 继续用作 access token 过期，新增 `RefreshTokenVerifier` 给 identity deps：

```go
type Platform struct {
	// ... 现有字段不变
	TokenVerifier auth.RefreshTokenVerifier  // 新增，给 identity 用
}
```

修改 `InitPlatform`：

```go
func InitPlatform(cfg *config.Config) (*Platform, error) {
	// ... 前面不变 ...

	jwtKey := []byte(cfg.JWTSecret)

	accessExpHours, err := strconv.Atoi(cfg.JwtAccessExpHours)
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_ACCESS_EXP_HOURS: %w", err)
	}
	accessExp := time.Duration(accessExpHours) * time.Hour

	refreshExpDays, err := strconv.Atoi(cfg.JwtRefreshExpDays)
	if err != nil {
		return nil, fmt.Errorf("invalid JWT_REFRESH_EXP_DAYS: %w", err)
	}
	refreshExp := time.Duration(refreshExpDays) * 24 * time.Hour

	// ... aiClient ...

	return &Platform{
		Pool:     pool,
		JWTKey:   jwtKey,
		JWTExp:   accessExp,
		Hasher:   auth.BcryptHasher{},
		Tokens:   auth.JWTIssuer{Secret: jwtKey, Exp: accessExp, RefreshExp: refreshExp},
		Logger:   logger,
		AIClient: aiClient,
		SmsService: newSmsService(cfg, pool),
		TokenVerifier: auth.RefreshTokenVerifier{Secret: jwtKey},  // 新增
	}, nil
}
```

- [ ] **Step 2: identity.go — 传入 Verifier**

```go
func NewIdentityHandler(p *Platform) pb.EgoServer {
	return identity.NewHandler(identity.Deps{
		DB:        p.Pool,
		Hasher:    p.Hasher,
		Tokens:    p.Tokens,
		SmsSender: p.SmsService,
		Verifier:  p.TokenVerifier,  // 新增
	})
}
```

- [ ] **Step 3: composite.go — 路由 RefreshToken**

在 `Login` 方法之后新增（按 proto 顺序，在 ResetPassword 之后）：

```go
func (h *EgoHandler) RefreshToken(ctx context.Context, req *pb.RefreshTokenReq) (*pb.RefreshTokenRes, error) {
	logger := logging.FromContext(ctx)
	logger.InfoContext(ctx, "RefreshToken: request")
	res, err := h.identity.RefreshToken(ctx, req)
	logger.InfoContext(ctx, "RefreshToken: response", "error", err)
	return res, err
}
```

---

### Task 9: 前端 — LocalStore 扩展

**Files:**
- Modify: `client/lib/data/repositories/local_store.dart`

- [ ] **Step 1: 新增 refresh token 存取**

```dart
// 新增缓存
static String? _cachedRefreshToken;

// 新增方法
static String? getRefreshToken() => _cachedRefreshToken;

static Future<void> setRefreshToken(String token) async {
  if (kIsWeb) {
    _auth.put('refresh_token', token);
  } else {
    await _secure.write(key: 'refresh_token', value: token);
  }
  _cachedRefreshToken = token;
}

static Future<void> clearRefreshToken() async {
  if (kIsWeb) {
    _auth.delete('refresh_token');
  } else {
    await _secure.delete(key: 'refresh_token');
  }
  _cachedRefreshToken = null;
}
```

- [ ] **Step 2: 修改 clearToken 同时清除 refresh token**

```dart
static Future<void> clearToken() async {
  if (kIsWeb) {
    _auth.delete('token');
    _auth.delete('refresh_token');  // 新增
  } else {
    await _secure.delete(key: 'token');
    await _secure.delete(key: 'refresh_token');  // 新增
  }
  _cachedToken = null;
  _cachedRefreshToken = null;  // 新增
}
```

- [ ] **Step 3: 修改 setToken 也同步缓存 access token**

当前 `setToken(String token)` 方法不变，但语义上它存的是 access token。为了清晰，不改方法名但内部实现保持一致。

- [ ] **Step 4: init() 中恢复 refresh token**

```dart
static Future<void> init() async {
  await Hive.initFlutter();
  _auth = await Hive.openBox(_authBox);
  _settings = await Hive.openBox(_settingsBox);
  if (kIsWeb) {
    _cachedToken = _auth.get('token');
    _cachedRefreshToken = _auth.get('refresh_token');  // 新增
  } else {
    _cachedToken = await _secure.read(key: 'token');
    _cachedRefreshToken = await _secure.read(key: 'refresh_token');  // 新增
  }
}
```

---

### Task 10: 前端 — AuthProvider 扩展

**Files:**
- Modify: `client/lib/core/providers/auth_provider.dart`

- [ ] **Step 1: AuthState 新增 refreshToken**

```dart
class AuthState {
  final String? accessToken;
  final String? refreshToken;
  final bool isLoggedIn;

  const AuthState({
    this.accessToken,
    this.refreshToken,
    this.isLoggedIn = false,
  });

  AuthState copyWith({String? accessToken, String? refreshToken, bool? isLoggedIn}) {
    return AuthState(
      accessToken: accessToken ?? this.accessToken,
      refreshToken: refreshToken ?? this.refreshToken,
      isLoggedIn: isLoggedIn ?? this.isLoggedIn,
    );
  }
}
```

- [ ] **Step 2: AuthNotifier 更新 login/logout/新增 refreshAccessToken**

```dart
class AuthNotifier extends StateNotifier<AuthState> {
  AuthNotifier() : super(const AuthState()) {
    _loadToken();
  }

  void _loadToken() {
    final token = LocalStore.getToken();
    final refreshToken = LocalStore.getRefreshToken();
    if (token != null) {
      state = AuthState(
        accessToken: token,
        refreshToken: refreshToken,
        isLoggedIn: true,
      );
    }
  }

  Future<void> login(String accessToken, String refreshToken) async {
    await LocalStore.setToken(accessToken);
    await LocalStore.setRefreshToken(refreshToken);
    state = AuthState(
      accessToken: accessToken,
      refreshToken: refreshToken,
      isLoggedIn: true,
    );
  }

  void refreshAccessToken(String accessToken) {
    // 仅更新 access token，不改变 refresh token
    LocalStore.setToken(accessToken); // fire-and-forget
    state = state.copyWith(accessToken: accessToken);
  }

  Future<void> logout() async {
    await LocalStore.clearToken();
    await LocalStore.clearRefreshToken();
    state = const AuthState();
  }
}
```

---

### Task 11: 前端 — EgoClient 自动刷新

**Files:**
- Modify: `client/lib/data/services/ego_client.dart`

- [ ] **Step 1: 新增 refreshToken 方法**

```dart
Future<grpc.RefreshTokenRes> refreshToken(String token) async {
  final req = grpc.RefreshTokenReq(refreshToken: token);
  return _stub.refreshToken(req);
}
```

- [ ] **Step 2: 新增 _autoRefresh 包装器**

在 `EgoClient` 类中新增：

```dart
/// Wraps an authenticated gRPC call with automatic token refresh on 401.
/// On UNAUTHENTICATED: attempts RefreshToken → retries once → logs out on failure.
Future<T> _autoRefresh<T>(Ref ref, Future<T> Function() grpcCall) async {
  try {
    return await grpcCall();
  } on GrpcError catch (e) {
    if (e.code != StatusCode.unauthenticated) rethrow;

    final storedRefresh = LocalStore.getRefreshToken();
    if (storedRefresh == null) rethrow;

    try {
      final res = await refreshToken(storedRefresh);
      ref.read(authProvider.notifier).refreshAccessToken(res.accessToken);
      return grpcCall(); // Retry once with new token
    } catch (_) {
      ref.read(authProvider.notifier).logout();
      rethrow;
    }
  }
}
```

- [ ] **Step 3: 将 _withAuth 方法替换为使用新字段名**

```dart
CallOptions _withAuth(Ref ref) {
  final token = ref.read(authProvider).accessToken;
  return authCallOptions(token);
}
```

- [ ] **Step 4: 包装所有需认证的方法**

所有当前使用 `_withAuth(ref)` 的方法改用 `_autoRefresh`。以 `createMoment` 为例：

```dart
Future<grpc.CreateMomentRes> createMoment(
  Ref ref, {
  required String content,
  String? traceId,
}) async {
  return _autoRefresh(ref, () {
    final req = grpc.CreateMomentReq(content: content, traceId: traceId ?? '');
    return _stub.createMoment(req, options: _withAuth(ref));
  });
}
```

其余方法类似改造：`getMoments`, `generateInsight`, `listTraces`, `getTraceDetail`, `getRandomMoments`, `stashTrace`, `listConstellations`, `getConstellation`, `startChat`, `sendMessage`, `getProfile`, `submitFeedback`。

`getProfile` 和 `submitFeedback` 当前用 `WidgetRef`，改为 `Ref` 以适配 `_autoRefresh`。同时更新调用方（在 setting_page.dart 和 feedback_page.dart 中改为传 `ref` 而非 `WidgetRef`）：

```dart
Future<grpc.GetProfileRes> getProfile(Ref ref) async {
  return _autoRefresh(ref, () {
    final req = grpc.GetProfileReq();
    return _stub.getProfile(req, options: _withAuth(ref));
  });
}

Future<grpc.SubmitFeedbackRes> submitFeedback(
  Ref ref, {
  required String content,
}) async {
  return _autoRefresh(ref, () {
    final req = grpc.SubmitFeedbackReq(content: content);
    return _stub.submitFeedback(req, options: _withAuth(ref));
  });
}
```

---

### Task 12: 前端 — LoginPage 适配新字段

**Files:**
- Modify: `client/lib/features/login/login_page.dart`

- [ ] **Step 1: 登录成功处理改为双 token**

三处 `ref.read(authProvider.notifier).login(res.token)` 改为：

```dart
// _login() 方法内:
ref.read(authProvider.notifier).login(res.accessToken, res.refreshToken);

// _register() 方法内:
ref.read(authProvider.notifier).login(res.accessToken, res.refreshToken);

// _resetPassword() 方法内:
ref.read(authProvider.notifier).login(res.accessToken, res.refreshToken);
```

---

### Task 13: 前端 — 调用方 WidgetRef → Ref 适配

**Files:**
- Modify: `client/lib/features/setting/setting_page.dart` (如果使用了 `WidgetRef` 传参)
- Modify: `client/lib/features/setting/feedback_page.dart` (同上)

- [ ] **Step 1: 检查并更新 getProfile / submitFeedback 调用**

如果 `getProfile(ref)` 和 `submitFeedback(ref, ...)` 当前接收 `WidgetRef`，由于 `_autoRefresh` 需要 `Ref` (Riverpod base type)，且 `WidgetRef` 是 `Ref` 的子类型... 实际上 `WidgetRef` 继承自 `Ref`，所以不需要改签名。`_autoRefresh(Ref ref, ...)` 可以接收 `WidgetRef`。

确认：`Ref` 是 Riverpod 的基类型，`WidgetRef` extends `Ref`，`Ref` (from `provider.dart`) extends `Ref`。所以 `Ref` 是最宽的类型。保持 `_autoRefresh` 使用 `Ref` 参数，`WidgetRef` 可以直接传入。

因此无需改动调用方。

---

### Task 14: Smoke 测试更新

**Files:**
- Modify: `smoke.sh`

- [ ] **Step 1: 新增 RefreshToken smoke 测试**

在现有 JWT 相关测试后新增：

```bash
# ─── RefreshToken ──────────────────────────────
echo ""
echo "=== RefreshToken ==="

# Extract tokens from Login response (JSON with grpcurl)
LOGIN_JSON=$(grpcurl -plaintext -d "{\"phone\":\"$TEST_PHONE\",\"password\":\"$TEST_PASSWORD\"}" \
  localhost:$GRPC_PORT ego.Ego/Login 2>/dev/null)
ACCESS_TOKEN=$(echo "$LOGIN_JSON" | jq -r '.accessToken')
REFRESH_TOKEN=$(echo "$LOGIN_JSON" | jq -r '.refreshToken')

if [ -z "$ACCESS_TOKEN" ] || [ "$ACCESS_TOKEN" = "null" ]; then
  echo "FAIL: Login returned no access token"
  exit 1
fi
echo "OK: Login returned access token"

if [ -z "$REFRESH_TOKEN" ] || [ "$REFRESH_TOKEN" = "null" ]; then
  echo "FAIL: Login returned no refresh token"
  exit 1
fi
echo "OK: Login returned refresh token"

# Refresh: get new access token
NEW_ACCESS=$(grpcurl -plaintext \
  -d "{\"refresh_token\":\"$REFRESH_TOKEN\"}" \
  localhost:$GRPC_PORT ego.Ego/RefreshToken 2>/dev/null | jq -r '.accessToken')

if [ -z "$NEW_ACCESS" ] || [ "$NEW_ACCESS" = "null" ]; then
  echo "FAIL: RefreshToken returned no access token"
  exit 1
fi
echo "OK: RefreshToken returned new access token"

# Verify new access token works
grpcurl -plaintext \
  -H "Authorization: Bearer $NEW_ACCESS" \
  -d "{}" \
  localhost:$GRPC_PORT ego.Ego/GetProfile > /dev/null 2>&1
if [ $? -ne 0 ]; then
  echo "FAIL: new access token from RefreshToken not accepted by GetProfile"
  exit 1
fi
echo "OK: new access token accepted by authenticated RPC"

# Verify refresh token cannot be used for authenticated RPC
grpcurl -plaintext \
  -H "Authorization: Bearer $REFRESH_TOKEN" \
  -d "{}" \
  localhost:$GRPC_PORT ego.Ego/GetProfile > /dev/null 2>&1
if [ $? -eq 0 ]; then
  echo "FAIL: refresh token should not be accepted for authenticated RPC"
  exit 1
fi
echo "OK: refresh token rejected for authenticated RPC (expected)"

# Verify RefreshToken without token returns UNAUTHENTICATED
grpcurl -plaintext \
  -d "{}" \
  localhost:$GRPC_PORT ego.Ego/RefreshToken 2>&1 | grep -q "UNAUTHENTICATED\|unauthenticated\|Login expired\|登录已过期"
if [ $? -ne 0 ]; then
  echo "FAIL: empty RefreshToken should return UNAUTHENTICATED"
  exit 1
fi
echo "OK: empty RefreshToken returns UNAUTHENTICATED"
```

注意：smoke.sh 中现有的 Login 测试需要更新：当前检查 `jq -r '.token'` 需要改为 `jq -r '.accessToken'`。

---

### Task 15: 编译验证 + 静态检查

- [ ] **Step 1: Go 编译**

```bash
cd server && go build ./...
```
预期：编译成功

- [ ] **Step 2: Flutter 静态分析**

```bash
cd client && flutter analyze
```
预期：零 issue

- [ ] **Step 3: Go vet**

```bash
cd server && go vet ./internal/identity/... ./internal/platform/auth/... ./internal/bootstrap/... ./internal/config/...
```
预期：无错误

---

### Task 16: 提交

全部检查通过 + 真机测试通过后，执行一次 commit：

```bash
git add -A
git commit -m "feat(identity): add JWT refresh token mechanism

- Add RefreshToken RPC with short-lived access (1h) + long-lived refresh (30d) tokens
- Proto: rename token→access_token+refresh_token in LoginRes/RegisterRes/ResetPasswordRes
- Server: add ParseJWTWithType, RefreshTokenUseCase, double-token issuing
- Client: auto-refresh on 401, dual token storage in LocalStore
- Update smoke.sh with RefreshToken end-to-end tests
- Config: replace JWT_EXP_HOURS with JWT_ACCESS_EXP_HOURS + JWT_REFRESH_EXP_DAYS

Co-Authored-By: Claude <noreply@anthropic.com>"
```
