package grpc

import (
	"context"
	"errors"

	"ego-server/internal/setting/app"
	"ego-server/internal/setting/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "ego-server/proto/ego"
)

// Handler implements pb.EgoServer for the setting module.
type Handler struct {
	pb.UnimplementedEgoServer
	getProfile     *app.GetProfileUseCase
	submitFeedback *app.SubmitFeedbackUseCase
}

func NewHandler(getProfile *app.GetProfileUseCase, submitFeedback *app.SubmitFeedbackUseCase) *Handler {
	return &Handler{getProfile: getProfile, submitFeedback: submitFeedback}
}

func (h *Handler) GetProfile(ctx context.Context, req *pb.GetProfileReq) (*pb.GetProfileRes, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "未登录")
	}

	result, err := h.getProfile.GetProfile(ctx, userID)
	if err != nil {
		return nil, mapError(err)
	}

	return &pb.GetProfileRes{
		Phone:     result.Phone,
		CreatedAt: result.CreatedAt,
	}, nil
}

func (h *Handler) SubmitFeedback(ctx context.Context, req *pb.SubmitFeedbackReq) (*pb.SubmitFeedbackRes, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, status.Error(codes.Unauthenticated, "未登录")
	}

	result, err := h.submitFeedback.Submit(ctx, userID, req.Content)
	if err != nil {
		return nil, mapError(err)
	}

	return &pb.SubmitFeedbackRes{
		Id:        result.ID,
		CreatedAt: result.CreatedAt,
	}, nil
}

func mapError(err error) error {
	if errors.Is(err, domain.ErrUserNotFound) {
		return status.Error(codes.NotFound, "用户不存在")
	}
	if errors.Is(err, domain.ErrFeedbackEmpty) {
		return status.Error(codes.InvalidArgument, "反馈内容不能为空")
	}
	return status.Error(codes.Internal, err.Error())
}
