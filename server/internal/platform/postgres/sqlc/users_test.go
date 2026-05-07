package sqlc

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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

func pgUUID(id string) pgtype.UUID {
	var u pgtype.UUID
	u.Scan(id)
	return u
}

func TestCreateUser(t *testing.T) {
	pool := testPool(t)
	q := New(pool)

	id := uuid.New()
	now := time.Now().UTC()

	err := q.CreateUser(context.Background(), CreateUserParams{
		ID:           pgUUID(id.String()),
		Account:      "test-create@example.com",
		PasswordHash: "$2a$10$placeholder",
		CreatedAt:    pgtype.Timestamptz{Time: now, Valid: true},
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	// cleanup
	pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", pgUUID(id.String()))
}

func TestCreateUser_DuplicateAccount(t *testing.T) {
	pool := testPool(t)
	q := New(pool)

	id1 := uuid.New()
	id2 := uuid.New()
	now := time.Now().UTC()

	err := q.CreateUser(context.Background(), CreateUserParams{
		ID:           pgUUID(id1.String()),
		Account:      "dup@example.com",
		PasswordHash: "$2a$10$placeholder",
		CreatedAt:    pgtype.Timestamptz{Time: now, Valid: true},
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	err = q.CreateUser(context.Background(), CreateUserParams{
		ID:           pgUUID(id2.String()),
		Account:      "dup@example.com",
		PasswordHash: "$2a$10$placeholder",
		CreatedAt:    pgtype.Timestamptz{Time: now, Valid: true},
	})
	if err == nil {
		t.Fatal("expected error on duplicate account")
	}

	// cleanup
	pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", pgUUID(id1.String()))
	pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", pgUUID(id2.String()))
}

func TestGetUserByAccount(t *testing.T) {
	pool := testPool(t)
	q := New(pool)

	id := uuid.New()
	now := time.Now().UTC()

	err := q.CreateUser(context.Background(), CreateUserParams{
		ID:           pgUUID(id.String()),
		Account:      "test-get@example.com",
		PasswordHash: "$2a$10$test-hash",
		CreatedAt:    pgtype.Timestamptz{Time: now, Valid: true},
	})
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	row, err := q.GetUserByAccount(context.Background(), "test-get@example.com")
	if err != nil {
		t.Fatalf("GetUserByAccount: %v", err)
	}

	gotUUID, err := uuid.FromBytes(row.ID.Bytes[:])
	if err != nil {
		t.Fatalf("uuid.FromBytes: %v", err)
	}
	if gotUUID != id {
		t.Fatalf("expected ID %s, got %s", id.String(), gotUUID.String())
	}
	if row.PasswordHash != "$2a$10$test-hash" {
		t.Fatalf("expected PasswordHash %s, got %s", "$2a$10$test-hash", row.PasswordHash)
	}

	// cleanup
	pool.Exec(context.Background(), "DELETE FROM users WHERE id = $1", pgUUID(id.String()))
}

func TestGetUserByAccount_NotFound(t *testing.T) {
	pool := testPool(t)
	q := New(pool)

	_, err := q.GetUserByAccount(context.Background(), "nonexistent@example.com")
	if err == nil {
		t.Fatal("expected error for nonexistent user")
	}
}
