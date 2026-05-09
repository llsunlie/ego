package auth

import (
	"context"
	"log/slog"
	"strings"

	"ego-server/internal/platform/logging"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func UnaryServerInterceptor(jwtSecret []byte, baseLogger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if strings.Contains(info.FullMethod, "Login") {
			logger := baseLogger.With("request_id", uuid.NewString(), "method", info.FullMethod)
			ctx = logging.WithLogger(ctx, logger)
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		values := md["authorization"]
		if len(values) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing authorization")
		}

		tokenStr := strings.TrimPrefix(values[0], "Bearer ")
		if tokenStr == values[0] {
			return nil, status.Error(codes.Unauthenticated, "invalid authorization format")
		}

		userID, err := ParseJWT(tokenStr, jwtSecret)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		ctx = context.WithValue(ctx, "user_id", userID)

		logger := baseLogger.With(
			"request_id", uuid.NewString(),
			"user_id", userID,
			"method", info.FullMethod,
		)
		ctx = logging.WithLogger(ctx, logger)

		return handler(ctx, req)
	}
}
