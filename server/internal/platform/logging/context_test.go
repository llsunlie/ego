package logging

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestWithLogger_FromContext(t *testing.T) {
	var buf bytes.Buffer
	logger, err := New(Config{Level: "debug", Format: "text", Output: &buf})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	ctx := WithLogger(context.Background(), logger)

	retrieved := FromContext(ctx)
	if retrieved != logger {
		t.Fatal("FromContext should return the stored logger")
	}

	retrieved.Info("from context")
	if !strings.Contains(buf.String(), "from context") {
		t.Fatalf("expected output from retrieved logger, got: %s", buf.String())
	}
}

func TestFromContext_EmptyContext(t *testing.T) {
	logger := FromContext(context.Background())
	if logger == nil {
		t.Fatal("FromContext should fall back to default logger, not nil")
	}
	logger.Info("should not panic")
}

func TestFromContext_WrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), loggerKey{}, "not a logger")
	logger := FromContext(ctx)
	if logger == nil {
		t.Fatal("FromContext should fall back to default logger")
	}
	logger.Info("should not panic")
}

func TestWithLogger_ChildDerivation(t *testing.T) {
	var buf bytes.Buffer
	base, err := New(Config{Level: "debug", Format: "text", Output: &buf})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	child := base.With("request_id", "abc-123", "user_id", "user-1")
	ctx := WithLogger(context.Background(), child)

	retrieved := FromContext(ctx)
	retrieved.Info("child log")

	out := buf.String()
	if !strings.Contains(out, "request_id") {
		t.Fatalf("expected request_id in output, got: %s", out)
	}
	if !strings.Contains(out, "user_id") {
		t.Fatalf("expected user_id in output, got: %s", out)
	}
	if !strings.Contains(out, "abc-123") {
		t.Fatalf("expected abc-123 in output, got: %s", out)
	}
}

func TestWithLogger_Overwrite(t *testing.T) {
	var buf1, buf2 bytes.Buffer
	logger1, _ := New(Config{Level: "debug", Format: "text", Output: &buf1})
	logger2, _ := New(Config{Level: "debug", Format: "text", Output: &buf2})

	ctx := WithLogger(context.Background(), logger1)
	ctx = WithLogger(ctx, logger2)

	retrieved := FromContext(ctx)
	if retrieved != logger2 {
		t.Fatal("second WithLogger should overwrite")
	}

	retrieved.Info("from logger2")
	if !strings.Contains(buf2.String(), "from logger2") {
		t.Fatal("expected logger2 output")
	}
	if buf1.Len() > 0 {
		t.Fatal("logger1 should have no output after overwrite")
	}
}

func TestWithLogger_NilLogger(t *testing.T) {
	ctx := WithLogger(context.Background(), nil)
	retrieved := FromContext(ctx)
	// nil stored as interface{} — type assertion to *slog.Logger fails, falls back.
	if retrieved == nil {
		t.Fatal("FromContext should fall back to default logger")
	}
	retrieved.Info("should not panic")
}

func TestFromContext_FallbackIsSlogDefault(t *testing.T) {
	// slog.Default() should be returned when no logger in context.
	ctx := context.Background()
	got := FromContext(ctx)
	if got != slog.Default() {
		t.Fatal("FromContext should return slog.Default() as fallback")
	}
}

func TestWithLogger_ConcurrentRead(t *testing.T) {
	var buf bytes.Buffer
	logger, _ := New(Config{Level: "debug", Format: "text", Output: &buf})
	ctx := WithLogger(context.Background(), logger)

	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func() {
			l := FromContext(ctx)
			l.Info("concurrent")
			done <- true
		}()
	}
	for i := 0; i < 10; i++ {
		<-done
	}
	if !strings.Contains(buf.String(), "concurrent") {
		t.Fatal("expected concurrent reads to succeed")
	}
}
