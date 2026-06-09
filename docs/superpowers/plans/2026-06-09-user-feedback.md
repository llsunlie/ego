# User Feedback Channel Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a text-only user feedback channel to the setting module — from proto to frontend.

**Architecture:** New `SubmitFeedback` RPC in setting service. Backend follows existing DDD three-layer pattern (domain/app/adapter). Frontend adds a new feedback page with a simple text form, accessible from the setting page's "About" section.

**Tech Stack:** Go (DDD), gRPC (proto), PostgreSQL + sqlc, Flutter (Riverpod + GoRouter)

**Commit Strategy:** Per ego-feature rules, ALL code changes accumulate in working tree. Only proto/sqlc generation steps may be committed early (pure generated code, no side effects). The single final commit happens after ALL checks pass + manual test confirmed.

---

### Task 1: Proto Contract

**Files:**
- Modify: `proto/ego/api.proto:22-23`
- Generate: `server/proto/ego/` (via `make proto-go`)
- Generate: `client/lib/data/generated/` (via `make proto-dart`)

- [ ] **Step 1: Add SubmitFeedback RPC and messages to proto**

In `proto/ego/api.proto`, after the `GetProfile` RPC (line 23), add:

```protobuf
  rpc SubmitFeedback(SubmitFeedbackReq) returns (SubmitFeedbackRes);
```

After the `GetProfileRes` message (line 117), add:

```protobuf
message SubmitFeedbackReq {
  string content = 1;  // 反馈文本内容
}

message SubmitFeedbackRes {
  string id         = 1;  // 反馈记录 ID
  int64  created_at = 2;  // 提交时间 unix timestamp ms
}
```

- [ ] **Step 2: Regenerate proto stubs**

```bash
make proto-go && make proto-dart
```

Expected: Both commands exit 0. Check `git diff --stat` to confirm generated files are updated.

- [ ] **Step 3: Stage proto changes (generated code — safe to stage early)**

```bash
git add proto/ego/api.proto server/proto/ego/ client/lib/data/generated/
```

---

### Task 2: Database Migration

**Files:**
- Create: `server/internal/platform/postgres/migrations/011_feedback.sql`

- [ ] **Step 1: Create migration SQL**

```sql
CREATE TABLE IF NOT EXISTS feedbacks (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id),
    content    TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_feedbacks_user_id ON feedbacks(user_id);
```

- [ ] **Step 2: Apply migration**

```bash
# Migration applies automatically on server start via golang-migrate.
# Verify by starting the server and checking the table exists:
# psql -h localhost -U ego -d ego -c "\d feedbacks"
```

- [ ] **Step 3: Verify migration applies**

```bash
# Migration applies automatically on server start via golang-migrate.
# After server start, verify:
# psql -h localhost -U ego -d ego -c "\d feedbacks"
```

---

### Task 3: sqlc Query

**Files:**
- Create: `server/internal/platform/postgres/queries/feedbacks.sql`
- Generate: `server/internal/platform/postgres/sqlc/feedbacks.sql.go` (via `make sqlc`)

- [ ] **Step 1: Create feedbacks.sql query file**

```sql
-- name: InsertFeedback :exec
INSERT INTO feedbacks (id, user_id, content, created_at) VALUES ($1, $2, $3, $4);
```

- [ ] **Step 2: Regenerate sqlc**

```bash
make sqlc
```

Expected: Exit 0. New file `server/internal/platform/postgres/sqlc/feedbacks.sql.go` generated.

- [ ] **Step 3: Stage sqlc changes**

```bash
git add server/internal/platform/postgres/queries/feedbacks.sql server/internal/platform/postgres/sqlc/feedbacks.sql.go
```

---

### Task 4: Domain Layer

**Files:**
- Create: `server/internal/setting/domain/feedback.go`
- Modify: `server/internal/setting/domain/ports.go`
- Modify: `server/internal/setting/domain/errors.go`

- [ ] **Step 1: Create feedback entity**

New file `server/internal/setting/domain/feedback.go`:

```go
package domain

import "time"

// Feedback represents a user-submitted feedback.
type Feedback struct {
	ID        string
	UserID    string
	Content   string
	CreatedAt time.Time
}
```

