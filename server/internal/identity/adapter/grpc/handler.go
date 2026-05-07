package identitygrpc

import (
	"context"
	"time"

	"ego-server/internal/platform/auth"
	"ego-server/internal/platform/postgres/sqlc"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "ego-server/proto/ego"
)

type UserQuerier interface {
	GetUserByAccount(ctx context.Context, account string) (sqlc.GetUserByAccountRow, error)
	CreateUser(ctx context.Context, arg sqlc.CreateUserParams) error
}

type Handler struct {
	pb.UnimplementedEgoServer
	users  UserQuerier
	jwtKey []byte
	jwtExp time.Duration
}

func NewHandler(users UserQuerier, jwtKey []byte, jwtExp time.Duration) *Handler {
	return &Handler{users: users, jwtKey: jwtKey, jwtExp: jwtExp}
}

func (h *Handler) Login(ctx context.Context, req *pb.LoginReq) (*pb.LoginRes, error) {
	user, err := h.users.GetUserByAccount(ctx, req.Account)
	if err != nil {
		return h.register(ctx, req)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, status.Error(codes.Unauthenticated, "密码错误")
	}

	token, err := auth.GenerateJWT(uuid.UUID(user.ID.Bytes).String(), h.jwtKey, h.jwtExp)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate token")
	}

	return &pb.LoginRes{Token: token, Created: false}, nil
}

func (h *Handler) register(ctx context.Context, req *pb.LoginReq) (*pb.LoginRes, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to hash password")
	}

	id := uuid.New()

	err = h.users.CreateUser(ctx, sqlc.CreateUserParams{
		ID:           pgtype.UUID{Bytes: [16]byte(id), Valid: true},
		Account:      req.Account,
		PasswordHash: string(hash),
		CreatedAt:    pgtype.Timestamptz{Time: time.Now(), Valid: true},
	})
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to create user")
	}

	token, err := auth.GenerateJWT(id.String(), h.jwtKey, h.jwtExp)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate token")
	}

	return &pb.LoginRes{Token: token, Created: true}, nil
}
