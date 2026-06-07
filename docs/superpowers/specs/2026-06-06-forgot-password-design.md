# Forgot Password Feature Design

**Date:** 2026-06-06
**Domain:** identity (Login module)
**Status:** Approved

## Overview

Add "Forgot Password" flow to the Login module. User clicks "Forgot Password?" on the password login step (Step 1), receives an SMS verification code, enters code + new password, and gets automatically logged in with a new JWT.

---

## 1. Proto Contract

### New RPC

```protobuf
service Ego {
  // ... existing RPCs ...
  rpc ResetPassword(ResetPasswordReq) returns (ResetPasswordRes);
}
```

### New Messages

```protobuf
message ResetPasswordReq {
  string phone        = 1;
  string code         = 2;  // SMS verification code
  string new_password = 3;
}

message ResetPasswordRes {
  string token = 1;  // JWT for auto-login after reset
}
```

### Auth

`ResetPassword` is a pre-auth RPC — added to the `preAuthMethods` whitelist alongside Login, CheckPhone, SendVerificationCode, and Register.

---

## 2. Backend (DDD)

### domain/

- **No new entity** — reuses `User{ID, Phone, PasswordHash}`
- **No new errors** — reuses `ErrUserNotFound`, `ErrInvalidPhone`, `ErrInvalidVerificationCode`, `ErrCodeExpired`

### app/reset_password.go — New Use Case

```
1. Validate phone format (1[3-9]xxxxxxxxx) → ErrInvalidPhone
2. Validate new password >= 6 chars → InvalidArgument
3. smsSender.Verify(phone, code) → ErrInvalidVerificationCode / ErrCodeExpired
4. userRepo.FindByPhone(phone) → ErrUserNotFound
5. hasher.Hash(newPassword) → bcrypt hash
6. userRepo.UpdatePassword(userID, hash)
7. issuer.Issue(userID) → token
```

### app/ports.go — No change

All interfaces (`PasswordHasher`, `TokenIssuer`, `SmsService`) reused as-is.

### adapter/postgres/user_repo.go — New Method

```go
UpdatePassword(ctx context.Context, userID, passwordHash string) error
```

### adapter/grpc/handler.go — New Handler

```go
ResetPassword(ctx, req) → app.ResetPassword(phone, code, newPassword) → {token}
```

### platform/auth/interceptor.go — Whitelist

Add `"ResetPassword": true` to `preAuthMethods`.

### bootstrap/composite.go — Routing

```go
EgoHandler.ResetPassword → identity.ResetPassword
```

---

## 3. Frontend

### Flow Update

```
Step 0: Enter phone → CheckPhone
  ├─ registered=true  → Step 1 (password login)
  │   └─ "Forgot Password?" → auto-send SMS → Step 3
  └─ registered=false → Step 2 (code registration)

Step 3: Enter code + new password → ResetPassword → auto-login → /now or /onboard
```

### login_page.dart Changes

1. **Step 1:** Add "Forgot Password?" text link below password input
2. **`_goToStep3()`:** Auto-send `sendVerificationCode(phone)`, set `_step = 3`, start countdown
3. **Step 3 UI:** Verification code input (with resend) + new password input + submit button
4. **`_resetPassword()`:** Call `client.resetPassword(phone, code, newPassword)`, on success auto-login via `authProvider`

### ego_client.dart — New Method

```dart
Future<ResetPasswordRes> resetPassword({
  required String phone,
  required String code,
  required String newPassword,
});
```

### router.dart — No change

Forgot password is an internal state transition within `/login`, no new route needed.

---

## 4. Database

No schema change. Only a new UPDATE query:

```sql
-- name: UpdateUserPassword :exec
UPDATE users SET password_hash = $1 WHERE id = $2;
```

---

## 5. Error Mapping

| Domain Error | gRPC Status | Frontend Message |
|---|---|---|
| `ErrInvalidPhone` | `InvalidArgument` | "请输入正确的手机号" |
| password < 6 | `InvalidArgument` | "密码至少6位" |
| `ErrInvalidVerificationCode` | `Unauthenticated` | "验证码错误" |
| `ErrCodeExpired` | `Unauthenticated` | "验证码已过期" |
| `ErrUserNotFound` | `NotFound` | "用户不存在" |

---

## 6. Auth Interceptor Whitelist

```go
var preAuthMethods = map[string]bool{
    "Login":                 true,
    "CheckPhone":            true,
    "SendVerificationCode":  true,
    "Register":              true,
    "ResetPassword":         true,  // NEW
}
```
