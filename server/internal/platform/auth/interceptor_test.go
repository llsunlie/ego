package auth

import (
	"bytes"
	"context"
	"strings"
	"testing"
	"time"

	"ego-server/internal/platform/logging"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestUnaryServerInterceptor_LoginWhitelist(t *testing.T) {
	var buf bytes.Buffer
	base, err := logging.New(logging.Config{Level: "debug", Format: "text", Output: &buf})
	if err != nil {
		t.Fatalf("logging.New: %v", err)
	}

	interceptor := UnaryServerInterceptor([]byte("secret"), base)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		logger := logging.FromContext(ctx)
		logger.Info("handler called")
		return "ok", nil
	}

	info := &grpc.UnaryServerInfo{FullMethod: "/ego.Ego/Login"}

	resp, err := interceptor(context.Background(), nil, info, handler)
	if err != nil {
		t.Fatalf("Login should skip auth: %v", err)
	}
	if resp != "ok" {
		t.Fatalf("expected ok, got %v", resp)
	}

	out := buf.String()
	if !strings.Contains(out, "handler called") {
		t.Fatalf("expected handler log, got: %s", out)
	}
	if !strings.Contains(out, "request_id") {
		t.Fatalf("expected request_id in Login log, got: %s", out)
	}
}

func TestUnaryServerInterceptor_LoggerInContext(t *testing.T) {
	var buf bytes.Buffer
	base, err := logging.New(logging.Config{Level: "debug", Format: "text", Output: &buf})
	if err != nil {
		t.Fatalf("logging.New: %v", err)
	}

	secret := []byte("test-secret")
	interceptor := UnaryServerInterceptor(secret, base)

	// Generate a valid token
	token, err := GenerateJWT("user-1", secret, time.Hour)
	if err != nil {
		t.Fatalf("GenerateJWT: %v", err)
	}

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		logger := logging.FromContext(ctx)
		logger.Info("authenticated request")

		userID, _ := ctx.Value("user_id").(string)
		if userID != "user-1" {
			t.Fatalf("expected user-1 in ctx, got %s", userID)
		}
		return "ok", nil
	}

	md := metadata.Pairs("authorization", "Bearer "+token)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	info := &grpc.UnaryServerInfo{FullMethod: "/ego.Ego/CreateMoment"}

	resp, err := interceptor(ctx, nil, info, handler)
	if err != nil {
		t.Fatalf("interceptor: %v", err)
	}
	if resp != "ok" {
		t.Fatalf("expected ok, got %v", resp)
	}

	out := buf.String()
	if !strings.Contains(out, "authenticated request") {
		t.Fatalf("expected authenticated request log, got: %s", out)
	}
	if !strings.Contains(out, "user_id") {
		t.Fatalf("expected user_id in log, got: %s", out)
	}
	if !strings.Contains(out, "method") {
		t.Fatalf("expected method in log, got: %s", out)
	}
}

func TestUnaryServerInterceptor_MissingAuth(t *testing.T) {
	base := logging.NewNop()
	interceptor := UnaryServerInterceptor([]byte("secret"), base)

	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "should not reach", nil
	}

	ctx := metadata.NewIncomingContext(context.Background(), metadata.Pairs())
	info := &grpc.UnaryServerInfo{FullMethod: "/ego.Ego/CreateMoment"}

	_, err := interceptor(ctx, nil, info, handler)
	if err == nil {
		t.Fatal("expected unauthenticated error")
	}
}
