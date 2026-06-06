package identitygrpc

import (
	"context"
	"testing"
	"time"

	"ego-server/internal/identity/app"
	"ego-server/internal/identity/domain"
	"ego-server/internal/platform/auth"

	"github.com/google/uuid"

	pb "ego-server/proto/ego"

	"golang.org/x/crypto/bcrypt"
)

type mockUserRepo struct {
	users map[string]*domain.User
}

func newMockUserRepo() *mockUserRepo {
	return &mockUserRepo{users: make(map[string]*domain.User)}
}

func (m *mockUserRepo) FindByPhone(_ context.Context, phone string) (*domain.User, error) {
	u, ok := m.users[phone]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}

func (m *mockUserRepo) Create(_ context.Context, user *domain.User) error {
	m.users[user.Phone] = user
	return nil
}

type mockIDGen struct{}

func (mockIDGen) New() string { return uuid.New().String() }

// mockSmsService always succeeds for Send and Verify — used in unit tests.
type mockSmsService struct{}

func (mockSmsService) Send(_ context.Context, _ string) error               { return nil }
func (mockSmsService) Verify(_ context.Context, _, _ string) (bool, error)  { return true, nil }

func newTestHandler(repo *mockUserRepo) *Handler {
	hasher := auth.BcryptHasher{}
	tokens := auth.JWTIssuer{Secret: []byte("secret"), Exp: 24 * time.Hour}
	ids := mockIDGen{}
	sms := mockSmsService{}

	login := app.NewLoginUseCase(repo, hasher, tokens)
	register := app.NewRegisterUseCase(repo, hasher, tokens, ids, sms)
	sendCode := app.NewSendCodeUseCase(sms)
	checkPhone := app.NewCheckPhoneUseCase(repo)

	return NewHandler(login, register, sendCode, checkPhone)
}

func TestRegister_Success(t *testing.T) {
	repo := newMockUserRepo()
	h := newTestHandler(repo)

	res, err := h.Register(context.Background(), registerReq("13800000001", "123456", "pass123"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Token == "" {
		t.Fatal("expected token")
	}
	if _, exists := repo.users["13800000001"]; !exists {
		t.Fatal("user not created in db")
	}
}

func TestRegister_DuplicatePhone(t *testing.T) {
	repo := newMockUserRepo()
	h := newTestHandler(repo)

	_, err := h.Register(context.Background(), registerReq("13800000002", "123456", "pass123"))
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	_, err = h.Register(context.Background(), registerReq("13800000002", "123456", "pass123"))
	if err == nil {
		t.Fatal("expected error for duplicate phone")
	}
}

func TestLogin_ExistingUserCorrectPassword(t *testing.T) {
	repo := newMockUserRepo()
	h := newTestHandler(repo)

	_, err := h.Register(context.Background(), registerReq("13800000003", "123456", "pass123"))
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	res, err := h.Login(context.Background(), loginReq("13800000003", "pass123"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Token == "" {
		t.Fatal("expected token")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	repo := newMockUserRepo()
	h := newTestHandler(repo)

	_, err := h.Register(context.Background(), registerReq("13800000004", "123456", "correct"))
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	_, err = h.Login(context.Background(), loginReq("13800000004", "wrong"))
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	repo := newMockUserRepo()
	h := newTestHandler(repo)

	_, err := h.Login(context.Background(), loginReq("13800000005", "pass1"))
	if err == nil {
		t.Fatal("expected error for non-existent user")
	}
}

func TestRegister_PasswordIsHashed(t *testing.T) {
	repo := newMockUserRepo()
	h := newTestHandler(repo)

	_, err := h.Register(context.Background(), registerReq("13800000006", "123456", "secret123"))
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	stored := repo.users["13800000006"]
	if err := bcrypt.CompareHashAndPassword([]byte(stored.PasswordHash), []byte("secret123")); err != nil {
		t.Fatal("stored password is not a bcrypt hash of the original")
	}
}

func TestLogin_TokenContainsUserID(t *testing.T) {
	repo := newMockUserRepo()
	h := newTestHandler(repo)

	_, err := h.Register(context.Background(), registerReq("13800000007", "123456", "pass123"))
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	res, err := h.Login(context.Background(), loginReq("13800000007", "pass123"))
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

func TestCheckPhone_NewPhone(t *testing.T) {
	repo := newMockUserRepo()
	h := newTestHandler(repo)

	res, err := h.CheckPhone(context.Background(), &pb.CheckPhoneReq{Phone: "13800000008"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Registered {
		t.Fatal("expected registered=false for new phone")
	}
}

func TestCheckPhone_RegisteredPhone(t *testing.T) {
	repo := newMockUserRepo()
	h := newTestHandler(repo)

	_, err := h.Register(context.Background(), registerReq("13800000009", "123456", "pass123"))
	if err != nil {
		t.Fatalf("register: %v", err)
	}

	res, err := h.CheckPhone(context.Background(), &pb.CheckPhoneReq{Phone: "13800000009"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Registered {
		t.Fatal("expected registered=true for existing phone")
	}
}

func TestCheckPhone_InvalidPhone(t *testing.T) {
	repo := newMockUserRepo()
	h := newTestHandler(repo)

	_, err := h.CheckPhone(context.Background(), &pb.CheckPhoneReq{Phone: "12345"})
	if err == nil {
		t.Fatal("expected error for invalid phone")
	}
}

func TestSendVerificationCode_InvalidPhone(t *testing.T) {
	repo := newMockUserRepo()
	h := newTestHandler(repo)

	err := h.sendCode.SendCode(context.Background(), "12345")
	if err == nil {
		t.Fatal("expected error for invalid phone")
	}
}

func loginReq(phone, password string) *pb.LoginReq {
	return &pb.LoginReq{Phone: phone, Password: password}
}

func registerReq(phone, code, password string) *pb.RegisterReq {
	return &pb.RegisterReq{Phone: phone, Code: code, Password: password}
}
