# Setting Module Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add an independent `setting` module with account info display and logout functionality, following the project's DDD + Flutter Riverpod architecture.

**Architecture:** New `server/internal/setting/` DDD module (domain → app → adapter), a `GetProfile` RPC added to proto, and a `client/lib/features/setting/` Flutter page. Setting module reads users table via a new `GetUserByID` sqlc query, entry point via settings icon in AppShell top-left.

**Tech Stack:** Go gRPC (proto3), Flutter Riverpod + GoRouter, sqlc, pgx

---

## File Structure

| Action | File | Purpose |
|--------|------|---------|
| Modify | `proto/ego/api.proto` | Add `GetProfile` RPC + messages |
| Modify | `server/internal/platform/postgres/queries/users.sql` | Add `GetUserByID` query |
| Create | `server/internal/setting/domain/ports.go` | `UserInfo` type + `UserReader` interface |
| Create | `server/internal/setting/domain/errors.go` | Domain errors |
| Create | `server/internal/setting/app/profile.go` | `GetProfileUseCase` |
| Create | `server/internal/setting/adapter/postgres/user_reader.go` | `UserReader` implementation |
| Create | `server/internal/setting/adapter/grpc/handler.go` | gRPC handler + error mapper |
| Create | `server/internal/setting/module.go` | DI wiring |
| Create | `server/internal/bootstrap/setting.go` | Bootstrap factory |
| Modify | `server/internal/bootstrap/composite.go` | Add setting handler to composite |
| Modify | `server/cmd/ego/main.go` | Wire setting handler |
| Create | `client/lib/features/setting/setting_page.dart` | Settings UI page |
| Modify | `client/lib/core/router/router.dart` | Add `/setting` route |
| Modify | `client/lib/data/services/ego_client.dart` | Add `getProfile()` method |
| Modify | `client/lib/shared/widgets/app_shell.dart` | Add settings icon entry (top-left) |

---

### Task 1: Proto — Add GetProfile RPC + Messages

**Files:**
- Modify: `proto/ego/api.proto`

- [ ] **Step 1: Add RPC to service Ego**

After the `ResetPassword` RPC line (line 20), add:

```protobuf
// ─── Setting（设置）──────────────────────────────────
rpc GetProfile(GetProfileReq) returns (GetProfileRes);
```

- [ ] **Step 2: Add message definitions**

After the `ResetPasswordRes` message block (after line 107), add:

```protobuf
message GetProfileReq {}

message GetProfileRes {
  string phone     = 1;  // 手机号
  int64  created_at = 2;  // 注册时间 unix timestamp ms
}
```

---

### Task 2: Regenerate Proto Code

**Files:**
- Create: `server/proto/ego/` (multiple generated files)
- Create: `client/lib/data/generated/` (multiple generated files)

- [ ] **Step 1: Generate Go proto stubs**

```bash
cd server && make proto-go
```

Expected: Regenerates `server/proto/ego/api.pb.go` and `api_grpc.pb.go` with new `GetProfile` RPC.

- [ ] **Step 2: Generate Dart proto stubs**

```bash
cd client && make proto-dart
```

Expected: Regenerates `client/lib/data/generated/` files including `api.pb.dart`, `api.pbgrpc.dart`, `api.pbjson.dart` with new `getProfile` method on `EgoClient`.

- [ ] **Step 3: Commit proto generation**

```bash
git add proto/ego/api.proto server/proto/ego/ client/lib/data/generated/
git commit -m "feat(setting): add GetProfile RPC to proto contract"
```

---

### Task 3: sqlc — Add GetUserByID Query

**Files:**
- Modify: `server/internal/platform/postgres/queries/users.sql`
- Modify: `server/internal/platform/postgres/sqlc/users.sql.go` (regenerated)

- [ ] **Step 1: Add query to users.sql**

Append to `server/internal/platform/postgres/queries/users.sql`:

```sql
-- name: GetUserByID :one
SELECT phone, created_at FROM users WHERE id = $1;
```

- [ ] **Step 2: Regenerate sqlc code**

```bash
cd server && make sqlc
```

Expected: `server/internal/platform/postgres/sqlc/users.sql.go` now contains `GetUserByID` function with `GetUserByIDRow` struct having `Phone string` and `CreatedAt pgtype.Timestamptz`.

- [ ] **Step 3: Commit sqlc changes**

```bash
git add server/internal/platform/postgres/queries/users.sql server/internal/platform/postgres/sqlc/users.sql.go
git commit -m "feat(sqlc): add GetUserByID query for setting module"
```

