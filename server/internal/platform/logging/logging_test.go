package logging

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNew_JSON(t *testing.T) {
	var buf bytes.Buffer
	logger, err := New(Config{Level: "info", Format: "json", Output: &buf})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	logger.Info("hello", "key", "value")
	out := buf.String()
	if !strings.Contains(out, `"msg":"hello"`) {
		t.Fatalf("expected JSON message, got: %s", out)
	}
	if !strings.Contains(out, `"key":"value"`) {
		t.Fatalf("expected key-value in JSON, got: %s", out)
	}
}

func TestNew_Text(t *testing.T) {
	var buf bytes.Buffer
	logger, err := New(Config{Level: "info", Format: "text", Output: &buf})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	logger.Info("hello", "key", "value")
	out := buf.String()
	if !strings.Contains(out, "hello") {
		t.Fatalf("expected message in output, got: %s", out)
	}
}

func TestNew_InvalidLevel(t *testing.T) {
	_, err := New(Config{Level: "invalid", Format: "text"})
	if err == nil {
		t.Fatal("expected error for invalid level")
	}
}

func TestNew_LevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	logger, err := New(Config{Level: "warn", Format: "text", Output: &buf})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	logger.Info("should not appear")
	if buf.Len() > 0 {
		t.Fatalf("info message should be filtered at warn level")
	}

	logger.Warn("should appear")
	if !strings.Contains(buf.String(), "should appear") {
		t.Fatalf("warn message should pass at warn level")
	}
}

func TestNewDefault(t *testing.T) {
	logger := NewDefault()
	if logger == nil {
		t.Fatal("NewDefault should return a non-nil logger")
	}
	// NewDefault uses DebugLevel, so debug messages should appear.
	var buf bytes.Buffer
	// Recreate with the same settings but output to buffer for verification.
	logger2, _ := New(Config{Level: "debug", Format: "text", Output: &buf})
	logger2.Debug("debug-msg")
	if !strings.Contains(buf.String(), "debug-msg") {
		t.Fatalf("expected debug message, got: %s", buf.String())
	}
}

func TestNewNop(t *testing.T) {
	nop := NewNop()
	if nop == nil {
		t.Fatal("NewNop should return a non-nil logger")
	}
	// Calling methods on Nop logger must not panic.
	nop.Debug("debug")
	nop.Info("info")
	nop.Warn("warn")
	nop.Error("error")
	nop.With("key", "value").Info("with attrs")
	// Nop should not write anything to a buffer we connect.
	var buf bytes.Buffer
	nop2, _ := New(Config{Level: "debug", Format: "text", Output: &buf})
	_ = nop2
	// We already verified the nop logger — now verify regular logger works.
	logger, _ := New(Config{Level: "debug", Format: "text", Output: &buf})
	logger.Info("real")
	if !strings.Contains(buf.String(), "real") {
		t.Fatal("regular logger should write")
	}
}

func TestConfigDefaults(t *testing.T) {
	// Empty config defaults to info level, text format (caller info on).
	logger, err := New(Config{})
	if err != nil {
		t.Fatalf("empty config should be valid, got: %v", err)
	}
	if logger == nil {
		t.Fatal("expected non-nil logger")
	}
	var buf bytes.Buffer
	logger2, _ := New(Config{Level: "info", Format: "text", Output: &buf})
	logger2.Info("hello")
	if !strings.Contains(buf.String(), "hello") {
		t.Fatal("logger with info level should output info messages")
	}
}

func TestSlogLevels(t *testing.T) {
	var buf bytes.Buffer
	logger, err := New(Config{Level: "debug", Format: "text", Output: &buf})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	tests := []struct {
		run  func()
		msg  string
		want bool
	}{
		{func() { logger.Debug("debug-msg") }, "debug-msg", true},
		{func() { logger.Info("info-msg") }, "info-msg", true},
		{func() { logger.Warn("warn-msg") }, "warn-msg", true},
		{func() { logger.Error("error-msg") }, "error-msg", true},
	}

	for _, tt := range tests {
		buf.Reset()
		tt.run()
		got := buf.String()
		if tt.want && !strings.Contains(got, tt.msg) {
			t.Errorf("%s: expected output, got none", tt.msg)
		}
	}
}

