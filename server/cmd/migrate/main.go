package main

import (
	"context"
	"log"
	"os"
	"path/filepath"

	"ego-server/internal/config"
	"ego-server/internal/db"
)

func main() {
	cfg := config.Load()
	pool, err := db.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	ctx := context.Background()

	sql, err := os.ReadFile(filepath.Join("internal", "db", "migrations", "001_users.sql"))
	if err != nil {
		log.Fatalf("read migration: %v", err)
	}

	if _, err := pool.Exec(ctx, string(sql)); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	log.Println("migration 001_users applied")
}
