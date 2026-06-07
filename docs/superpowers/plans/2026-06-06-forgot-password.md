# Forgot Password Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add ResetPassword RPC and frontend "Forgot Password" flow to the Login module.

**Architecture:** New `ResetPassword` RPC (pre-auth) in the identity domain. Reuses existing SMS verification, password hashing, JWT issuing, and UserRepository. Frontend adds a new Step 3 within the existing `/login` page.

**Tech Stack:** Go (gRPC, sqlc/pgx), Flutter (Riverpod, grpc-dart), Protocol Buffers

---

### Task 1: Proto Change + Regenerate

**Files:**
- Modify: `proto/ego/api.proto`
- Regenerate: `server/proto/ego/` (Go), `client/lib/data/generated/` (Dart)

- [ ] **Step 1: Add ResetPassword rpc to Ego service**

In `proto/ego/api.proto`, add the new RPC entry after the existing `Register` rpc:

```protobuf
  rpc ResetPassword(ResetPasswordReq) returns (ResetPasswordRes);
```

Insert it right after line 18 (`rpc Register...`) so it stays in the Auth block:

```protobuf
  // ─── Auth（认证）─────────────────────────────────────
  rpc Login(LoginReq) returns (LoginRes);
  rpc CheckPhone(CheckPhoneReq) returns (CheckPhoneRes);
  rpc SendVerificationCode(SendVerificationCodeReq) returns (SendVerificationCodeRes);
  rpc Register(RegisterReq) returns (RegisterRes);
  rpc ResetPassword(ResetPasswordReq) returns (ResetPasswordRes);
```

- [ ] **Step 2: Add ResetPasswordReq/Res messages**

In `proto/ego/api.proto`, add the new messages after the existing `RegisterRes` message (after line 96):

```protobuf
message ResetPasswordReq {
  string phone        = 1;
  string code         = 2;  // SMS verification code
  string new_password = 3;
}

message ResetPasswordRes {
  string token = 1;
}
```

- [ ] **Step 3: Regenerate Go proto**

```bash
make proto-go
```

Expected: `server/proto/ego/api.pb.go` and `server/proto/ego/api_grpc.pb.go` updated with `ResetPassword` types.

- [ ] **Step 4: Regenerate Dart proto**

```bash
make proto-dart
```

Expected: `client/lib/data/generated/api.pb.dart` and `client/lib/data/generated/api.pbgrpc.dart` updated with `ResetPassword` types.

- [ ] **Step 5: Commit**

```bash
git add proto/ego/api.proto server/proto/ego/ client/lib/data/generated/
git commit -m "feat(proto): add ResetPassword rpc and messages

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 2: sqlc Query — UpdateUserPassword

**Files:**
- Modify: `server/internal/platform/postgres/queries/users.sql`
- Regenerate: `server/internal/platform/postgres/sqlc/users.sql.go`

- [ ] **Step 1: Add UpdateUserPassword query**

Append to `server/internal/platform/postgres/queries/users.sql`:

```sql

-- name: UpdateUserPassword :exec
UPDATE users SET password_hash = $2 WHERE id = $1;
```

- [ ] **Step 2: Run sqlc to regenerate**

```bash
cd server && make sqlc
```

Expected: `server/internal/platform/postgres/sqlc/users.sql.go` updated with `UpdateUserPasswordParams` and `UpdateUserPassword` method.

- [ ] **Step 3: Verify generated code compiles**

```bash
cd server && go build ./...
```

Expected: build succeeds.

- [ ] **Step 4: Commit**

```bash
git add server/internal/platform/postgres/queries/users.sql server/internal/platform/postgres/sqlc/users.sql.go
git commit -m "feat(sqlc): add UpdateUserPassword query

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 3: Domain — Add UpdatePassword to Repository Interface

**Files:**
- Modify: `server/internal/identity/domain/user.go`

- [ ] **Step 1: Add UpdatePassword method to UserRepository interface**

In `server/internal/identity/domain/user.go`, change the `UserRepository` interface to add `UpdatePassword`:

```go
type UserRepository interface {
	FindByPhone(ctx context.Context, phone string) (*User, error)
	Create(ctx context.Context, user *User) error
	UpdatePassword(ctx context.Context, userID, passwordHash string) error
}
```

