package bootstrap

import (
	"context"

	pb "ego-server/proto/ego"
)

// EgoHandler is a composite gRPC handler that delegates each RPC to the
// appropriate bounded context handler. gRPC only allows one service
// registration, so this struct routes calls to the owning module.
type EgoHandler struct {
	pb.UnimplementedEgoServer
	identity pb.EgoServer
	writing  pb.EgoServer
}

func NewEgoHandler(identity, writing pb.EgoServer) *EgoHandler {
	return &EgoHandler{
		identity: identity,
		writing:  writing,
	}
}

// Auth — delegated to identity.
func (h *EgoHandler) Login(ctx context.Context, req *pb.LoginReq) (*pb.LoginRes, error) {
	return h.identity.Login(ctx, req)
}

// Moment — delegated to writing.
func (h *EgoHandler) CreateMoment(ctx context.Context, req *pb.CreateMomentReq) (*pb.CreateMomentRes, error) {
	return h.writing.CreateMoment(ctx, req)
}

// Insight — delegated to writing.
func (h *EgoHandler) GenerateInsight(ctx context.Context, req *pb.GenerateInsightReq) (*pb.GenerateInsightRes, error) {
	return h.writing.GenerateInsight(ctx, req)
}
