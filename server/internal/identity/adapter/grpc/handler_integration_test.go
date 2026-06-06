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
	ids := mockIDGen{}
	sms := mockSmsService{}

	loginUseCase := identityapp.NewLoginUseCase(userRepo, hasher, tokens)
	registerUseCase := identityapp.NewRegisterUseCase(userRepo, hasher, tokens, ids, sms)
	sendCodeUseCase := identityapp.NewSendCodeUseCase(sms)
	checkPhoneUseCase := identityapp.NewCheckPhoneUseCase(userRepo)

	return NewHandler(loginUseCase, registerUseCase, sendCodeUseCase, checkPhoneUseCase), pool
}

func cleanupUser(t *testing.T, pool *pgxpool.Pool, phone string) {
	t.Helper()
	pool.Exec(context.Background(), "DELETE FROM users WHERE phone = $1", phone)
}

func TestIntegration_Register(t *testing.T) {
	h, pool := newTestHandlerRealDB(t)
	phone := "13800010001"
	t.Cleanup(func() { cleanupUser(t, pool, phone) })

	res, err := h.Register(context.Background(), &pb.RegisterReq{Phone: phone, Code: "123456", Password: "pass123"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Token == "" {
		t.Fatal("expected token")
	}
}

func TestIntegration_LoginAfterRegister(t *testing.T) {
	h, pool := newTestHandlerRealDB(t)
	phone := "13800010002"
	t.Cleanup(func() { cleanupUser(t, pool, phone) })

	_, err := h.Register(context.Background(), &pb.RegisterReq{Phone: phone, Code: "123456", Password: "pass123"})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	res, err := h.Login(context.Background(), &pb.LoginReq{Phone: phone, Password: "pass123"})
	if err != nil {
		t.Fatalf("login: %v", err)
	}
	if res.Token == "" {
		t.Fatal("expected token")
	}
}

func TestIntegration_WrongPassword(t *testing.T) {
	h, pool := newTestHandlerRealDB(t)
	phone := "13800010003"
	t.Cleanup(func() { cleanupUser(t, pool, phone) })

	_, err := h.Register(context.Background(), &pb.RegisterReq{Phone: phone, Code: "123456", Password: "correct"})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	_, err = h.Login(context.Background(), &pb.LoginReq{Phone: phone, Password: "wrong"})
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
}

func TestIntegration_TokenContainsUserID(t *testing.T) {
	h, pool := newTestHandlerRealDB(t)
	phone := "13800010004"
	t.Cleanup(func() { cleanupUser(t, pool, phone) })

	_, err := h.Register(context.Background(), &pb.RegisterReq{Phone: phone, Code: "123456", Password: "pass123"})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	res, err := h.Login(context.Background(), &pb.LoginReq{Phone: phone, Password: "pass123"})
	if err != nil {
		t.Fatalf("login: %v", err)
	}

	userID, err := auth.ParseJWT(res.Token, []byte("secret"))
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	if userID == "" {
		t.Fatal("token should contain user_id")
	}
}

func TestIntegration_DuplicatePhone(t *testing.T) {
	h, pool := newTestHandlerRealDB(t)
	phone := "13800010005"
	t.Cleanup(func() { cleanupUser(t, pool, phone) })

	_, err := h.Register(context.Background(), &pb.RegisterReq{Phone: phone, Code: "123456", Password: "pass123"})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	_, err = h.Register(context.Background(), &pb.RegisterReq{Phone: phone, Code: "123456", Password: "pass456"})
	if err == nil {
		t.Fatal("expected error for duplicate phone")
	}
}

func TestIntegration_CheckPhone_NewPhone(t *testing.T) {
	h, _ := newTestHandlerRealDB(t)
	phone := "13800010006"

	res, err := h.CheckPhone(context.Background(), &pb.CheckPhoneReq{Phone: phone})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Registered {
		t.Fatal("expected registered=false for new phone")
	}
}

func TestIntegration_CheckPhone_RegisteredPhone(t *testing.T) {
	h, pool := newTestHandlerRealDB(t)
	phone := "13800010007"
	t.Cleanup(func() { cleanupUser(t, pool, phone) })

	_, err := h.Register(context.Background(), &pb.RegisterReq{Phone: phone, Code: "123456", Password: "pass123"})
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	res, err := h.CheckPhone(context.Background(), &pb.CheckPhoneReq{Phone: phone})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Registered {
		t.Fatal("expected registered=true for existing phone")
	}
}
