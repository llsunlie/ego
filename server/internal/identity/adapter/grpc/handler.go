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
	login *app.LoginUseCase
}

func NewHandler(login *app.LoginUseCase) *Handler {
	return &Handler{login: login}
}

func (h *Handler) Login(ctx context.Context, req *pb.LoginReq) (*pb.LoginRes, error) {
	result, err := h.login.Login(ctx, req.Account, req.Password)
	if err != nil {
		return nil, mapError(err)
	}
	return &pb.LoginRes{Token: result.Token, Created: result.Created}, nil
}

func mapError(err error) error {
	if errors.Is(err, domain.ErrInvalidPassword) {
		return status.Error(codes.Unauthenticated, "密码错误")
	}
	return status.Error(codes.Internal, err.Error())
}