- [ ] **Step 2: Commit**

```bash
git add server/internal/identity/domain/user.go
git commit -m "feat(domain): add UpdatePassword to UserRepository interface

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 4: App — ResetPassword Use Case

**Files:**
- Create: `server/internal/identity/app/reset_password.go`

- [ ] **Step 1: Create ResetPassword use case**

Create `server/internal/identity/app/reset_password.go`:

```go
package app

import (
	"context"
	"errors"
	"fmt"

	"ego-server/internal/identity/domain"
)

type ResetPasswordUseCase struct {
	users     domain.UserRepository
	hasher    PasswordHasher
	tokens    TokenIssuer
	smsSender SmsService
}

func NewResetPasswordUseCase(
	users domain.UserRepository,
	hasher PasswordHasher,
	tokens TokenIssuer,
	smsSender SmsService,
) *ResetPasswordUseCase {
	return &ResetPasswordUseCase{
		users:     users,
		hasher:    hasher,
		tokens:    tokens,
		smsSender: smsSender,
	}
}

type ResetPasswordResult struct {
	Token string
}

func (uc *ResetPasswordUseCase) ResetPassword(ctx context.Context, phone, code, newPassword string) (*ResetPasswordResult, error) {
	if !phonePattern.MatchString(phone) {
		return nil, domain.ErrInvalidPhone
	}
	if len(newPassword) < 6 {
		return nil, fmt.Errorf("password too short")
	}

	ok, err := uc.smsSender.Verify(ctx, phone, code)
	if err != nil {
		return nil, fmt.Errorf("verify code: %w", err)
	}
	if !ok {
		return nil, domain.ErrInvalidVerificationCode
	}

	user, err := uc.users.FindByPhone(ctx, phone)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("find user by phone: %w", err)
	}

	hash, err := uc.hasher.Hash(newPassword)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	if err := uc.users.UpdatePassword(ctx, user.ID, hash); err != nil {
		return nil, fmt.Errorf("update password: %w", err)
	}

	token, err := uc.tokens.Issue(user.ID)
	if err != nil {
		return nil, fmt.Errorf("issue token: %w", err)
	}

	return &ResetPasswordResult{Token: token}, nil
}
```

- [ ] **Step 2: Commit**

```bash
git add server/internal/identity/app/reset_password.go
git commit -m "feat(app): add ResetPassword use case

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 5: Adapter — Postgres UpdatePassword Implementation

**Files:**
- Modify: `server/internal/identity/adapter/postgres/user_repo.go`

- [ ] **Step 1: Implement UpdatePassword on UserRepository**

Add the `UpdatePassword` method to `server/internal/identity/adapter/postgres/user_repo.go`.

Insert after the `Create` method (after line 60):

```go
func (r *UserRepository) UpdatePassword(ctx context.Context, userID, passwordHash string) error {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	return r.queries.UpdateUserPassword(ctx, sqlc.UpdateUserPasswordParams{
		ID:           pgtype.UUID{Bytes: [16]byte(uid), Valid: true},
		PasswordHash: passwordHash,
	})
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd server && go build ./...
```

Expected: build succeeds.

- [ ] **Step 3: Commit**