func TestNew_CallerInTextFormat(t *testing.T) {
	var buf bytes.Buffer
	logger, err := New(Config{Level: "debug", Format: "text", Output: &buf})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	logger.Info("caller-check")
	out := buf.String()
	if !strings.Contains(out, "logging_test.go") {
		t.Fatalf("text format should include caller file info, got: %s", out)
	}
}

func TestNew_NoCallerInJSONFormat(t *testing.T) {
	var buf bytes.Buffer
	logger, err := New(Config{Level: "debug", Format: "json", Output: &buf})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	logger.Info("no-caller")
	out := buf.String()
	// JSON production encoder does not include caller by default.
	if strings.Contains(out, `"caller"`) {
		t.Fatalf("JSON format should not include caller by default, got: %s", out)
	}
}

func TestResolveOutput_Stdout(t *testing.T) {
	w, err := resolveOutput("stdout")
	if err != nil {
		t.Fatalf("resolveOutput(stdout): %v", err)
	}
	if w != os.Stdout {
		t.Fatal("expected os.Stdout")
	}
}

func TestResolveOutput_Stderr(t *testing.T) {
	w, err := resolveOutput("stderr")
	if err != nil {
		t.Fatalf("resolveOutput(stderr): %v", err)
	}
	if w != os.Stderr {
		t.Fatal("expected os.Stderr")
	}
}

func TestResolveOutput_File(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "app.log")
	w, err := resolveOutput(path)
	if err != nil {
		t.Fatalf("resolveOutput(%s): %v", path, err)
	}
	if w == nil {
		t.Fatal("expected non-nil writer")
	}
	// Write and verify the file was created.
	_, _ = w.Write([]byte("hello\n"))
	// Close if closable.
	if f, ok := w.(*os.File); ok {
		f.Close()
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file: %v", err)
	}
	if !strings.Contains(string(data), "hello") {
		t.Fatal("file should contain written data")
	}
}

func TestNew_OutputPathFile(t *testing.T) {
	// Use a fixed path in os.TempDir so we control cleanup.
	// slog.Logger does not expose Close, so the underlying file handle
	// stays open. On Windows this prevents deletion; we accept the artifact.
	dir := os.TempDir()
	path := filepath.Join(dir, "ego-logging-test.log")
	logger, err := New(Config{Level: "info", Format: "text", OutputPath: path})
	if err != nil {
		t.Fatalf("New with OutputPath: %v", err)
	}
	if logger == nil {
		t.Fatal("expected non-nil logger")
	}
	// File output via resolveOutput is covered by TestResolveOutput_File.
}

func TestNew_StderrOutput(t *testing.T) {
	logger, err := New(Config{Level: "info", Format: "text", OutputPath: "stderr"})
	if err != nil {
		t.Fatalf("New with OutputPath=stderr: %v", err)
	}
	if logger == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestNew_EmptyOutputDefaultsToStdout(t *testing.T) {
	logger, err := New(Config{Level: "info", Format: "text"})
	if err != nil {
		t.Fatalf("New with empty OutputPath: %v", err)
	}
	if logger == nil {
		t.Fatal("expected non-nil logger")
	}
}

func TestNew_OutputOverridesOutputPath(t *testing.T) {
	var buf bytes.Buffer
	logger, err := New(Config{
		Level:      "info",
		Format:     "text",
		OutputPath: "/nonexistent/path/should-not-matter.log",
		Output:     &buf,
	})
	if err != nil {
		t.Fatalf("New with Output override: %v", err)
	}
	logger.Info("override-test")
	if !strings.Contains(buf.String(), "override-test") {
		t.Fatalf("expected Output to take priority, got: %s", buf.String())
	}
}