- [ ] **Step 2: Add FeedbackWriter interface to ports.go**

In `server/internal/setting/domain/ports.go`, append:

```go
// FeedbackWriter persists user feedback.
type FeedbackWriter interface {
	Save(ctx context.Context, feedback *Feedback) error
}
```

Add `"context"` to imports.

- [ ] **Step 3: Add ErrFeedbackEmpty to errors.go**

In `server/internal/setting/domain/errors.go`, append:

```go
var ErrFeedbackEmpty = errors.New("feedback content is empty")
```

- [ ] **Step 4: Compile check**

```bash
go build ./internal/setting/...
```

Expected: Exit 0 (no new code references these yet, just type definitions).

- [ ] **Step 5: Verify compilation**

```bash
go build ./internal/setting/...
```

Expected: Exit 0.

---

### Task 5: ID Generator

**Files:**
- Create: `server/internal/setting/adapter/id/uuid.go`

- [ ] **Step 1: Create UUID generator**

New file `server/internal/setting/adapter/id/uuid.go`:

```go
package id

import "github.com/google/uuid"

// UUIDGenerator generates UUID v4 strings.
type UUIDGenerator struct{}

func NewUUIDGenerator() UUIDGenerator {
	return UUIDGenerator{}
}

func (UUIDGenerator) New() string {
	return uuid.New().String()
}
```

- [ ] **Step 2: Compile check**

```bash
go build ./internal/setting/adapter/id/
```

Expected: Exit 0.

---

### Task 6: Application Layer

**Files:**
- Create: `server/internal/setting/app/feedback.go`

- [ ] **Step 1: Create SubmitFeedbackUseCase**

New file `server/internal/setting/app/feedback.go`:

```go
package app

import (
	"context"
	"strings"
	"time"

	"ego-server/internal/setting/domain"
)

// IDGenerator creates unique identifiers.
type IDGenerator interface {
	New() string
}

// SubmitFeedbackUseCase handles user feedback submission.
type SubmitFeedbackUseCase struct {
	feedbackWriter domain.FeedbackWriter
	idGenerator    IDGenerator
}

func NewSubmitFeedbackUseCase(
	feedbackWriter domain.FeedbackWriter,
	idGenerator IDGenerator,
) *SubmitFeedbackUseCase {
	return &SubmitFeedbackUseCase{
		feedbackWriter: feedbackWriter,
		idGenerator:    idGenerator,
	}
}

// FeedbackResult holds the result of a successful feedback submission.
type FeedbackResult struct {
	ID        string
	CreatedAt int64 // unix timestamp ms
}

// Submit validates and persists user feedback.
func (uc *SubmitFeedbackUseCase) Submit(ctx context.Context, userID, content string) (*FeedbackResult, error) {
	if strings.TrimSpace(content) == "" {
		return nil, domain.ErrFeedbackEmpty
	}

	fb := &domain.Feedback{
		ID:        uc.idGenerator.New(),
		UserID:    userID,
		Content:   strings.TrimSpace(content),
		CreatedAt: time.Now(),
	}

	if err := uc.feedbackWriter.Save(ctx, fb); err != nil {
		return nil, err
	}

	return &FeedbackResult{
		ID:        fb.ID,
		CreatedAt: fb.CreatedAt.UnixMilli(),
	}, nil
}
```

- [ ] **Step 2: Compile check**

```bash
go build ./internal/setting/...
```

Expected: Exit 0.

---

### Task 7: Postgres Adapter

**Files:**
- Create: `server/internal/setting/adapter/postgres/feedback_writer.go`

- [ ] **Step 1: Create FeedbackWriter implementation**

New file `server/internal/setting/adapter/postgres/feedback_writer.go`:

```go
package postgres

import (
	"context"
	"time"

	"ego-server/internal/platform/postgres/sqlc"
	"ego-server/internal/setting/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// FeedbackWriter implements domain.FeedbackWriter using sqlc.
type FeedbackWriter struct {
	queries *sqlc.Queries
}

func NewFeedbackWriter(queries *sqlc.Queries) *FeedbackWriter {
	return &FeedbackWriter{queries: queries}
}

func (w *FeedbackWriter) Save(ctx context.Context, fb *domain.Feedback) error {
	id, err := uuid.Parse(fb.ID)
	if err != nil {
		return err
	}
	userID, err := uuid.Parse(fb.UserID)
	if err != nil {
		return err
	}

	return w.queries.InsertFeedback(ctx, sqlc.InsertFeedbackParams{
		ID:        uuidToPGType(id),
		UserID:    uuidToPGType(userID),
		Content:   fb.Content,
		CreatedAt: pgtype.Timestamptz{Time: fb.CreatedAt, Valid: true},
	})
}

// uuidToPGType converts a [16]byte array to pgtype.UUID.
func uuidToPGType(id uuid.UUID) pgtype.UUID {
	var arr [16]byte
	copy(arr[:], id[:])
	return pgtype.UUID{Bytes: arr, Valid: true}
}
```

Wait — the existing `user_reader.go` already uses a similar `uuidToPGType`-style conversion inline. To be consistent and avoid duplication, add the helper to this file only (private to the package). Actually, looking at `user_reader.go` it does the conversion inline without a helper. Let me follow that pattern exactly:

```go
package postgres

import (
	"context"

	"ego-server/internal/platform/postgres/sqlc"
	"ego-server/internal/setting/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// FeedbackWriter implements domain.FeedbackWriter using sqlc.
type FeedbackWriter struct {
	queries *sqlc.Queries
}

func NewFeedbackWriter(queries *sqlc.Queries) *FeedbackWriter {
	return &FeedbackWriter{queries: queries}
}

func (w *FeedbackWriter) Save(ctx context.Context, fb *domain.Feedback) error {
	id, err := uuid.Parse(fb.ID)
	if err != nil {
		return err
	}
	userID, err := uuid.Parse(fb.UserID)
	if err != nil {
		return err
	}

	var idArr, userIDArr [16]byte
	copy(idArr[:], id[:])
	copy(userIDArr[:], userID[:])

	return w.queries.InsertFeedback(ctx, sqlc.InsertFeedbackParams{
		ID:        pgtype.UUID{Bytes: idArr, Valid: true},
		UserID:    pgtype.UUID{Bytes: userIDArr, Valid: true},
		Content:   fb.Content,
		CreatedAt: pgtype.Timestamptz{Time: fb.CreatedAt, Valid: true},
	})
}
```

- [ ] **Step 2: Compile check**

```bash
go build ./internal/setting/...
```

Expected: Exit 0.

---

### Task 8: gRPC Handler

**Files:**
- Modify: `server/internal/setting/adapter/grpc/handler.go`

- [ ] **Step 1: Add SubmitFeedback to handler**

In `server/internal/setting/adapter/grpc/handler.go`, modify the `Handler` struct and `NewHandler` to accept the new use case, and add the `SubmitFeedback` method:

```go
// Handler implements pb.EgoServer for the setting module.
type Handler struct {
	pb.UnimplementedEgoServer
	getProfile     *app.GetProfileUseCase
	submitFeedback *app.SubmitFeedbackUseCase
}

func NewHandler(getProfile *app.GetProfileUseCase, submitFeedback *app.SubmitFeedbackUseCase) *Handler {
	return &Handler{getProfile: getProfile, submitFeedback: submitFeedback}
}

// Add this method after GetProfile:

func (h *Handler) SubmitFeedback(ctx context.Context, req *pb.SubmitFeedbackReq) (*pb.SubmitFeedbackRes, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "未登录")
	}

	result, err := h.submitFeedback.Submit(ctx, userID, req.Content)
	if err != nil {
		return nil, mapError(err)
	}

	return &pb.SubmitFeedbackRes{
		Id:        result.ID,
		CreatedAt: result.CreatedAt,
	}, nil
}
```

Also update `mapError` to handle `ErrFeedbackEmpty`:

```go
func mapError(err error) error {
	if errors.Is(err, domain.ErrUserNotFound) {
		return status.Error(codes.NotFound, "用户不存在")
	}
	if errors.Is(err, domain.ErrFeedbackEmpty) {
		return status.Error(codes.InvalidArgument, "反馈内容不能为空")
	}
	return status.Error(codes.Internal, err.Error())
}
```

