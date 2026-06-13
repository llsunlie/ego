# Rate Limit Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add per-user_id + per-IP dual-dimension token bucket rate limiting to the ego server gRPC layer.

**Architecture:** New `ratelimit` platform package wraps `golang.org/x/time/rate` token buckets in a `sync.Map`. A gRPC unary interceptor, chained after auth, checks each request against IP and (for authenticated RPCs) user_id buckets. Config exposed via 5 env vars with sensible defaults.

**Tech Stack:** Go `golang.org/x/time/rate`, `sync.Map`, gRPC `ChainUnaryInterceptor`; Flutter Dart `grpc_error_mapper.dart`

---

### Task 1: Add golang.org/x/time dependency + Config fields

**Files:**
- Modify: `server/internal/config/config.go:27-55`, `server/internal/config/config.go:65-99`

- [ ] **Step 1: Add `golang.org/x/time/rate` to go.mod**

```bash
cd server && go get golang.org/x/time/rate
```

Expected: `go.mod` updated with new dependency.

- [ ] **Step 2: Add 5 ratelimit fields to Config struct**

In `server/internal/config/config.go`, add to the `Config` struct:

```go
type Config struct {
	DatabaseURL           string
	JWTSecret             string
	WebPort               string
	WebTLSPort            string
	GRPCPort              string
	WebDir                string
	JWTExpHours           string
	LogLevel              string
	LogFormat             string
	LogOutput             string
	AIAPIKey              string
	AIBaseURL             string
	AIEmbeddingModel      string
	AIEmbeddingAPIKey     string
	AIEmbeddingBaseURL    string
	AIChatModel           string
	AIChatAPIKey          string
	AIChatBaseURL         string
	AliyunAccessKeyID     string
	AliyunAccessKeySecret string
	AliyunSmsSignName     string
	AliyunSmsTemplateCode string
	AliyunSmsCodeLength   string
	AliyunSmsValidTime    string
	AliyunSmsInterval     string
	TLSDomain             string
	CORSAllowedOrigins    string
	// Rate limit
	RateLimitAuthRate     string
	RateLimitAuthBurst    string
	RateLimitPreAuthRate  string
	RateLimitPreAuthBurst string
	RateLimitMaxBuckets   string
}
```

- [ ] **Step 3: Add env var loading in Load()**

In `server/internal/config/config.go`, add to the `Load()` function return:

```go
return &Config{
	// ... existing fields ...
	TLSDomain:             os.Getenv("TLS_DOMAIN"),
	CORSAllowedOrigins:    os.Getenv("CORS_ALLOWED_ORIGINS"),
	// Rate limit
	RateLimitAuthRate:     getEnvDefault("RATELIMIT_AUTH_RATE", "10"),
	RateLimitAuthBurst:    getEnvDefault("RATELIMIT_AUTH_BURST", "20"),
	RateLimitPreAuthRate:  getEnvDefault("RATELIMIT_PREAUTH_RATE", "10"),
	RateLimitPreAuthBurst: getEnvDefault("RATELIMIT_PREAUTH_BURST", "30"),
	RateLimitMaxBuckets:   getEnvDefault("RATELIMIT_MAX_BUCKETS", "500"),
}
```

- [ ] **Step 4: Add getEnvDefault helper**

In `server/internal/config/config.go`, add the helper function (before or after `getEnvWithFallback`):

```go
// getEnvDefault returns os.Getenv(key), or fallback if empty.
func getEnvDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
```

- [ ] **Step 5: Verify config compiles**

```bash
cd server && go build ./internal/config/...
```

Expected: no errors.

---

### Task 2: Create ratelimit package — core Limiter

**Files:**
- Create: `server/internal/platform/ratelimit/ratelimit.go`

- [ ] **Step 1: Create the ratelimit.go file**

