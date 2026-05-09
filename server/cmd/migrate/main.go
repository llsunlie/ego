package main

import (
	"context"
	"os"
	"path/filepath"

	"ego-server/internal/config"
	"ego-server/internal/platform/logging"
	"ego-server/internal/platform/postgres"
)

func main() {
	logger := logging.NewDefault()

	cfg := config.Load()
	pool, err := postgres.Connect(cfg.DatabaseURL)
	if err != nil {
		logger.Error("db connect failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	ctx := context.Background()

	sql, err := os.ReadFile(filepath.Join("internal", "platform", "postgres", "migrations", "001_users.sql"))
	if err != nil {
		logger.Error("read migration failed", "error", err)
		os.Exit(1)
	}

	if _, err := pool.Exec(ctx, string(sql)); err != nil {
		logger.Error("migrate failed", "error", err)
		os.Exit(1)
	}
	logger.Info("migration applied", "file", "001_users.sql")
}
