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
	login         *app.LoginUseCase
	register      *app.RegisterUseCase
	sendCode      *app.SendCodeUseCase
	checkPhone    *app.CheckPhoneUseCase
	resetPassword *app.ResetPasswordUseCase
	refreshToken  *app.RefreshTokenUseCase
}

func NewHandler(
	login *app.LoginUseCase,
	register *app.RegisterUseCase,
	sendCode *app.SendCodeUseCase,
	checkPhone *app.CheckPhoneUseCase,
	resetPassword *app.ResetPasswordUseCase,
	refreshToken *app.RefreshTokenUseCase,
) *Handler {
	return &Handler{
		login: login, register: register,
		sendCode: sendCode, checkPhone: checkPhone,
		resetPassword: resetPassword,
		refreshToken: refreshToken,
	}
}

func (h *Handler) CheckPhone(ctx context.Context, req *pb.CheckPhoneReq) (*pb.CheckPhoneRes, error) {
	result, err := h.checkPhone.Check(ctx, req.Phone)
	if err != nil {
		return nil, mapError(err)
	}
	return &pb.CheckPhoneRes{Registered: result.Registered}, nil
}

func (h *Handler) SendVerificationCode(ctx context.Context, req *pb.SendVerificationCodeReq) (*pb.SendVerificationCodeRes, error) {
	if err := h.sendCode.SendCode(ctx, req.Phone); err != nil {
		return nil, mapError(err)
	}
	return &pb.SendVerificationCodeRes{}, nil
}

func (h *Handler) Register(ctx context.Context, req *pb.RegisterReq) (*pb.RegisterRes, error) {
	result, err := h.register.Register(ctx, req.Phone, req.Code, req.Password)
	if err != nil {
		return nil, mapError(err)
	}
	return &pb.RegisterRes{AccessToken: result.AccessToken, RefreshToken: result.RefreshToken}, nil
}

func (h *Handler) Login(ctx context.Context, req *pb.LoginReq) (*pb.LoginRes, error) {
	result, err := h.login.Login(ctx, req.Phone, req.Password)
	if err != nil {
		return nil, mapError(err)
	}
	return &pb.LoginRes{AccessToken: result.AccessToken, RefreshToken: result.RefreshToken}, nil
}

func (h *Handler) ResetPassword(ctx context.Context, req *pb.ResetPasswordReq) (*pb.ResetPasswordRes, error) {
	result, err := h.resetPassword.ResetPassword(ctx, req.Phone, req.Code, req.NewPassword)
	if err != nil {
		return nil, mapError(err)
	}
	return &pb.ResetPasswordRes{AccessToken: result.AccessToken, RefreshToken: result.RefreshToken}, nil
}

func (h *Handler) RefreshToken(ctx context.Context, req *pb.RefreshTokenReq) (*pb.RefreshTokenRes, error) {
	accessToken, err := h.refreshToken.Refresh(ctx, req.RefreshToken)
	if err != nil {
		return nil, mapError(err)
	}
	return &pb.RefreshTokenRes{AccessToken: accessToken}, nil
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
	if errors.Is(err, domain.ErrInvalidRefreshToken) {
		return status.Error(codes.Unauthenticated, "登录已过期，请重新登录")
	}
	return status.Error(codes.Internal, err.Error())
}