---

### Task 4: Setting Domain Layer

**Files:**
- Create: `server/internal/setting/domain/ports.go`
- Create: `server/internal/setting/domain/errors.go`

- [ ] **Step 1: Create ports.go**

```go
package domain

import (
	"context"
	"time"
)

// UserInfo is the read-only view of a user for the setting module.
type UserInfo struct {
	Phone     string
	CreatedAt time.Time
}

// UserReader reads user data by ID.
type UserReader interface {
	FindByID(ctx context.Context, id string) (*UserInfo, error)
}
```

- [ ] **Step 2: Create errors.go**

```go
package domain

import "errors"

var ErrUserNotFound = errors.New("user not found")
```

---

### Task 5: Setting App Layer

**Files:**
- Create: `server/internal/setting/app/profile.go`

- [ ] **Step 1: Create profile.go**

```go
package app

import (
	"context"

	"ego-server/internal/setting/domain"
)

// GetProfileUseCase retrieves the current user's profile information.
type GetProfileUseCase struct {
	userReader domain.UserReader
}

func NewGetProfileUseCase(userReader domain.UserReader) *GetProfileUseCase {
	return &GetProfileUseCase{userReader: userReader}
}

// ProfileResult holds the profile data returned to the caller.
type ProfileResult struct {
	Phone     string
	CreatedAt int64  // unix timestamp ms
}

func (uc *GetProfileUseCase) GetProfile(ctx context.Context, userID string) (*ProfileResult, error) {
	user, err := uc.userReader.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &ProfileResult{
		Phone:     user.Phone,
		CreatedAt: user.CreatedAt.UnixMilli(),
	}, nil
}
```

---

### Task 6: Setting Adapter — Postgres UserReader

**Files:**
- Create: `server/internal/setting/adapter/postgres/user_reader.go`

- [ ] **Step 1: Create user_reader.go**

```go
package postgres

import (
	"context"
	"errors"

	"ego-server/internal/platform/postgres/sqlc"
	settingdomain "ego-server/internal/setting/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// UserReader implements setting/domain.UserReader using sqlc.
type UserReader struct {
	queries *sqlc.Queries
}

func NewUserReader(queries *sqlc.Queries) *UserReader {
	return &UserReader{queries: queries}
}

func (r *UserReader) FindByID(ctx context.Context, id string) (*settingdomain.UserInfo, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, settingdomain.ErrUserNotFound
	}

	row, err := r.queries.GetUserByID(ctx, pgtype.UUID{Bytes: uid, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, settingdomain.ErrUserNotFound
		}
		return nil, err
	}

	return &settingdomain.UserInfo{
		Phone:     row.Phone,
		CreatedAt: row.CreatedAt.Time,
	}, nil
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd server && go build ./internal/setting/adapter/postgres/
```

---

### Task 7: Setting Adapter — gRPC Handler

**Files:**
- Create: `server/internal/setting/adapter/grpc/handler.go`

- [ ] **Step 1: Create handler.go**

```go
package grpc

import (
	"context"
	"errors"

	"ego-server/internal/setting/app"
	"ego-server/internal/setting/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "ego-server/proto/ego"
)

// Handler implements pb.EgoServer for the setting module.
type Handler struct {
	pb.UnimplementedEgoServer
	getProfile *app.GetProfileUseCase
}

func NewHandler(getProfile *app.GetProfileUseCase) *Handler {
	return &Handler{getProfile: getProfile}
}

func (h *Handler) GetProfile(ctx context.Context, req *pb.GetProfileReq) (*pb.GetProfileRes, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "未登录")
	}

	result, err := h.getProfile.GetProfile(ctx, userID)
	if err != nil {
		return nil, mapError(err)
	}

	return &pb.GetProfileRes{
		Phone:     result.Phone,
		CreatedAt: result.CreatedAt,
	}, nil
}

func mapError(err error) error {
	if errors.Is(err, domain.ErrUserNotFound) {
		return status.Error(codes.NotFound, "用户不存在")
	}
	return status.Error(codes.Internal, err.Error())
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd server && go build ./internal/setting/adapter/grpc/
```

---

### Task 8: Setting Module — Wiring

**Files:**
- Create: `server/internal/setting/module.go`

- [ ] **Step 1: Create module.go**

