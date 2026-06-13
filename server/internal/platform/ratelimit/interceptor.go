package ratelimit

import (
	"context"
	"net"
	"strings"

	"ego-server/internal/platform/logging"

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
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		ip := extractClientIP(ctx)
		userID, _ := ctx.Value("user_id").(string)

		// Inject client IP into the request logger so all downstream
		// logs (composite handler, domain handlers) include the IP.
		if logger := logging.FromContext(ctx); logger != nil {
			ctx = logging.WithLogger(ctx, logger.With("ip", ip))
		}

		allowed, dim := l.Allow(ctx, info.FullMethod, userID, ip)
		if !allowed {
			l.logDenied(ctx, info.FullMethod, ip, userID, dim)
			return nil, status.Errorf(codes.ResourceExhausted, "请求过于频繁，请稍后再试")
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
