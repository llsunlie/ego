package main

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"ego-server/internal/config"
	"ego-server/internal/platform/postgres"
)

func main() {
	cfg := config.Load()
	pool, err := postgres.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	ctx := context.Background()

	sql, err := os.ReadFile(filepath.Join("internal", "platform", "postgres", "migrations", "001_users.sql"))
	if err != nil {
		log.Fatalf("read migration: %v", err)
	}

	if _, err := pool.Exec(ctx, string(sql)); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	log.Println("migration 001_users applied")
}