```go
package setting

import (
	settinggrpc "ego-server/internal/setting/adapter/grpc"
	settingpostgres "ego-server/internal/setting/adapter/postgres"
	settingapp "ego-server/internal/setting/app"
	"ego-server/internal/platform/postgres/sqlc"
)

// Deps contains process-level resources needed by the setting module.
type Deps struct {
	DB sqlc.DBTX
}

// NewHandler wires the setting module's adapters, use cases, and gRPC handler.
func NewHandler(deps Deps) *settinggrpc.Handler {
	queries := sqlc.New(deps.DB)
	userReader := settingpostgres.NewUserReader(queries)
	getProfileUseCase := settingapp.NewGetProfileUseCase(userReader)
	return settinggrpc.NewHandler(getProfileUseCase)
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd server && go build ./internal/setting/
```

---

### Task 9: Bootstrap & Main Wiring

**Files:**
- Create: `server/internal/bootstrap/setting.go`
- Modify: `server/internal/bootstrap/composite.go`
- Modify: `server/cmd/ego/main.go`

- [ ] **Step 1: Create bootstrap/setting.go**

```go
package bootstrap

import (
	"ego-server/internal/setting"

	pb "ego-server/proto/ego"
)

func NewSettingHandler(p *Platform) pb.EgoServer {
	return setting.NewHandler(setting.Deps{
		DB: p.Pool,
	})
}
```

- [ ] **Step 2: Modify composite.go — add setting field**

In `EgoHandler` struct (line 16-22), add `setting pb.EgoServer` field:

```go
type EgoHandler struct {
	pb.UnimplementedEgoServer
	identity pb.EgoServer
	writing  pb.EgoServer
	timeline pb.EgoServer
	starmap  pb.EgoServer
	chat     pb.EgoServer
	setting  pb.EgoServer
}
```

- [ ] **Step 3: Modify composite.go — update NewEgoHandler signature**

Replace the `NewEgoHandler` function (line 24-32):

```go
func NewEgoHandler(identity, writing, timeline, starmap, chat, setting pb.EgoServer) *EgoHandler {
	return &EgoHandler{
		identity: identity,
		writing:  writing,
		timeline: timeline,
		starmap:  starmap,
		chat:     chat,
		setting:  setting,
	}
}
```

- [ ] **Step 4: Modify composite.go — add GetProfile delegation**

Append at end of file (after SendMessage method, before end of file):

```go
// Setting — delegated to setting.
func (h *EgoHandler) GetProfile(ctx context.Context, req *pb.GetProfileReq) (*pb.GetProfileRes, error) {
	logger := logging.FromContext(ctx)
	logger.InfoContext(ctx, "GetProfile: request", "req", fmt.Sprintf("%+v", req))
	res, err := h.setting.GetProfile(ctx, req)
	logger.InfoContext(ctx, "GetProfile: response", "res", fmt.Sprintf("%+v", res), "error", err)
	return res, err
}
```

- [ ] **Step 5: Modify main.go**

In `server/cmd/ego/main.go`, add after `chatHandler` and before `handler` lines:

```go
settingHandler := bootstrap.NewSettingHandler(p)
```

And update the `NewEgoHandler` call to include `settingHandler`:

```go
handler := bootstrap.NewEgoHandler(identityHandler, writingHandler, timelineHandler, starmapHandler, chatHandler, settingHandler)
```

- [ ] **Step 6: Verify server compiles**

```bash
cd server && go build ./...
```

---

### Task 10: Frontend — Setting Page

**Files:**
- Create: `client/lib/features/setting/setting_page.dart`

- [ ] **Step 1: Create setting_page.dart**