```go
package ratelimit

import (
	"context"
	"strconv"
	"sync"
	"time"

	"ego-server/internal/config"
	"ego-server/internal/platform/logging"

	"golang.org/x/time/rate"
)

// Limiter holds token buckets per dimension and enforces rate limits.
type Limiter struct {
	authRate       rate.Limit
	authBurst      int
	preAuthRate    rate.Limit
	preAuthBurst   int
	maxBuckets     int
	buckets        sync.Map // string → *rateLimiterEntry
	preAuthMethods map[string]bool

	stopCh chan struct{}
	doneCh chan struct{}
}

type rateLimiterEntry struct {
	limiter  *rate.Limiter
	lastUsed time.Time
}

// New creates a Limiter from config, starting a background goroutine to
// clean up stale buckets. Call Close() to stop the cleanup goroutine.
func New(cfg *config.Config) *Limiter {
	authRate, _ := strconv.ParseFloat(cfg.RateLimitAuthRate, 64)
	authBurst, _ := strconv.Atoi(cfg.RateLimitAuthBurst)
	preAuthRate, _ := strconv.ParseFloat(cfg.RateLimitPreAuthRate, 64)
	preAuthBurst, _ := strconv.Atoi(cfg.RateLimitPreAuthBurst)
	maxBuckets, _ := strconv.Atoi(cfg.RateLimitMaxBuckets)

	if authRate <= 0 {
		authRate = 10
	}
	if authBurst <= 0 {
		authBurst = 20
	}
	if preAuthRate <= 0 {
		preAuthRate = 10
	}
	if preAuthBurst <= 0 {
		preAuthBurst = 30
	}
	if maxBuckets <= 0 {
		maxBuckets = 500
	}

	l := &Limiter{
		authRate:     rate.Limit(authRate),
		authBurst:    authBurst,
		preAuthRate:  rate.Limit(preAuthRate),
		preAuthBurst: preAuthBurst,
		maxBuckets:   maxBuckets,
		preAuthMethods: map[string]bool{
			"Login":                true,
			"CheckPhone":           true,
			"SendVerificationCode": true,
			"Register":             true,
			"ResetPassword":        true,
		},
		stopCh: make(chan struct{}),
		doneCh: make(chan struct{}),
	}

	go l.cleanupLoop()
	return l
}

// Close stops the background cleanup goroutine.
func (l *Limiter) Close() {
	close(l.stopCh)
	<-l.doneCh
}

// Allow checks whether the request should be allowed.
// fullMethod is the gRPC method path (e.g. "/ego.Ego/Login").
// userID is "" for pre-auth methods.
// ip is the client IP.
// Returns (allowed, deniedDim). deniedDim is "ip" or "user" when denied.
func (l *Limiter) Allow(ctx context.Context, fullMethod string, userID string, ip string) (bool, string) {
	methodName := extractMethodName(fullMethod)
	isPreAuth := l.preAuthMethods[methodName]

	if isPreAuth {
		return l.checkPreAuth(ip), "ip"
	}

	// Authenticated: check IP first, then user.
	if !l.checkAuthIP(ip) {
		return false, "ip"
	}
	if userID != "" && !l.checkAuthUser(userID) {
		return false, "user"
	}
	return true, ""
}

func (l *Limiter) checkPreAuth(ip string) bool {
	key := "preauth:ip:" + ip
	return l.allowKey(key, l.preAuthRate, l.preAuthBurst)
}

func (l *Limiter) checkAuthIP(ip string) bool {
	key := "auth:ip:" + ip
	return l.allowKey(key, l.authRate, l.authBurst)
}

func (l *Limiter) checkAuthUser(userID string) bool {
	key := "auth:user:" + userID
	return l.allowKey(key, l.authRate, l.authBurst)
}

func (l *Limiter) allowKey(key string, r rate.Limit, burst int) bool {
	entryAny, _ := l.buckets.LoadOrStore(key, &rateLimiterEntry{
		limiter:  rate.NewLimiter(r, burst),
		lastUsed: time.Now(),
	})
	entry := entryAny.(*rateLimiterEntry)
	entry.lastUsed = time.Now()

	// Check bucket count: if over limit, trigger immediate cleanup then fail-open.
	if count := l.bucketCount(); count > l.maxBuckets {
		l.cleanupOnce()
		if l.bucketCount() > l.maxBuckets {
			// Fail-open: allow the request, don't store a new bucket.
			l.buckets.Delete(key)
			return true
		}
	}

	return entry.limiter.Allow()
}

// bucketCount returns the approximate number of entries in the map.
func (l *Limiter) bucketCount() int {
	count := 0
	l.buckets.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}

// cleanupLoop runs every minute to remove buckets not used for 5 minutes.
func (l *Limiter) cleanupLoop() {
	defer close(l.doneCh)
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			l.cleanupOnce()
		case <-l.stopCh:
			return
		}
	}
}

func (l *Limiter) cleanupOnce() {
	cutoff := time.Now().Add(-5 * time.Minute)
	l.buckets.Range(func(key, value interface{}) bool {
		entry := value.(*rateLimiterEntry)
		if entry.lastUsed.Before(cutoff) {
			l.buckets.Delete(key)
		}
		return true
	})
}

// extractMethodName extracts the RPC method name from a full gRPC method path.
// e.g. "/ego.Ego/Login" → "Login"
func extractMethodName(fullMethod string) string {
	for i := len(fullMethod) - 1; i >= 0; i-- {
		if fullMethod[i] == '/' {
			return fullMethod[i+1:]
		}
	}
	return fullMethod
}
```