```bash
git add server/internal/identity/adapter/postgres/user_repo.go
git commit -m "feat(adapter): implement UpdatePassword in postgres repo

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 6: Adapter — gRPC Handler

**Files:**
- Modify: `server/internal/identity/adapter/grpc/handler.go`

- [ ] **Step 1: Add ResetPasswordUseCase field to Handler struct**

In `server/internal/identity/adapter/grpc/handler.go`, update the `Handler` struct to include the new use case:

```go
type Handler struct {
	pb.UnimplementedEgoServer
	login         *app.LoginUseCase
	register      *app.RegisterUseCase
	sendCode      *app.SendCodeUseCase
	checkPhone    *app.CheckPhoneUseCase
	resetPassword *app.ResetPasswordUseCase
}
```

- [ ] **Step 2: Update NewHandler constructor**

Update `NewHandler` to accept the new use case:

```go
func NewHandler(
	login *app.LoginUseCase,
	register *app.RegisterUseCase,
	sendCode *app.SendCodeUseCase,
	checkPhone *app.CheckPhoneUseCase,
	resetPassword *app.ResetPasswordUseCase,
) *Handler {
	return &Handler{login: login, register: register, sendCode: sendCode, checkPhone: checkPhone, resetPassword: resetPassword}
}
```

- [ ] **Step 3: Add ResetPassword handler method**

Add after the `Login` handler method (after line 62):

```go
func (h *Handler) ResetPassword(ctx context.Context, req *pb.ResetPasswordReq) (*pb.ResetPasswordRes, error) {
	result, err := h.resetPassword.ResetPassword(ctx, req.Phone, req.Code, req.NewPassword)
	if err != nil {
		return nil, mapError(err)
	}
	return &pb.ResetPasswordRes{Token: result.Token}, nil
}
```

- [ ] **Step 4: Verify compilation**

```bash
cd server && go build ./...
```

Expected: build fails at `module.go` — that's expected, will fix in next task.

- [ ] **Step 5: Commit**

```bash
git add server/internal/identity/adapter/grpc/handler.go
git commit -m "feat(adapter): add ResetPassword gRPC handler

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 7: Wiring — module.go + auth interceptor + composite

**Files:**
- Modify: `server/internal/identity/module.go`
- Modify: `server/internal/platform/auth/interceptor.go`
- Modify: `server/internal/bootstrap/composite.go`

- [ ] **Step 1: Wire ResetPasswordUseCase in module.go**

In `server/internal/identity/module.go`, update `NewHandler` to create and inject the new use case:

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

	return identitygrpc.NewHandler(loginUseCase, registerUseCase, sendCodeUseCase, checkPhoneUseCase, resetPasswordUseCase)
}
```

- [ ] **Step 2: Add ResetPassword to preAuthMethods**

In `server/internal/platform/auth/interceptor.go`, add `"ResetPassword": true` to the `preAuthMethods` map:

```go
var preAuthMethods = map[string]bool{
	"Login":                 true,
	"CheckPhone":            true,
	"SendVerificationCode":  true,
	"Register":              true,
	"ResetPassword":         true,
}
```

- [ ] **Step 3: Add ResetPassword delegation in composite.go**

In `server/internal/bootstrap/composite.go`, add the `ResetPassword` delegation method after the `Register` delegation (after line 65):

```go
func (h *EgoHandler) ResetPassword(ctx context.Context, req *pb.ResetPasswordReq) (*pb.ResetPasswordRes, error) {
	logger := logging.FromContext(ctx)
	logger.InfoContext(ctx, "ResetPassword: request", "req", fmt.Sprintf("%+v", req))
	res, err := h.identity.ResetPassword(ctx, req)
	logger.InfoContext(ctx, "ResetPassword: response", "res", fmt.Sprintf("%+v", res), "error", err)
	return res, err
}
```

- [ ] **Step 4: Verify compilation**

```bash
cd server && go build ./...
```

Expected: build succeeds.

- [ ] **Step 5: Commit**

```bash
git add server/internal/identity/module.go server/internal/platform/auth/interceptor.go server/internal/bootstrap/composite.go
git commit -m "feat(wiring): wire ResetPassword through module, interceptor, composite

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 8: Frontend — ego_client.dart

**Files:**
- Modify: `client/lib/data/services/ego_client.dart`

- [ ] **Step 1: Add resetPassword method to EgoClient**

In `client/lib/data/services/ego_client.dart`, add the new method in the Auth section, after the `login` method (after line 55):

```dart
  Future<grpc.ResetPasswordRes> resetPassword({
    required String phone,
    required String code,
    required String newPassword,
  }) async {
    final req = grpc.ResetPasswordReq(phone: phone, code: code, newPassword: newPassword);
    return _stub.resetPassword(req);
  }
```

- [ ] **Step 2: Verify Flutter static analysis**

```bash
cd client && flutter analyze
```

Expected: no new errors (existing errors unrelated to this change may appear).

- [ ] **Step 3: Commit**

