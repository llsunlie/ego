package grpc_test

import (
	"context"
	"testing"
	"time"

	settinggrpc "ego-server/internal/setting/adapter/grpc"
	settingapp "ego-server/internal/setting/app"
	settingdomain "ego-server/internal/setting/domain"

	pb "ego-server/proto/ego"
)

// stubUserReader implements domain.UserReader for testing.
type stubUserReader struct {
	user *settingdomain.UserInfo
	err  error
}

func (s *stubUserReader) FindByID(_ context.Context, _ string) (*settingdomain.UserInfo, error) {
	return s.user, s.err
}

func TestGetProfile_WithValidUserID(t *testing.T) {
	reader := &stubUserReader{
		user: &settingdomain.UserInfo{
			Phone:     "13812348888",
			CreatedAt: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		},
	}
	uc := settingapp.NewGetProfileUseCase(reader)
	h := settinggrpc.NewHandler(uc)

	ctx := context.WithValue(context.Background(), "user_id", "abc-123")
	resp, err := h.GetProfile(ctx, &pb.GetProfileReq{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Phone != "13812348888" {
		t.Errorf("expected phone 13812348888, got %s", resp.Phone)
	}
	if resp.CreatedAt == 0 {
		t.Error("created_at should not be empty")
	}
}

func TestGetProfile_UserNotFound(t *testing.T) {
	reader := &stubUserReader{err: settingdomain.ErrUserNotFound}
	uc := settingapp.NewGetProfileUseCase(reader)
	h := settinggrpc.NewHandler(uc)

	ctx := context.WithValue(context.Background(), "user_id", "nonexistent")
	_, err := h.GetProfile(ctx, &pb.GetProfileReq{})
	if err == nil {
		t.Fatal("expected error for user not found, got nil")
	}
}

func TestGetProfile_MissingUserID(t *testing.T) {
	reader := &stubUserReader{}
	uc := settingapp.NewGetProfileUseCase(reader)
	h := settinggrpc.NewHandler(uc)

	ctx := context.Background() // no user_id
	_, err := h.GetProfile(ctx, &pb.GetProfileReq{})
	if err == nil {
		t.Fatal("expected error for missing user_id, got nil")
	}
}