The `logging` import is intentionally unused here but may be used later for logging rate limit decisions. Remove it if the compiler complains:

If `logging` import causes compile error, remove:
```go
"ego-server/internal/platform/logging"
```
from the import block.

- [ ] **Step 2: Verify compilation**

```bash
cd server && go build ./internal/platform/ratelimit/...
```

Expected: no errors.

---

### Task 3: Create ratelimit gRPC interceptor

**Files:**
- Create: `server/internal/platform/ratelimit/interceptor.go`

- [ ] **Step 1: Create interceptor.go**

```go
package ratelimit

import (
	"context"
	"net"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

// UnaryServerInterceptor returns a gRPC unary interceptor that enforces
// rate limits. It must be chained AFTER the auth interceptor so user_id
// is available in the context for authenticated RPCs.
func UnaryServerInterceptor(l *Limiter) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		ip := extractClientIP(ctx)
		userID, _ := ctx.Value("user_id").(string)

		allowed, dim := l.Allow(ctx, info.FullMethod, userID, ip)
		if !allowed {
			return nil, status.Errorf(codes.ResourceExhausted, "rate limit exceeded: %s", dim)
		}

		return handler(ctx, req)
	}
}

// extractClientIP extracts the client IP from gRPC context.
// Checks x-forwarded-for metadata first, then falls back to peer.Addr.
func extractClientIP(ctx context.Context) string {
	// 1. Try x-forwarded-for from metadata.
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		for _, key := range []string{"x-forwarded-for", "X-Forwarded-For"} {
			if vals := md[key]; len(vals) > 0 {
				ip := strings.TrimSpace(vals[0])
				// x-forwarded-for format: "client, proxy1, proxy2"
				if idx := strings.IndexByte(ip, ','); idx >= 0 {
					ip = strings.TrimSpace(ip[:idx])
				}
				if ip != "" {
					return ip
				}
			}
		}
	}

	// 2. Fall back to peer address.
	if p, ok := peer.FromContext(ctx); ok {
		addr := p.Addr.String()
		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			return addr
		}
		return host
	}

	return ""
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd server && go build ./internal/platform/ratelimit/...
```

Expected: no errors.

---

### Task 4: Write ratelimit unit tests

**Files:**
- Create: `server/internal/platform/ratelimit/ratelimit_test.go`

- [ ] **Step 1: Create ratelimit_test.go**

