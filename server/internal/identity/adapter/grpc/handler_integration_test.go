package identitygrpc

import (
	"context"
	"testing"
	"time"

	identityapp "ego-server/internal/identity/app"
	identitypostgres "ego-server/internal/identity/adapter/postgres"
	"ego-server/internal/platform/auth"
	"ego-server/internal/platform/postgres/sqlc"

	"github.com/jackc/pgx/v5/pgxpool"

	pb "ego-server/proto/ego"
)

const testDSN = "postgres://ego:ego@localhost:5432/ego?sslmode=disable"

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	pool, err := pgxpool.New(context.Background(), testDSN)
	if err != nil {
		t.Fatalf("pgxpool.New: %v", err)
	}
	if err := pool.Ping(context.Background()); err != nil {
		t.Fatalf("ping: %v", err)
	}
	t.Cleanup(func() { pool.Close() })
	return pool
}

func newTestHandlerRealDB(t *testing.T) (*Handler, *pgxpool.Pool) {
	t.Helper()
	pool := testPool(t)
	queries := sqlc.New(pool)
	userRepo := identitypostgres.NewUserRepository(queries)
	hasher := auth.BcryptHasher{}
	tokens := auth.JWTIssuer{Secret: []byte("secret"), Exp: 24 * time.Hour}
	loginUseCase := identityapp.NewLoginUseCase(userRepo, hasher, tokens)
	return NewHandler(loginUseCase), pool
}

func cleanupUser(t *testing.T, pool *pgxpool.Pool, account string) {
	t.Helper()
	pool.Exec(context.Background(), "DELETE FROM users WHERE account = $1", account)
}

func TestIntegration_AutoRegister(t *testing.T) {
	h, pool := newTestHandlerRealDB(t)
	account := "intg-auto@test.com"
	t.Cleanup(func() { cleanupUser(t, pool, account) })

	res, err := h.Login(context.Background(), &pb.LoginReq{Account: account, Password: "pass1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Created {
		t.Fatal("expected created=true for new user")
	}
	if res.Token == "" {
		t.Fatal("expected token")
	}
}

func TestIntegration_ExistingUserLogin(t *testing.T) {
	h, pool := newTestHandlerRealDB(t)
	account := "intg-existing@test.com"
	t.Cleanup(func() { cleanupUser(t, pool, account) })

	_, err := h.Login(context.Background(), &pb.LoginReq{Account: account, Password: "pass1"})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	res, err := h.Login(context.Background(), &pb.LoginReq{Account: account, Password: "pass1"})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if res.Created {
		t.Fatal("expected created=false for existing user")
	}
	if res.Token == "" {
		t.Fatal("expected token")
	}
}

func TestIntegration_WrongPassword(t *testing.T) {
	h, pool := newTestHandlerRealDB(t)
	account := "intg-wrong@test.com"
	t.Cleanup(func() { cleanupUser(t, pool, account) })

	_, err := h.Login(context.Background(), &pb.LoginReq{Account: account, Password: "correct"})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	_, err = h.Login(context.Background(), &pb.LoginReq{Account: account, Password: "wrong"})
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
}

func TestIntegration_TokenContainsUserID(t *testing.T) {
	h, pool := newTestHandlerRealDB(t)
	account := "intg-token@test.com"
	t.Cleanup(func() { cleanupUser(t, pool, account) })

	res, err := h.Login(context.Background(), &pb.LoginReq{Account: account, Password: "pass"})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	userID, err := auth.ParseJWT(res.Token, []byte("secret"))
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	if userID == "" {
		t.Fatal("token should contain user_id")
	}
}
