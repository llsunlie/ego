package identitygrpc

import (
	"context"
	"testing"
	"time"

	"ego-server/internal/identity/app"
	"ego-server/internal/identity/domain"
	"ego-server/internal/platform/auth"

	pb "ego-server/proto/ego"

	"golang.org/x/crypto/bcrypt"
)

type mockUserRepo struct {
	users map[string]*domain.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[string]*domain.User)}
}

func (m *mockUserRepo) FindByAccount(_ context.Context, account string) (*domain.User, error) {
	u, ok := m.users[account]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

func (m *mockUserRepo) Create(_ context.Context, user *domain.User) error {
	m.users[user.Account] = user
	return nil
}

func newTestHandler(repo *mockUserRepo) *Handler {
	hasher := auth.BcryptHasher{}
	tokens := auth.JWTIssuer{Secret: []byte("secret"), Exp: 24 * time.Hour}
	login := app.NewLoginUseCase(repo, hasher, tokens)
	return NewHandler(login)
}

func TestLogin_AutoRegister(t *testing.T) {
	repo := newMockUserRepo()
	h := newTestHandler(repo)

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
	if _, exists := repo.users["alice"]; !exists {
		t.Fatal("user not created in db")
	}
}

func TestLogin_ExistingUserCorrectPassword(t *testing.T) {
	repo := newMockUserRepo()
	h := newTestHandler(repo)

	_, err := h.Login(context.Background(), loginReq("bob", "pass1"))
	if err != nil {
		t.Fatalf("register: %v", err)
	}

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
	repo := newMockUserRepo()
	h := newTestHandler(repo)

	_, err := h.Login(context.Background(), loginReq("carol", "correct"))
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	_, err = h.Login(context.Background(), loginReq("carol", "wrong"))
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
}

func TestRegister_PasswordIsHashed(t *testing.T) {
	repo := newMockUserRepo()
	h := newTestHandler(repo)

	_, err := h.Login(context.Background(), loginReq("dave", "secret123"))
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	stored := repo.users["dave"]
	if err := bcrypt.CompareHashAndPassword([]byte(stored.PasswordHash), []byte("secret123")); err != nil {
		t.Fatal("stored password is not a bcrypt hash of the original")
	}
}

func TestLogin_TokenContainsUserID(t *testing.T) {
	repo := newMockUserRepo()
	h := newTestHandler(repo)

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

func loginReq(account, password string) *pb.LoginReq {
	return &pb.LoginReq{Account: account, Password: password}
}