- [ ] **Step 2: Compile check** (will fail until module.go is updated in Task 9)

```bash
go build ./internal/setting/...
```

- [ ] **Step 3: Commit with module.go changes (Task 9)**

Hold — commit together with Task 9 since they're interdependent.

---

### Task 9: Module Wiring

**Files:**
- Modify: `server/internal/setting/module.go`

- [ ] **Step 1: Wire the new use case into module.go**

Replace `server/internal/setting/module.go` with:

```go
package setting

import (
	settinggrpc "ego-server/internal/setting/adapter/grpc"
	settingid "ego-server/internal/setting/adapter/id"
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
	feedbackWriter := settingpostgres.NewFeedbackWriter(queries)
	ids := settingid.NewUUIDGenerator()

	getProfileUseCase := settingapp.NewGetProfileUseCase(userReader)
	submitFeedbackUseCase := settingapp.NewSubmitFeedbackUseCase(feedbackWriter, ids)

	return settinggrpc.NewHandler(getProfileUseCase, submitFeedbackUseCase)
}
```

- [ ] **Step 2: Compile check**

```bash
go build ./internal/setting/...
```

Expected: Exit 0.

- [ ] **Step 3: Verify compilation**

```bash
go build ./internal/setting/... && go build ./internal/bootstrap/...
```

Expected: Both exit 0.

---

### Task 10: Composite Handler

**Files:**
- Modify: `server/internal/bootstrap/composite.go`

- [ ] **Step 1: Add SubmitFeedback delegation**

After the `GetProfile` method (line 179), add:

```go
func (h *EgoHandler) SubmitFeedback(ctx context.Context, req *pb.SubmitFeedbackReq) (*pb.SubmitFeedbackRes, error) {
	logger := logging.FromContext(ctx)
	logger.InfoContext(ctx, "SubmitFeedback: request", "req", fmt.Sprintf("%+v", req))
	res, err := h.setting.SubmitFeedback(ctx, req)
	logger.InfoContext(ctx, "SubmitFeedback: response", "res", fmt.Sprintf("%+v", res), "error", err)
	return res, err
}
```

- [ ] **Step 2: Compile check**

```bash
go build ./internal/bootstrap/...
go build ./cmd/ego/
```

Expected: Both exit 0.

- [ ] **Step 3: Verify full server build**

```bash
go build ./cmd/ego/
```

Expected: Exit 0.

---

### Task 11: Go Tests

**Files:**
- Create: `server/internal/setting/adapter/grpc/handler_test.go`

- [ ] **Step 1: Write unit tests for SubmitFeedback handler**

Create `server/internal/setting/adapter/grpc/handler_test.go`:

```go
package grpc_test

import (
	"context"
	"testing"

	"ego-server/internal/setting/adapter/grpc"
	settingapp "ego-server/internal/setting/app"
	settingdomain "ego-server/internal/setting/domain"

	pb "ego-server/proto/ego"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// stubFeedbackWriter is a test double for domain.FeedbackWriter.
type stubFeedbackWriter struct {
	saved *settingdomain.Feedback
	err   error
}

func (s *stubFeedbackWriter) Save(ctx context.Context, fb *settingdomain.Feedback) error {
	if s.err != nil {
		return s.err
	}
	s.saved = fb
	return nil
}

// stubIDGenerator returns a fixed ID for testing.
type stubIDGenerator struct{}

func (stubIDGenerator) New() string { return "test-id-123" }

func TestSubmitFeedback_Success(t *testing.T) {
	writer := &stubFeedbackWriter{}
	ids := stubIDGenerator{}
	uc := settingapp.NewSubmitFeedbackUseCase(writer, ids)
	h := grpc.NewHandler(nil, uc)

	ctx := context.WithValue(context.Background(), "user_id", "user-1")
	res, err := h.SubmitFeedback(ctx, &pb.SubmitFeedbackReq{Content: "Great app!"})

	require.NoError(t, err)
	assert.Equal(t, "test-id-123", res.Id)
	assert.NotZero(t, res.CreatedAt)
	assert.Equal(t, "Great app!", writer.saved.Content)
	assert.Equal(t, "user-1", writer.saved.UserID)
}

func TestSubmitFeedback_EmptyContent(t *testing.T) {
	writer := &stubFeedbackWriter{}
	ids := stubIDGenerator{}
	uc := settingapp.NewSubmitFeedbackUseCase(writer, ids)
	h := grpc.NewHandler(nil, uc)

	ctx := context.WithValue(context.Background(), "user_id", "user-1")
	_, err := h.SubmitFeedback(ctx, &pb.SubmitFeedbackReq{Content: "   "})

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestSubmitFeedback_Unauthenticated(t *testing.T) {
	writer := &stubFeedbackWriter{}
	ids := stubIDGenerator{}
	uc := settingapp.NewSubmitFeedbackUseCase(writer, ids)
	h := grpc.NewHandler(nil, uc)

	_, err := h.SubmitFeedback(context.Background(), &pb.SubmitFeedbackReq{Content: "test"})

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}
```