```go
package ratelimit

import (
	"context"
	"testing"
	"time"
)

func TestAllow_PreAuth_FirstRequest(t *testing.T) {
	l := newTestLimiter()
	defer l.Close()

	allowed, _ := l.Allow(context.Background(), "/ego.Ego/Login", "", "1.2.3.4")
	if !allowed {
		t.Fatal("first pre-auth request should be allowed")
	}
}

func TestAllow_PreAuth_ExceedsBurst(t *testing.T) {
	l := newTestLimiter()
	defer l.Close()

	// Burst is 30, so 31st request should fail.
	for i := 0; i < 30; i++ {
		allowed, dim := l.Allow(context.Background(), "/ego.Ego/SendVerificationCode", "", "1.2.3.4")
		if !allowed {
			t.Fatalf("request %d should be allowed, denied: %s", i+1, dim)
		}
	}

	allowed, dim := l.Allow(context.Background(), "/ego.Ego/SendVerificationCode", "", "1.2.3.4")
	if allowed {
		t.Fatal("31st request should be denied (burst exceeded)")
	}
	if dim != "ip" {
		t.Fatalf("expected denied dim 'ip', got '%s'", dim)
	}
}

func TestAllow_Auth_UserAndIPIndependent(t *testing.T) {
	l := newTestLimiter()
	defer l.Close()

	ctx := context.Background()

	// Deplete IP bucket (burst 20).
	for i := 0; i < 20; i++ {
		l.Allow(ctx, "/ego.Ego/GetProfile", "user-a", "1.1.1.1")
	}

	// 21st: IP bucket exhausted for 1.1.1.1.
	allowed, dim := l.Allow(ctx, "/ego.Ego/GetProfile", "user-a", "1.1.1.1")
	if allowed {
		t.Fatal("should be denied by IP exhaustion")
	}
	if dim != "ip" {
		t.Fatalf("expected denied dim 'ip', got '%s'", dim)
	}

	// But same user from different IP should fail on user bucket (also depleted).
	allowed, dim = l.Allow(ctx, "/ego.Ego/GetProfile", "user-a", "2.2.2.2")
	if allowed {
		t.Fatal("should be denied by user exhaustion from different IP")
	}
	if dim != "user" {
		t.Fatalf("expected denied dim 'user', got '%s'", dim)
	}

	// Different user from different IP should be allowed.
	allowed, _ = l.Allow(ctx, "/ego.Ego/GetProfile", "user-b", "2.2.2.2")
	if !allowed {
		t.Fatal("different user from different IP should be allowed")
	}
}

func TestAllow_UnknownMethod(t *testing.T) {
	l := newTestLimiter()
	defer l.Close()

	// Unknown methods (not in preAuthMethods) → treated as auth RPC.
	allowed, dim := l.Allow(context.Background(), "/ego.Ego/SomeNewMethod", "user-x", "1.2.3.4")
	if !allowed {
		t.Fatalf("allowed should be true, got denied: %s", dim)
	}
}

func TestAllow_PreAuth_DifferentIPsIndependent(t *testing.T) {
	l := newTestLimiter()
	defer l.Close()

	// Deplete burst for IP 1.1.1.1
	for i := 0; i < 30; i++ {
		l.Allow(context.Background(), "/ego.Ego/Login", "", "1.1.1.1")
	}

	// 1.1.1.1 should be denied.
	allowed, _ := l.Allow(context.Background(), "/ego.Ego/Login", "", "1.1.1.1")
	if allowed {
		t.Fatal("1.1.1.1 should be denied after depleting burst")
	}

	// 2.2.2.2 should still be allowed (independent bucket).
	allowed, _ = l.Allow(context.Background(), "/ego.Ego/Login", "", "2.2.2.2")
	if !allowed {
		t.Fatal("2.2.2.2 should be allowed (different IP)")
	}
}

func TestCleanup_RemovesStaleBuckets(t *testing.T) {
	l := newTestLimiter()
	defer l.Close()

	// Use a bucket.
	l.Allow(context.Background(), "/ego.Ego/Login", "", "1.2.3.4")

	// Manually age the entry.
	entryAny, ok := l.buckets.Load("preauth:ip:1.2.3.4")
	if !ok {
		t.Fatal("expected bucket to exist")
	}
	entry := entryAny.(*rateLimiterEntry)
	entry.lastUsed = time.Now().Add(-10 * time.Minute)

	// Run cleanup.
	l.cleanupOnce()

	// Bucket should be gone.
	if _, ok := l.buckets.Load("preauth:ip:1.2.3.4"); ok {
		t.Fatal("stale bucket should have been cleaned up")
	}
}

func TestExtractMethodName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/ego.Ego/Login", "Login"},
		{"/ego.Ego/SendVerificationCode", "SendVerificationCode"},
		{"/ego.Ego/GetProfile", "GetProfile"},
		{"NoSlash", "NoSlash"},
	}

	for _, tt := range tests {
		got := extractMethodName(tt.input)
		if got != tt.expected {
			t.Errorf("extractMethodName(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func newTestLimiter() *Limiter {
	return &Limiter{
		authRate:     10,
		authBurst:    20,
		preAuthRate:  10,
		preAuthBurst: 30,
		maxBuckets:   500,
		preAuthMethods: map[string]bool{
			"Login":                true,
			"CheckPhone":           true,
			"SendVerificationCode": true,
			"Register":             true,
			"ResetPassword":        true,
		},
		stopCh: make(chan struct{}),
		doneCh: make(chan struct{}),
	}
}
```

