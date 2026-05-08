package postgres

import (
	"context"
	"testing"

	"ego-server/internal/platform/postgres/sqlc"

	"github.com/jackc/pgx/v5/pgxpool"
)

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	pool, err := pgxpool.New(context.Background(), "postgres://ego:ego@localhost:5432/ego?sslmode=disable")
	if err != nil {
		t.Fatalf("pgxpool.New: %v", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		t.Fatalf("ping: %v", err)
	}
	t.Cleanup(func() { pool.Close() })
	return pool
}

func testQueries(t *testing.T) *sqlc.Queries {
	return sqlc.New(testPool(t))
}