- [ ] **Step 2: Run Go tests + vet**

```bash
go test ./internal/setting/... -v -count=1 && go vet ./internal/setting/...
```

Expected: All tests PASS, no vet issues.

---

### Task 12: Smoke Test

**Files:**
- Modify: `smoke.sh`

- [ ] **Step 1: Add SubmitFeedback smoke test to smoke.sh**

Add after the GetProfile test section:

```bash
# ─── SubmitFeedback ──────────────────────────────
echo ""
echo "▶ SubmitFeedback (normal)"
grpcurl \
  -plaintext \
  -H "authorization: Bearer $TOKEN" \
  -d '{"content":"这是一条来自 smoke 测试的反馈"}' \
  "$HOST" \
  ego.Ego/SubmitFeedback \
  | jq -e '.id != "" and .createdAt != "0"'

echo ""
echo "▶ SubmitFeedback (no token → UNAUTHENTICATED)"
grpcurl \
  -plaintext \
  -d '{"content":"test"}' \
  "$HOST" \
  ego.Ego/SubmitFeedback 2>&1 \
  | grep -i "unauthenticated\|Unauthenticated"

echo ""
echo "▶ SubmitFeedback (empty content → INVALID_ARGUMENT)"
grpcurl \
  -plaintext \
  -H "authorization: Bearer $TOKEN" \
  -d '{"content":"  "}' \
  "$HOST" \
  ego.Ego/SubmitFeedback 2>&1 \
  | grep -i "invalid\|InvalidArgument"
```

- [ ] **Step 2: Run smoke test**

```bash
bash smoke.sh
```

Expected: All assertions pass, including new SubmitFeedback tests.

---

### Task 13: EgoClient — Dart RPC Method

**Files:**
- Modify: `client/lib/data/services/ego_client.dart`

- [ ] **Step 1: Add submitFeedback method**

After the `getProfile` method (line 72), add:

```dart
  Future<grpc.SubmitFeedbackRes> submitFeedback(
    Ref ref, {
    required String content,
  }) async {
    final req = grpc.SubmitFeedbackReq(content: content);
    return _stub.submitFeedback(req, options: _withAuth(ref));
  }
```

- [ ] **Step 2: Verify Flutter analysis**

```bash
cd client && flutter analyze
```

Expected: No issues found.

---

### Task 14: Feedback Page (Frontend)

**Files:**
- Create: `client/lib/features/setting/feedback_page.dart`

- [ ] **Step 1: Create FeedbackPage**

New file `client/lib/features/setting/feedback_page.dart`:

