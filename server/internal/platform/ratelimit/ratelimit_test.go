package ratelimit

import (
	"context"
	"testing"
	"time"

	"ego-server/internal/platform/logging"
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
		authRate:        10,
		authBurst:       20,
		preAuthRate:     10,
		preAuthBurst:    30,
		maxBuckets:      500,
		cleanupInterval: time.Minute,
		bucketTTL:       5 * time.Minute,
		logger:          logging.NewNop(),
	}
}
