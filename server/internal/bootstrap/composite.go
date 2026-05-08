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
	timeline pb.EgoServer
}

func NewEgoHandler(identity, writing, timeline pb.EgoServer) *EgoHandler {
	return &EgoHandler{
		identity: identity,
		writing:  writing,
		timeline: timeline,
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

// Trace — delegated to timeline.
func (h *EgoHandler) ListTraces(ctx context.Context, req *pb.ListTracesReq) (*pb.ListTracesRes, error) {
	return h.timeline.ListTraces(ctx, req)
}

func (h *EgoHandler) GetTraceDetail(ctx context.Context, req *pb.GetTraceDetailReq) (*pb.GetTraceDetailRes, error) {
	return h.timeline.GetTraceDetail(ctx, req)
}

// Memory Dot — delegated to timeline.
func (h *EgoHandler) GetRandomMoments(ctx context.Context, req *pb.GetRandomMomentsReq) (*pb.GetRandomMomentsRes, error) {
	return h.timeline.GetRandomMoments(ctx, req)
}