```dart
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import '../../core/theme/colors.dart';
import '../../data/services/ego_client.dart';
import '../now/widgets/starry_background.dart';

enum _FeedbackState { idle, submitting, success, error }

class FeedbackPage extends ConsumerStatefulWidget {
  const FeedbackPage({super.key});

  @override
  ConsumerState<FeedbackPage> createState() => _FeedbackPageState();
}

class _FeedbackPageState extends ConsumerState<FeedbackPage> {
  final _controller = TextEditingController();
  _FeedbackState _state = _FeedbackState.idle;
  String? _errorMsg;

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  Future<void> _submit() async {
    final content = _controller.text.trim();
    if (content.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('请输入反馈内容', style: TextStyle(color: Colors.white)),
          backgroundColor: AppColors.surface,
          behavior: SnackBarBehavior.floating,
        ),
      );
      return;
    }

    setState(() {
      _state = _FeedbackState.submitting;
      _errorMsg = null;
    });

    try {
      final client = ref.read(EgoClient.provider);
      await client.submitFeedback(ref, content: content);
      setState(() => _state = _FeedbackState.success);

      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(
          content: Text('感谢你的反馈！', style: TextStyle(color: Colors.white)),
          backgroundColor: AppColors.surface,
          behavior: SnackBarBehavior.floating,
        ),
      );
      context.pop();
    } catch (e) {
      setState(() {
        _state = _FeedbackState.error;
        _errorMsg = e.toString();
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: AppColors.darkBg,
      appBar: AppBar(
        backgroundColor: Colors.transparent,
        elevation: 0,
        leading: IconButton(
          icon: const Icon(Icons.arrow_back, color: AppColors.gold),
          onPressed: () => context.pop(),
        ),
        title: const Text(
          '用户反馈',
          style: TextStyle(
            color: AppColors.gold,
            fontSize: 18,
            fontWeight: FontWeight.w500,
          ),
        ),
        centerTitle: true,
      ),
      body: Stack(
        children: [
          const StarryBackground(),
          Padding(
            padding: const EdgeInsets.all(24),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.stretch,
              children: [
                const Text(
                  '你的建议将帮助我们做得更好',
                  style: TextStyle(
                    color: AppColors.textHint,
                    fontSize: 14,
                  ),
                  textAlign: TextAlign.center,
                ),
                const SizedBox(height: 24),
                Expanded(
                  child: TextField(
                    controller: _controller,
                    maxLines: null,
                    expands: true,
                    textAlignVertical: TextAlignVertical.top,
                    style: const TextStyle(
                      color: AppColors.textPrimary,
                      fontSize: 15,
                    ),
                    decoration: InputDecoration(
                      hintText: '请输入反馈内容...',
                      hintStyle: const TextStyle(
                        color: AppColors.textHint,
                        fontSize: 15,
                      ),
                      filled: true,
                      fillColor: AppColors.surface,
                      border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(12),
                        borderSide: BorderSide.none,
                      ),
                      contentPadding: const EdgeInsets.all(16),
                    ),
                  ),
                ),
                if (_errorMsg != null) ...[
                  const SizedBox(height: 12),
                  Text(
                    _errorMsg!,
                    style: const TextStyle(
                      color: Color(0xFFE53935),
                      fontSize: 13,
                    ),
                    textAlign: TextAlign.center,
                  ),
                ],
                const SizedBox(height: 24),
                SizedBox(
                  height: 48,
                  child: ElevatedButton(
                    onPressed: _state == _FeedbackState.submitting ? null : _submit,
                    style: ElevatedButton.styleFrom(
                      backgroundColor: AppColors.gold,
                      foregroundColor: AppColors.darkBg,
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(12),
                      ),
                      disabledBackgroundColor: AppColors.gold.withOpacity(0.5),
                    ),
                    child: _state == _FeedbackState.submitting
                        ? const SizedBox(
                            width: 20,
                            height: 20,
                            child: CircularProgressIndicator(
                              strokeWidth: 2,
                              color: AppColors.darkBg,
                            ),
                          )
                        : const Text(
                            '提交反馈',
                            style: TextStyle(
                              fontSize: 16,
                              fontWeight: FontWeight.w500,
                            ),
                          ),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }
}
```

- [ ] **Step 2: Flutter analyze**

```bash
cd client && flutter analyze
```

Expected: No issues found.

---

### Task 15: Setting Page — Add Feedback Entry

**Files:**
- Modify: `client/lib/features/setting/setting_page.dart`

- [ ] **Step 1: Add feedback row in "About" section**

After the privacy policy row (line 223, the `_settingRow` for privacy), add:

```dart
                      _rowDivider(),
                      _settingRow(
                        icon: Icons.feedback_outlined,
                        label: '用户反馈',
                        showArrow: true,
                        onTap: () => context.push('/feedback'),
                      ),
```

- [ ] **Step 2: Flutter analyze**

```bash
cd client && flutter analyze
```

