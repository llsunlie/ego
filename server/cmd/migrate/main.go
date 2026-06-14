package main

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"

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

	migrationsDir := filepath.Join("internal", "platform", "postgres", "migrations")
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		logger.Error("read migrations dir failed", "error", err)
		os.Exit(1)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, name := range files {
		sql, err := os.ReadFile(filepath.Join(migrationsDir, name))
		if err != nil {
			logger.Error("read migration failed", "file", name, "error", err)
			os.Exit(1)
		}
		if _, err := pool.Exec(ctx, string(sql)); err != nil {
			logger.Error("migrate failed", "file", name, "error", err)
			os.Exit(1)
		}
		logger.Info("migration applied", "file", name)
	}
}