- [ ] **Step 2: Run tests**

```bash
cd server && go test ./internal/platform/ratelimit/ -v -count=1
```

Expected: all tests pass.

---

### Task 5: Wire ratelimit into server bootstrap

**Files:**
- Modify: `server/internal/bootstrap/server.go:51-53`

- [ ] **Step 1: Add ratelimit import and create limiter**

In `server/internal/bootstrap/server.go`, add the import:

```go
import (
	// ... existing imports ...
	"ego-server/internal/platform/ratelimit"
	// ... other imports ...
)
```

- [ ] **Step 2: Chain interceptors in NewServer**

In `server/internal/bootstrap/server.go`, change line 51-53 from:

```go
grpcServer := grpc.NewServer(
    grpc.UnaryInterceptor(auth.UnaryServerInterceptor(p.JWTKey, p.Logger)),
)
```

To:

```go
rateLimiter := ratelimit.New(cfg)
grpcServer := grpc.NewServer(
    grpc.ChainUnaryInterceptor(
        auth.UnaryServerInterceptor(p.JWTKey, p.Logger),
        ratelimit.UnaryServerInterceptor(rateLimiter),
    ),
)
```

Note: The `rateLimiter` var is created here; its `Close()` is handled implicitly — the cleanup goroutine runs until the process exits. If you want explicit lifecycle management, store it on the `Server` struct and call `Close()` in a shutdown hook. For simplicity, the goroutine exits when the process does.

- [ ] **Step 3: Verify compilation**

```bash
cd server && go build ./internal/bootstrap/...
```

Expected: no errors.

---

### Task 6: Server-wide build and test verification

**Files:**
- Test: `server/internal/platform/ratelimit/` (run tests)
- Verify: `server/internal/bootstrap/` (compiles)

- [ ] **Step 1: Run ratelimit unit tests**

```bash
cd server && go test ./internal/platform/ratelimit/ -v -count=1
```

Expected: 7 tests pass.

- [ ] **Step 2: Run go vet on ratelimit package**

```bash
cd server && go vet ./internal/platform/ratelimit/...
```

Expected: no issues.

- [ ] **Step 3: Full server build**

```bash
cd server && go build ./...
```

Expected: no errors.

---

### Task 7: Client — grpc_error_mapper utility

**Files:**
- Create: `client/lib/core/providers/grpc_error_mapper.dart`

- [ ] **Step 1: Create grpc_error_mapper.dart**

```dart
import 'package:grpc/grpc_or_grpcweb.dart';

/// Maps gRPC status codes to user-facing Chinese error messages.
String grpcErrorMessage(GrpcError e) {
  switch (e.code) {
    case StatusCode.resourceExhausted:
      return '请求过于频繁，请稍后再试';
    case StatusCode.unauthenticated:
      return '登录已过期，请重新登录';
    case StatusCode.unavailable:
      return '服务暂不可用，请稍后重试';
    case StatusCode.deadlineExceeded:
      return '请求超时，请检查网络后重试';
    default:
      return e.message ?? '网络错误，请稍后重试';
  }
}

/// Returns true if the error is caused by rate limiting.
bool isRateLimitError(GrpcError e) {
  return e.code == StatusCode.resourceExhausted;
}
```

- [ ] **Step 2: Verify Flutter analysis**

