package login

import (
	"context"
	"io"
	"testing"
	"time"

	"ego-server/internal/auth"
	"ego-server/internal/db/sqlc"

	pb "ego-server/proto/ego"

	"golang.org/x/crypto/bcrypt"
)

// mockDB is an in-memory UserQuerier simulating the users table.
type mockDB struct {
	users map[string]sqlc.GetUserByAccountRow
}

func newMockDB() *mockDB {
	return &mockDB{users: make(map[string]sqlc.GetUserByAccountRow)}
}

func (m *mockDB) GetUserByAccount(_ context.Context, account string) (sqlc.GetUserByAccountRow, error) {
	u, ok := m.users[account]
	if !ok {
		return sqlc.GetUserByAccountRow{}, io.EOF // pgx returns io.EOF for no rows
	}
	return u, nil
}

func (m *mockDB) CreateUser(_ context.Context, arg sqlc.CreateUserParams) error {
	m.users[arg.Account] = sqlc.GetUserByAccountRow{
		ID:           arg.ID,
		PasswordHash: arg.PasswordHash,
	}
	return nil
}

func TestLogin_AutoRegister(t *testing.T) {
	db := newMockDB()
	h := NewHandler(db, []byte("secret"), 24*time.Hour)

	res, err := h.Login(context.Background(), loginReq("alice", "pass1"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Created {
		t.Fatal("expected created=true for new user")
	}
	if res.Token == "" {
		t.Fatal("expected token")
	}
	if _, exists := db.users["alice"]; !exists {
		t.Fatal("user not created in db")
	}
}

func TestLogin_ExistingUserCorrectPassword(t *testing.T) {
	db := newMockDB()
	h := NewHandler(db, []byte("secret"), 24*time.Hour)

	// auto-register
	_, err := h.Login(context.Background(), loginReq("bob", "pass1"))
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	// login again
	res, err := h.Login(context.Background(), loginReq("bob", "pass1"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Created {
		t.Fatal("expected created=false for existing user")
	}
	if res.Token == "" {
		t.Fatal("expected token")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	db := newMockDB()
	h := NewHandler(db, []byte("secret"), 24*time.Hour)

	// auto-register
	_, err := h.Login(context.Background(), loginReq("carol", "correct"))
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	// wrong password
	_, err = h.Login(context.Background(), loginReq("carol", "wrong"))
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
}

func TestRegister_PasswordIsHashed(t *testing.T) {
	db := newMockDB()
	h := NewHandler(db, []byte("secret"), 24*time.Hour)

	_, err := h.Login(context.Background(), loginReq("dave", "secret123"))
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	stored := db.users["dave"]
	if err := bcrypt.CompareHashAndPassword([]byte(stored.PasswordHash), []byte("secret123")); err != nil {
		t.Fatal("stored password is not a bcrypt hash of the original")
	}
}

func TestLogin_TokenContainsUserID(t *testing.T) {
	db := newMockDB()
	h := NewHandler(db, []byte("secret"), 24*time.Hour)

	res, err := h.Login(context.Background(), loginReq("eve", "pass"))
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

// helpers

func loginReq(account, password string) *pb.LoginReq {
	return &pb.LoginReq{Account: account, Password: password}
}