Expected: No issues found.

---

### Task 16: Router — Register /feedback Route

**Files:**
- Modify: `client/lib/core/router/router.dart`

- [ ] **Step 1: Add /feedback route and import**

In imports (`client/lib/core/router/router.dart`), after the setting page import (line 12), add:

```dart
import '../../features/setting/feedback_page.dart';
```

In the routes list, after the `/setting` GoRoute (line 52), add:

```dart
      GoRoute(
        path: '/feedback',
        builder: (context, state) => const FeedbackPage(),
      ),
```

- [ ] **Step 2: Flutter analyze**

```bash
cd client && flutter analyze
```

Expected: No issues found.

---

### Task 17: Final Verification

**All code committed. Run full verification suite before merge.**

- [ ] **Step 1: Go tests**

```bash
go test ./internal/setting/... -v -count=1
```

Expected: All tests PASS.

- [ ] **Step 2: Go vet**

```bash
go vet ./internal/setting/...
```

Expected: No issues.

- [ ] **Step 3: Full Go build**

```bash
go build ./...
```

Expected: Exit 0.

- [ ] **Step 4: Flutter analyze**

```bash
cd client && flutter analyze
```

Expected: No issues found.

- [ ] **Step 5: Smoke test**

```bash
bash smoke.sh
```

Expected: All assertions pass, including new SubmitFeedback tests.

- [ ] **Step 6: sqlc side-effect check**

```bash
make sqlc
git diff --stat
```

Verify only `server/internal/platform/postgres/sqlc/feedbacks.sql.go` is new — no unrelated sqlc changes. If other sqlc files changed, revert them with `git checkout -- <unrelated-file>`.

- [ ] **Step 7: Manual test checklist (user performs on device)**

| # | 场景 | 步骤 | 预期 |
|---|------|------|------|
| 1 | 正常反馈 | 设置页 → 用户反馈 → 输入内容 → 提交 | SnackBar「感谢你的反馈」→ 自动返回设置页 |
| 2 | 空内容 | 设置页 → 用户反馈 → 不输入 → 提交 | SnackBar「请输入反馈内容」→ 留在当前页 |
| 3 | 空格内容 | 输入纯空格 → 提交 | SnackBar「请输入反馈内容」 |
| 4 | 数据持久化 | 正常提交后查库 `SELECT * FROM feedbacks` | 记录存在，user_id 正确 |
| 5 | 未登录 | 清除 token → grpcurl SubmitFeedback | UNAUTHENTICATED |
| 6 | 页面返回 | 反馈页点击左上角箭头 | 返回设置页，无异常 |

---

### Task 18: Commit & Push

**ONLY after user confirms manual test pass (Step 7 above).**

- [ ] **Step 1: Single final commit**

```bash
git add -A
git commit -m "feat(setting): add user feedback channel

- Proto: SubmitFeedback RPC + messages
- Backend: domain entity, FeedbackWriter, SubmitFeedbackUseCase, gRPC handler, postgres adapter
- Database: feedbacks table migration + sqlc query
- Frontend: FeedbackPage, setting page entry, /feedback route, EgoClient method

Co-Authored-By: Claude Opus 4.8 <noreply@anthropic.com>"
```

- [ ] **Step 2: Push branch**

```bash
git push origin test
```

- [ ] **Step 3: Create PR**

```bash
gh pr create \
  --base test \
  --head test \
  --title "feat(setting): add user feedback channel" \
  --body "## Changes

- **Proto:** Add \`SubmitFeedback\` RPC with \`SubmitFeedbackReq\`/\`SubmitFeedbackRes\` messages
- **Backend:** Full DDD implementation — domain entity, FeedbackWriter interface, SubmitFeedbackUseCase, gRPC handler, postgres adapter
- **Database:** New \`feedbacks\` table (migration 011)
- **Frontend:** New \`FeedbackPage\` with text input + submit, entry row in setting page's 'About' section, \`/feedback\` route

## How to test

1. Go to Settings → 用户反馈
2. Enter feedback text and submit
3. Verify success snackbar and auto-navigate back
4. Check database for persisted record

🤖 Generated with [Claude Code](https://claude.com/claude-code)"
```