```bash
cd client && flutter analyze lib/core/providers/grpc_error_mapper.dart
```

Expected: no issues.

---

### Task 8: Client — Update login_page.dart for rate limit errors

**Files:**
- Modify: `client/lib/features/login/login_page.dart`

- [ ] **Step 1: Add import**

Add to imports of `client/lib/features/login/login_page.dart`:

```dart
import '../../core/providers/grpc_error_mapper.dart';
```

- [ ] **Step 2: Update _login() error handling (line 127-137)**

Replace:

```dart
    } on GrpcError catch (e) {
      if (mounted) {
        setState(() => _loading = false);
        if (e.code == StatusCode.unauthenticated) {
          _setError('密码错误');
        } else if (e.code == StatusCode.notFound) {
          _setError('用户不存在');
        } else {
          _setError('登录失败，请稍后重试');
        }
      }
```

With:

```dart
    } on GrpcError catch (e) {
      if (mounted) {
        setState(() => _loading = false);
        if (e.code == StatusCode.resourceExhausted) {
          _setError(grpcErrorMessage(e));
        } else if (e.code == StatusCode.unauthenticated) {
          _setError('密码错误');
        } else if (e.code == StatusCode.notFound) {
          _setError('用户不存在');
        } else {
          _setError('登录失败，请稍后重试');
        }
      }
```

- [ ] **Step 3: Update _register() error handling (line 173-183)**

Replace:

```dart
    } on GrpcError catch (e) {
      if (mounted) {
        setState(() => _loading = false);
        if (e.code == StatusCode.unauthenticated) {
          _setError('验证码错误');
        } else if (e.code == StatusCode.alreadyExists) {
          _setError('该手机号已注册，请返回登录');
        } else {
          _setError('注册失败，请稍后重试');
        }
      }
```

With:

```dart
    } on GrpcError catch (e) {
      if (mounted) {
        setState(() => _loading = false);
        if (e.code == StatusCode.resourceExhausted) {
          _setError(grpcErrorMessage(e));
        } else if (e.code == StatusCode.unauthenticated) {
          _setError('验证码错误');
        } else if (e.code == StatusCode.alreadyExists) {
          _setError('该手机号已注册，请返回登录');
        } else {
          _setError('注册失败，请稍后重试');
        }
      }
```

- [ ] **Step 4: Update _resetPassword() error handling (line 247-251)**

Replace:

```dart
    } on GrpcError catch (e) {
      if (mounted) {
        setState(() => _loading = false);
        _setError(e.message ?? '重置密码失败，请稍后重试');
      }
```

With:

```dart
    } on GrpcError catch (e) {
      if (mounted) {
        setState(() => _loading = false);
        if (e.code == StatusCode.resourceExhausted) {
          _setError(grpcErrorMessage(e));
        } else {
          _setError(e.message ?? '重置密码失败，请稍后重试');
        }
      }
```

- [ ] **Step 5: Verify Flutter analysis**

```bash
cd client && flutter analyze lib/features/login/login_page.dart
```

Expected: no issues.

---

### Task 9: Client — Update providers for rate limit errors

**Files:**
- Modify: `client/lib/features/now/providers/now_page_provider.dart`
- Modify: `client/lib/features/past/providers/past_page_provider.dart`
- Modify: `client/lib/features/starmap/providers/starmap_provider.dart`
- Modify: `client/lib/features/setting/setting_page.dart`
- Modify: `client/lib/features/past/trace_detail_page.dart`
- Modify: `client/lib/features/starmap/widgets/chat_sheet.dart`
- Modify: `client/lib/features/setting/feedback_page.dart`
- Modify: `client/lib/features/starmap/constellation_detail_page.dart`

- [ ] **Step 1: Add import to each file**

For each file listed above, add:

```dart
import '../../../core/providers/grpc_error_mapper.dart';
```

Adjust the relative path based on the file's depth:
- `features/<name>/` → `../../core/providers/grpc_error_mapper.dart`
- `features/<name>/providers/` → `../../../core/providers/grpc_error_mapper.dart`
- `features/<name>/widgets/` → `../../../core/providers/grpc_error_mapper.dart`

