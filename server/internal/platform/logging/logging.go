package logging

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/exp/zapslog"
	"go.uber.org/zap/zapcore"

	"gopkg.in/natefinch/lumberjack.v2"
)

type Config struct {
	Level      string // debug, info, warn, error
	Format     string // json, text
	OutputPath string // stdout, stderr, or file path
	Output     io.Writer
}

func New(cfg Config) (*slog.Logger, error) {
	zapLevel, err := zapcore.ParseLevel(cfg.Level)
	if err != nil {
		return nil, err
	}

	w := cfg.Output
	if w == nil {
		w, err = resolveOutput(cfg.OutputPath)
		if err != nil {
			return nil, err
		}
	}

	encoder, handlerOpts := buildEncoder(cfg.Format)

	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(w),
		zapLevel,
	)

	return slog.New(zapslog.NewHandler(core, handlerOpts...)), nil
}

func resolveOutput(path string) (io.Writer, error) {
	switch path {
	case "", "stdout":
		return os.Stdout, nil
	case "stderr":
		return os.Stderr, nil
	default:
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return nil, fmt.Errorf("log output: %w", err)
		}
		return &lumberjack.Logger{
			Filename:   path,
			MaxSize:    100, // MB before rotation
			MaxBackups: 30,  // keep at most 30 old files
			MaxAge:     30,  // days to retain old logs
			Compress:   true,
			LocalTime:  true,
		}, nil
	}
}

func NewDefault() *slog.Logger {
	return slog.New(zapslog.NewHandler(
		zapcore.NewCore(
			zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
			zapcore.AddSync(os.Stdout),
			zapcore.DebugLevel,
		),
		zapslog.WithCaller(true),
	))
}

func NewNop() *slog.Logger {
	return slog.New(zapslog.NewHandler(zapcore.NewNopCore()))
}

func buildEncoder(format string) (zapcore.Encoder, []zapslog.HandlerOption) {
	switch format {
	case "json":
		return zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig()), nil
	default:
		return zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()), []zapslog.HandlerOption{zapslog.WithCaller(true)}
	}
}
