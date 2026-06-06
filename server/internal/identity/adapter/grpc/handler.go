package identitygrpc

import (
	"context"
	"errors"

	"ego-server/internal/identity/app"
	"ego-server/internal/identity/domain"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "ego-server/proto/ego"
)

type Handler struct {
	pb.UnimplementedEgoServer
	login    *app.LoginUseCase
	register *app.RegisterUseCase
	sendCode *app.SendCodeUseCase
}

func NewHandler(
	login *app.LoginUseCase,
	register *app.RegisterUseCase,
	sendCode *app.SendCodeUseCase,
) *Handler {
	return &Handler{login: login, register: register, sendCode: sendCode}
}

func (h *Handler) SendVerificationCode(ctx context.Context, req *pb.SendVerificationCodeReq) (*pb.SendVerificationCodeRes, error) {
	result, err := h.sendCode.SendCode(ctx, req.Phone)
	if err != nil {
		return nil, mapError(err)
	}
	return &pb.SendVerificationCodeRes{Registered: result.Registered}, nil
}

func (h *Handler) Register(ctx context.Context, req *pb.RegisterReq) (*pb.RegisterRes, error) {
	result, err := h.register.Register(ctx, req.Phone, req.Code, req.Password)
	if err != nil {
		return nil, mapError(err)
	}
	return &pb.RegisterRes{Token: result.Token}, nil
}

func (h *Handler) Login(ctx context.Context, req *pb.LoginReq) (*pb.LoginRes, error) {
	result, err := h.login.Login(ctx, req.Phone, req.Password)
	if err != nil {
		return nil, mapError(err)
	}
	return &pb.LoginRes{Token: result.Token}, nil
}

func mapError(err error) error {
	if errors.Is(err, domain.ErrInvalidPassword) {
		return status.Error(codes.Unauthenticated, "密码错误")
	}
	if errors.Is(err, domain.ErrUserNotFound) {
		return status.Error(codes.NotFound, "用户不存在")
	}
	if errors.Is(err, domain.ErrInvalidPhone) {
		return status.Error(codes.InvalidArgument, "请输入正确的手机号")
	}
	if errors.Is(err, domain.ErrInvalidVerificationCode) {
		return status.Error(codes.Unauthenticated, "验证码错误")
	}
	if errors.Is(err, domain.ErrCodeExpired) {
		return status.Error(codes.Unauthenticated, "验证码已过期")
	}
	if errors.Is(err, domain.ErrPhoneAlreadyRegistered) {
		return status.Error(codes.AlreadyExists, "该手机号已注册")
	}
	return status.Error(codes.Internal, err.Error())
}