- [ ] **Step 2: Add GrpcError catch before generic catch in each provider**

**Pattern:** In each file, find `catch (e)` blocks that wrap gRPC calls and add a `on GrpcError catch (e)` block before them:

For provider `.copyWith` pattern (now_page_provider.dart, past_page_provider.dart, starmap_provider.dart):

```dart
    } on GrpcError catch (e) {
      state = state.copyWith(isLoading: false, error: grpcErrorMessage(e));
    } catch (e) {
      state = state.copyWith(isLoading: false, error: e.toString());
    }
```

For page `setState` pattern (setting_page.dart, trace_detail_page.dart, etc.):

```dart
    } on GrpcError catch (e) {
      if (mounted) {
        setState(() { _loading = false; _error = grpcErrorMessage(e); });
      }
    } catch (_) {
      if (mounted) {
        setState(() { _loading = false; _error = '操作失败，请稍后重试'; });
      }
    }
```

- [ ] **Step 3: Full Flutter static analysis**

```bash
cd client && flutter analyze
```

Expected: no issues (zero).

---

### Task 10: Update smoke.sh for rate limit verification

**Files:**
- Modify: `smoke.sh`

- [ ] **Step 1: Add rate limit test to smoke.sh**

Add a new `## Rate Limit` section to `smoke.sh` before the final "all tests passed" line. The test should:

1. Send requests rapidly (more than the pre-auth burst of 30) to a pre-auth RPC
2. Verify that one of the requests returns RESOURCE_EXHAUSTED

Use grpcurl to send rapid Login requests and check for the rate limit response:

```bash
echo ""
echo "=== Rate Limit ==="

# Test pre-auth (Login) rate limiting.
# Burst is 30, so 31 rapid requests should trigger RESOURCE_EXHAUSTED.
RATE_LIMITED=false
for i in $(seq 1 35); do
  STATUS=$(grpcurl -plaintext \
    -d '{"phone":"13800000001","password":"test123456"}' \
    localhost:${GRPC_PORT:-9444} ego.Ego/Login 2>&1)
  if echo "$STATUS" | grep -q "RESOURCE_EXHAUSTED"; then
    echo "  ✓ Rate limit triggered at request $i"
    RATE_LIMITED=true
    break
  fi
done

if [ "$RATE_LIMITED" = true ]; then
  echo "  PASS Rate Limit"
else
  echo "  FAIL: Rate limit not triggered after 35 requests"
  exit 1
fi
```

- [ ] **Step 2: Verify smoke.sh syntax**

```bash
bash -n smoke.sh
```

Expected: no syntax errors.

---

### Task 11: Final verification — all checks

**Files:**
- Verify: all changes compile and pass tests

- [ ] **Step 1: Go unit tests**

```bash
go test ./internal/platform/ratelimit/ -v -count=1
```

Expected: all tests pass.

- [ ] **Step 2: Go vet**

```bash
go vet ./internal/platform/ratelimit/... ./internal/config/... ./internal/bootstrap/...
```

Expected: no issues.

- [ ] **Step 3: Full server build**

```bash
cd server && go build ./...
```

Expected: no errors.

- [ ] **Step 4: Flutter analyze**

```bash
cd client && flutter analyze
```

Expected: zero issues.

- [ ] **Step 5: Smoke test**

```bash
bash smoke.sh
```

Expected: all tests pass, including new rate limit test.

---

## 手动测试清单（真机）

| # | 场景 | 步骤 | 预期 |
|---|------|------|------|
| 1 | 正常使用不受影响 | 正常登录、浏览页面、操作 | 所有功能正常，无异常提示 |
| 2 | 快速连续发送验证码 | 登录页快速连续点击"发送验证码" | 触发限流后显示"请求过于频繁，请稍后再试" |
| 3 | 快速连续登录 | 登录页快速连续点击"登录" | 触发限流后显示"请求过于频繁，请稍后再试" |
| 4 | 快速连续注册 | 注册页快速连续点击"注册" | 触发限流后显示"请求过于频繁，请稍后再试" |
| 5 | 限流恢复 | 等待几秒后重试被限流的操作 | 操作正常成功 |
