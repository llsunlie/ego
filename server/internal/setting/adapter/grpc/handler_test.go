package grpc_test

import (
	"context"
	"testing"
	"time"

	settinggrpc "ego-server/internal/setting/adapter/grpc"
	settingapp "ego-server/internal/setting/app"
	settingdomain "ego-server/internal/setting/domain"

	pb "ego-server/proto/ego"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// stubUserReader implements domain.UserReader for testing.
type stubUserReader struct {
	user *settingdomain.UserInfo
	err  error
}

func (s *stubUserReader) FindByID(_ context.Context, _ string) (*settingdomain.UserInfo, error) {
	return s.user, s.err
}

// stubFeedbackWriter is a test double for domain.FeedbackWriter.
type stubFeedbackWriter struct {
	saved *settingdomain.Feedback
	err   error
}

func (s *stubFeedbackWriter) Save(_ context.Context, fb *settingdomain.Feedback) error {
	if s.err != nil {
		return s.err
	}
	s.saved = fb
	return nil
}

// stubIDGenerator returns a fixed ID for testing.
type stubIDGenerator struct{}

func (stubIDGenerator) New() string { return "test-id-123" }

func TestGetProfile_WithValidUserID(t *testing.T) {
	reader := &stubUserReader{
		user: &settingdomain.UserInfo{
			Phone:     "13812348888",
			CreatedAt: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		},
	}
	uc := settingapp.NewGetProfileUseCase(reader)
	h := settinggrpc.NewHandler(uc, nil)

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
	h := settinggrpc.NewHandler(uc, nil)

	ctx := context.WithValue(context.Background(), "user_id", "nonexistent")
	_, err := h.GetProfile(ctx, &pb.GetProfileReq{})
	if err == nil {
		t.Fatal("expected error for user not found, got nil")
	}
}

func TestGetProfile_MissingUserID(t *testing.T) {
	reader := &stubUserReader{}
	uc := settingapp.NewGetProfileUseCase(reader)
	h := settinggrpc.NewHandler(uc, nil)

	ctx := context.Background() // no user_id
	_, err := h.GetProfile(ctx, &pb.GetProfileReq{})
	if err == nil {
		t.Fatal("expected error for missing user_id, got nil")
	}
}

func TestSubmitFeedback_Success(t *testing.T) {
	writer := &stubFeedbackWriter{}
	ids := stubIDGenerator{}
	uc := settingapp.NewSubmitFeedbackUseCase(writer, ids)
	h := settinggrpc.NewHandler(nil, uc)

	ctx := context.WithValue(context.Background(), "user_id", "user-1")
	res, err := h.SubmitFeedback(ctx, &pb.SubmitFeedbackReq{Content: "Great app!"})

	require.NoError(t, err)
	assert.Equal(t, "test-id-123", res.Id)
	assert.NotZero(t, res.CreatedAt)
	assert.Equal(t, "Great app!", writer.saved.Content)
	assert.Equal(t, "user-1", writer.saved.UserID)
}

func TestSubmitFeedback_EmptyContent(t *testing.T) {
	writer := &stubFeedbackWriter{}
	ids := stubIDGenerator{}
	uc := settingapp.NewSubmitFeedbackUseCase(writer, ids)
	h := settinggrpc.NewHandler(nil, uc)

	ctx := context.WithValue(context.Background(), "user_id", "user-1")
	_, err := h.SubmitFeedback(ctx, &pb.SubmitFeedbackReq{Content: "   "})

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
}

func TestSubmitFeedback_Unauthenticated(t *testing.T) {
	writer := &stubFeedbackWriter{}
	ids := stubIDGenerator{}
	uc := settingapp.NewSubmitFeedbackUseCase(writer, ids)
	h := settinggrpc.NewHandler(nil, uc)

	_, err := h.SubmitFeedback(context.Background(), &pb.SubmitFeedbackReq{Content: "test"})

	require.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Unauthenticated, st.Code())
}