```dart
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../core/providers/auth_provider.dart';
import '../../core/theme/colors.dart';
import '../../data/services/ego_client.dart';
import '../../data/generated/api.pbgrpc.dart' as grpc;

class SettingPage extends ConsumerStatefulWidget {
  const SettingPage({super.key});

  @override
  ConsumerState<SettingPage> createState() => _SettingPageState();
}

class _SettingPageState extends ConsumerState<SettingPage> {
  grpc.GetProfileRes? _profile;
  bool _loading = true;
  String? _error;

  @override
  void initState() {
    super.initState();
    _loadProfile();
  }

  Future<void> _loadProfile() async {
    try {
      final client = ref.read(EgoClient.provider);
      final res = await client.getProfile(ref);
      setState(() {
        _profile = res;
        _loading = false;
      });
    } catch (e) {
      setState(() {
        _error = e.toString();
        _loading = false;
      });
    }
  }

  void _logout() {
    ref.read(authProvider.notifier).logout();
    context.go('/login');
  }

  String _maskPhone(String phone) {
    if (phone.length < 7) return phone;
    return '${phone.substring(0, 3)}****${phone.substring(phone.length - 4)}';
  }

  String _formatDate(int unixMs) {
    final dt = DateTime.fromMillisecondsSinceEpoch(unixMs);
    return '${dt.year}/${dt.month.toString().padLeft(2, '0')}/${dt.day.toString().padLeft(2, '0')}';
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: const Color(0xFF0D0D14),
      appBar: AppBar(
        backgroundColor: Colors.transparent,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back, color: Color(0xFFCCA880)),
          onPressed: () => context.pop(),
        ),
        title: const Text(
          '设置',
          style: TextStyle(
            color: Color(0xFFCCA880),
            fontSize: 18,
            fontWeight: FontWeight.w500,
          ),
        ),
        centerTitle: true,
      ),
      body: _loading
          ? const Center(
              child: CircularProgressIndicator(color: Color(0xFFCCA880)),
            )
          : _error != null
              ? Center(
                  child: Column(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      const Icon(Icons.error_outline,
                          color: Color(0xFF5A5A70), size: 48),
                      const SizedBox(height: 16),
                      Text(
                        '加载失败',
                        style: TextStyle(color: Color(0xFF5A5A70)),
                      ),
                    ],
                  ),
                )
              : Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 24),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      const SizedBox(height: 32),
                      const Text(
                        '账号信息',
                        style: TextStyle(
                          color: Color(0xFF5A5A70),
                          fontSize: 13,
                        ),
                      ),
                      const SizedBox(height: 16),
                      _infoRow(
                        '手机号',
                        _maskPhone(_profile!.phone),
                      ),
                      const SizedBox(height: 12),
                      _infoRow(
                        '注册时间',
                        _formatDate(_profile!.createdAt.toInt()),
                      ),
                      const SizedBox(height: 48),
                      SizedBox(
                        width: double.infinity,
                        child: TextButton(
                          onPressed: _logout,
                          style: TextButton.styleFrom(
                            padding: const EdgeInsets.symmetric(vertical: 14),
                            shape: RoundedRectangleBorder(
                              borderRadius: BorderRadius.circular(12),
                              side: const BorderSide(
                                color: Color(0xFFE53935),
                                width: 0.5,
                              ),
                            ),
                          ),
                          child: const Text(
                            '退出登录',
                            style: TextStyle(
                              color: Color(0xFFE53935),
                              fontSize: 16,
                              fontWeight: FontWeight.w500,
                            ),
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
    );
  }

  Widget _infoRow(String label, String value) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Text(
          label,
          style: const TextStyle(
            color: Color(0xFF5A5A70),
            fontSize: 15,
          ),
        ),
        Text(
          value,
          style: const TextStyle(
            color: Color(0xFFE8D5B0),
            fontSize: 15,
          ),
        ),
      ],
    );
  }
}
```

---

### Task 11: Frontend — Router + Client + Entry

**Files:**
- Modify: `client/lib/core/router/router.dart`
- Modify: `client/lib/data/services/ego_client.dart`
- Modify: `client/lib/shared/widgets/app_shell.dart`

- [ ] **Step 1: Add /setting route to router.dart**

In `router.dart`, add the setting import (after other feature imports, line 10):

```dart
import '../../features/setting/setting_page.dart';
```

Add a new `GoRoute` before the `StatefulShellRoute.indexedStack` (after line 40, after the `/onboard` route):

```dart
GoRoute(
  path: '/setting',
  builder: (context, state) => const SettingPage(),
),
```

- [ ] **Step 2: Add getProfile to ego_client.dart**

In `client/lib/data/services/ego_client.dart`, add after the auth section (after `resetPassword` method, around line 64):

```dart
// ─── Setting ───────────────────────────────────

Future<grpc.GetProfileRes> getProfile(Ref ref) async {
  final req = grpc.GetProfileReq();
  return _stub.getProfile(req, options: _withAuth(ref));
}
```

- [ ] **Step 3: Add settings icon to AppShell**

In `client/lib/shared/widgets/app_shell.dart`, modify the Scaffold to include the settings icon. The pages use custom backgrounds and no AppBar, so add a Positioned settings button:

Replace the `Scaffold` widget content — wrap `navigationShell` in a `Stack` and add a positioned settings icon:

```dart
@override
Widget build(BuildContext context, WidgetRef ref) {
  final tabIndex = ref.watch(tabProvider);

  return Scaffold(
    body: Stack(
      children: [
        navigationShell,
        Positioned(
          top: MediaQuery.of(context).padding.top + 8,
          left: 4,
          child: IconButton(
            icon: const Icon(
              Icons.settings_outlined,
              color: Color(0xFF5A5A70),
              size: 22,
            ),
            onPressed: () => context.go('/setting'),
          ),
        ),
      ],
    ),
    bottomNavigationBar: ClipRect(
      // ... rest remains unchanged
```

The existing `bottomNavigationBar` content stays exactly the same as before.

- [ ] **Step 4: Run Flutter static check**

```bash
cd client && flutter analyze
```

Expected: No new errors.

---

### Task 12: Go Unit Test

**Files:**
- Create: `server/internal/setting/adapter/grpc/handler_test.go`

- [ ] **Step 1: Create handler_test.go**

```go
package grpc_test

import (
	"context"
	"testing"

	settinggrpc "ego-server/internal/setting/adapter/grpc"
	settingapp "ego-server/internal/setting/app"
	settingdomain "ego-server/internal/setting/domain"

	pb "ego-server/proto/ego"
)

// stubUserReader is a test stub that implements domain.UserReader.
type stubUserReader struct {
	user *settingdomain.UserInfo
	err  error
}

func (s *stubUserReader) FindByID(_ context.Context, _ string) (*settingdomain.UserInfo, error) {
	return s.user, s.err
}

func TestGetProfile_WithValidUserID(t *testing.T) {
	reader := &stubUserReader{
		user: &settingdomain.UserInfo{Phone: "13812348888", CreatedAt: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)},
	}
	uc := settingapp.NewGetProfileUseCase(reader)
	h := settinggrpc.NewHandler(uc)

	ctx := context.WithValue(context.Background(), "user_id", "abc-123")
	resp, err := h.GetProfile(ctx, &pb.GetProfileReq{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Phone != "13812348888" {
		t.Errorf("expected phone 13812348888, got %s", resp.Phone)
	}
	if resp.CreatedAt == 0 {
		t.Error("created_at should not be empty")
	}
}

func TestGetProfile_UserNotFound(t *testing.T) {
	reader := &stubUserReader{err: settingdomain.ErrUserNotFound}
	uc := settingapp.NewGetProfileUseCase(reader)
	h := settinggrpc.NewHandler(uc)

	ctx := context.WithValue(context.Background(), "user_id", "nonexistent")
	_, err := h.GetProfile(ctx, &pb.GetProfileReq{})
	if err == nil {
		t.Fatal("expected error for user not found, got nil")
	}
}

func TestGetProfile_MissingUserID(t *testing.T) {
	reader := &stubUserReader{}
	uc := settingapp.NewGetProfileUseCase(reader)
	h := settinggrpc.NewHandler(uc)

	ctx := context.Background() // no user_id
	_, err := h.GetProfile(ctx, &pb.GetProfileReq{})
	if err == nil {
		t.Fatal("expected error for missing user_id, got nil")
	}
}
```

- [ ] **Step 2: Run unit tests**

```bash
cd server && go test ./internal/setting/adapter/grpc/ -v -count=1
```

Expected: 3 tests PASS.

---

### Task 13: Final Verification & Commit

- [ ] **Step 1: Full server build**

```bash
cd server && go build ./...
```

- [ ] **Step 2: Full Go tests**

```bash
cd server && go test ./internal/setting/... -v -count=1
```

- [ ] **Step 3: Full Flutter analyze**

```bash
cd client && flutter analyze
```

- [ ] **Step 4: Single commit for all feature changes**

```bash
git add -A
git commit -m "feat(setting): add setting module with account info and logout

- Proto: add GetProfile RPC with phone + created_at response
- sqlc: add GetUserByID query for user lookup
- New setting DDD module (domain/app/adapter grpc+postgres)
- Bootstrap wiring: setting handler + composite routing
- Frontend: /setting page with masked phone, register date, red logout button
- Settings entry via top-left icon in AppShell

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```
```

---

## Self-Review

1. **Spec coverage:** All design sections covered — Proto (Task 1-2), Backend DDD (Task 4-9), Frontend (Task 10-11), Test (Task 12)
2. **Placeholder scan:** No TBD/TODO/vague steps. All code is explicit.
3. **Type consistency:** `GetProfileReq{}` and `GetProfileRes{Phone, CreatedAt}` used consistently from proto → handler → test. Handler returns `settinggrpc.Handler` implementing `pb.EgoServer` for composite compatibility.
