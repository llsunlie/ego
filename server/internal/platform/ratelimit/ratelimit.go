package ratelimit

import (
	"context"
	"log/slog"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"ego-server/internal/config"
	"ego-server/internal/platform/auth"

	"golang.org/x/time/rate"
)

// Limiter holds token buckets per dimension and enforces rate limits.
type Limiter struct {
	authRate     rate.Limit
	authBurst    int
	preAuthRate  rate.Limit
	preAuthBurst int
	maxBuckets   int
	buckets      sync.Map     // string → *rateLimiterEntry
	bucketCount  atomic.Int64 // approximate live entry count
	logger       *slog.Logger

	stopCh chan struct{}
	doneCh chan struct{}
}

type rateLimiterEntry struct {
	limiter  *rate.Limiter
	lastUsed time.Time
}

// New creates a Limiter from config, starting a background goroutine to
// clean up stale buckets. Call Close() to stop the cleanup goroutine.
func New(cfg *config.Config, logger *slog.Logger) *Limiter {
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
		logger:       logger,
		stopCh:       make(chan struct{}),
		doneCh:       make(chan struct{}),
	}

	logger.Info("ratelimit started",
		"auth_rate", authRate,
		"auth_burst", authBurst,
		"preauth_rate", preAuthRate,
		"preauth_burst", preAuthBurst,
		"max_buckets", maxBuckets,
	)
	go l.cleanupLoop()
	return l
}

// Close stops the background cleanup goroutine.
// Safe to call when no goroutine is running (stopCh is nil).
func (l *Limiter) Close() {
	if l.stopCh == nil {
		return
	}
	close(l.stopCh)
	<-l.doneCh
}

func (l *Limiter) logDenied(ctx context.Context, fullMethod, ip, userID, dim string) {
	l.logger.WarnContext(ctx, "ratelimit denied",
		"method", fullMethod,
		"ip", ip,
		"user_id", userID,
		"dim", dim,
	)
}

// Allow checks whether the request should be allowed.
// fullMethod is the gRPC method path (e.g. "/ego.Ego/Login").
// userID is "" for pre-auth methods.
// ip is the client IP.
// Returns (allowed, deniedDim). deniedDim is "ip" or "user" when denied.
func (l *Limiter) Allow(ctx context.Context, fullMethod string, userID string, ip string) (bool, string) {
	methodName := extractMethodName(fullMethod)
	isPreAuth := auth.PreAuthMethods[methodName]

	if isPreAuth {
		if !l.checkPreAuth(ip) {
			return false, "ip"
		}
		return true, ""
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
	entryAny, loaded := l.buckets.LoadOrStore(key, &rateLimiterEntry{
		limiter:  rate.NewLimiter(r, burst),
		lastUsed: time.Now(),
	})
	entry := entryAny.(*rateLimiterEntry)
	entry.lastUsed = time.Now()

	if !loaded {
		l.bucketCount.Add(1)

		// Check bucket count: if over limit, trigger immediate cleanup then fail-open.
		if int(l.bucketCount.Load()) > l.maxBuckets {
			l.cleanupOnce()
			if int(l.bucketCount.Load()) > l.maxBuckets {
				// Fail-open: allow the request, don't store a new bucket.
				l.buckets.Delete(key)
				l.bucketCount.Add(-1)
				return true
			}
		}
	}

	return entry.limiter.Allow()
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
	l.buckets.Range(func(key, value any) bool {
		entry := value.(*rateLimiterEntry)
		if entry.lastUsed.Before(cutoff) {
			l.buckets.Delete(key)
			l.bucketCount.Add(-1)
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