```bash
git add client/lib/data/services/ego_client.dart
git commit -m "feat(client): add resetPassword method to EgoClient

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

---

### Task 9: Frontend — login_page.dart (Forgot Password Flow)

**Files:**
- Modify: `client/lib/features/login/login_page.dart`

- [ ] **Step 1: Add `_goToStep3` method**

Insert after `_backToStep0` (after line 188), before the `build` method:

```dart
  Future<void> _goToStep3() async {
    final phone = _phoneCtrl.text.trim();
    setState(() { _loading = true; _error = null; });

    try {
      final client = ref.read(EgoClient.provider);
      await client.sendVerificationCode(phone);
      if (!mounted) return;
      setState(() { _loading = false; _step = 3; _countdown = 60; _lastSentPhone = phone; });
      _startCountdown();
    } catch (_) {
      if (!mounted) return;
      setState(() => _loading = false);
      _setError('发送验证码失败，请稍后重试');
    }
  }
```

- [ ] **Step 2: Add `_resetPassword` method**

Insert after `_goToStep3`:

```dart
  Future<void> _resetPassword() async {
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
      final res = await client.resetPassword(
        phone: _phoneCtrl.text.trim(),
        code: code,
        newPassword: password,
      );
      if (mounted) {
        ref.read(authProvider.notifier).login(res.token);
      }
    } on GrpcError catch (e) {
      if (mounted) {
        setState(() => _loading = false);
        if (e.code == StatusCode.unauthenticated) {
          _setError('验证码错误');
        } else if (e.code == StatusCode.notFound) {
          _setError('用户不存在');
        } else {
          _setError('重置密码失败，请稍后重试');
        }
      }
    } catch (_) {
      if (mounted) setState(() => _loading = false);
      _setError('重置密码失败，请稍后重试');
    }
  }
```

- [ ] **Step 3: Add "Forgot Password" link in Step 1 UI**

In the `build` method, find the Step 1 password input (lines ~262-272) and add the link below it. Change this block:

```dart
                // Step 1: Password login
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
```

To:

```dart
                // Step 1: Password login
                if (_step == 1) ...[
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
                  const SizedBox(height: 8),
                  Align(
                    alignment: Alignment.centerRight,
                    child: GestureDetector(
                      onTap: _loading ? null : _goToStep3,
                      child: const Text(
                        '忘记密码？',
                        style: TextStyle(fontSize: 12, color: Color(0xFFCCA880)),
                      ),
                    ),
                  ),
                ],
```

- [ ] **Step 4: Add Step 3 UI**

Insert after the Step 2 UI block and before the error/button section. Add this after line 316 (after `]` closing Step 2's column children):

```dart
                // Step 3: Forgot password — code + new password
                if (_step == 3) ...[
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
                          onPressed: _countdown > 0 || _loading ? null : () => _resendCode(),
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
                      hintText: '设置新密码（至少 6 位）',
                      prefixIcon: Icon(Icons.lock_outline),
                    ),
                    textInputAction: TextInputAction.done,
                    onSubmitted: (_) => _resetPassword(),
                  ),
                ],
```

- [ ] **Step 5: Add Step 3 title text**

In the title section (around line 210), update the title to cover Step 3. Change:

```dart
                  Padding(
                    padding: const EdgeInsets.only(bottom: 16),
                    child: Text(
                      _step == 1 ? '密码登录' : '创建账号',
```

To:

```dart
                  Padding(
                    padding: const EdgeInsets.only(bottom: 16),
                    child: Text(
                      _step == 1 ? '密码登录' : _step == 3 ? '重置密码' : '创建账号',
```

- [ ] **Step 6: Add Step 3 button**

In the button section (around line 325-354), add the Step 3 button before the closing `]` of the bottom buttons. Insert after the Step 2 button:

```dart
                if (_step == 3)
                  ElevatedButton(
                    onPressed: _loading ? null : _resetPassword,
                    child: _loading
                        ? const SizedBox(
                            height: 20, width: 20,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : const Text('重置密码', style: TextStyle(fontSize: 16)),
                  ),
```

- [ ] **Step 7: Run Flutter static analysis**

```bash
cd client && flutter analyze
```

Expected: no new errors.

- [ ] **Step 8: Commit**

```bash
git add client/lib/features/login/login_page.dart
git commit -m "feat(client): add forgot password Step 3 flow

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```
