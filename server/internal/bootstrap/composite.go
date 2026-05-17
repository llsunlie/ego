package bootstrap

import (
	"context"
	"fmt"

	"ego-server/internal/platform/logging"

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
	starmap  pb.EgoServer
	chat     pb.EgoServer
}

func NewEgoHandler(identity, writing, timeline, starmap, chat pb.EgoServer) *EgoHandler {
	return &EgoHandler{
		identity: identity,
		writing:  writing,
		timeline: timeline,
		starmap:  starmap,
		chat:     chat,
	}
}

// Auth — delegated to identity.
func (h *EgoHandler) Login(ctx context.Context, req *pb.LoginReq) (*pb.LoginRes, error) {
	logger := logging.FromContext(ctx)
	logger.InfoContext(ctx, "Login: request", "req", fmt.Sprintf("%+v", req))
	res, err := h.identity.Login(ctx, req)
	logger.InfoContext(ctx, "Login: response", "res", fmt.Sprintf("%+v", res), "error", err)
	return res, err
}

// Moment — delegated to writing.
func (h *EgoHandler) CreateMoment(ctx context.Context, req *pb.CreateMomentReq) (*pb.CreateMomentRes, error) {
	logger := logging.FromContext(ctx)
	logger.InfoContext(ctx, "CreateMoment: request", "req", fmt.Sprintf("%+v", req))
	res, err := h.writing.CreateMoment(ctx, req)
	logger.InfoContext(ctx, "CreateMoment: response", "res", fmt.Sprintf("%+v", res), "error", err)
	return res, err
}

func (h *EgoHandler) GetMoments(ctx context.Context, req *pb.GetMomentsReq) (*pb.GetMomentsRes, error) {
	logger := logging.FromContext(ctx)
	logger.InfoContext(ctx, "GetMoments: request", "req", fmt.Sprintf("%+v", req))
	res, err := h.writing.GetMoments(ctx, req)
	logger.InfoContext(ctx, "GetMoments: response", "res", fmt.Sprintf("%+v", res), "error", err)
	return res, err
}

// Insight — delegated to writing.
func (h *EgoHandler) GenerateInsight(ctx context.Context, req *pb.GenerateInsightReq) (*pb.GenerateInsightRes, error) {
	logger := logging.FromContext(ctx)
	logger.InfoContext(ctx, "GenerateInsight: request", "req", fmt.Sprintf("%+v", req))
	res, err := h.writing.GenerateInsight(ctx, req)
	logger.InfoContext(ctx, "GenerateInsight: response", "res", fmt.Sprintf("%+v", res), "error", err)
	return res, err
}

// Trace — delegated to timeline.
func (h *EgoHandler) ListTraces(ctx context.Context, req *pb.ListTracesReq) (*pb.ListTracesRes, error) {
	logger := logging.FromContext(ctx)
	logger.InfoContext(ctx, "ListTraces: request", "req", fmt.Sprintf("%+v", req))
	res, err := h.timeline.ListTraces(ctx, req)
	logger.InfoContext(ctx, "ListTraces: response", "res", fmt.Sprintf("%+v", res), "error", err)
	return res, err
}

func (h *EgoHandler) GetTraceDetail(ctx context.Context, req *pb.GetTraceDetailReq) (*pb.GetTraceDetailRes, error) {
	logger := logging.FromContext(ctx)
	logger.InfoContext(ctx, "GetTraceDetail: request", "req", fmt.Sprintf("%+v", req))
	res, err := h.timeline.GetTraceDetail(ctx, req)
	logger.InfoContext(ctx, "GetTraceDetail: response", "res", fmt.Sprintf("%+v", res), "error", err)
	return res, err
}

// Memory Dot — delegated to timeline.
func (h *EgoHandler) GetRandomMoments(ctx context.Context, req *pb.GetRandomMomentsReq) (*pb.GetRandomMomentsRes, error) {
	logger := logging.FromContext(ctx)
	logger.InfoContext(ctx, "GetRandomMoments: request", "req", fmt.Sprintf("%+v", req))
	res, err := h.timeline.GetRandomMoments(ctx, req)
	logger.InfoContext(ctx, "GetRandomMoments: response", "res", fmt.Sprintf("%+v", res), "error", err)
	return res, err
}

// Stash — delegated to starmap.
func (h *EgoHandler) StashTrace(ctx context.Context, req *pb.StashTraceReq) (*pb.StashTraceRes, error) {
	logger := logging.FromContext(ctx)
	logger.InfoContext(ctx, "StashTrace: request", "req", fmt.Sprintf("%+v", req))
	res, err := h.starmap.StashTrace(ctx, req)
	logger.InfoContext(ctx, "StashTrace: response", "res", fmt.Sprintf("%+v", res), "error", err)
	return res, err
}

// Constellation — delegated to starmap.
func (h *EgoHandler) ListConstellations(ctx context.Context, req *pb.ListConstellationsReq) (*pb.ListConstellationsRes, error) {
	logger := logging.FromContext(ctx)
	logger.InfoContext(ctx, "ListConstellations: request", "req", fmt.Sprintf("%+v", req))
	res, err := h.starmap.ListConstellations(ctx, req)
	logger.InfoContext(ctx, "ListConstellations: response", "res", fmt.Sprintf("%+v", res), "error", err)
	return res, err
}

func (h *EgoHandler) GetConstellation(ctx context.Context, req *pb.GetConstellationReq) (*pb.GetConstellationRes, error) {
	logger := logging.FromContext(ctx)
	logger.InfoContext(ctx, "GetConstellation: request", "req", fmt.Sprintf("%+v", req))
	res, err := h.starmap.GetConstellation(ctx, req)
	logger.InfoContext(ctx, "GetConstellation: response", "res", fmt.Sprintf("%+v", res), "error", err)
	return res, err
}

// Chat — delegated to chat.
func (h *EgoHandler) StartChat(ctx context.Context, req *pb.StartChatReq) (*pb.StartChatRes, error) {
	logger := logging.FromContext(ctx)
	logger.InfoContext(ctx, "StartChat: request", "req", fmt.Sprintf("%+v", req))
	res, err := h.chat.StartChat(ctx, req)
	logger.InfoContext(ctx, "StartChat: response", "res", fmt.Sprintf("%+v", res), "error", err)
	return res, err
}

func (h *EgoHandler) SendMessage(ctx context.Context, req *pb.SendMessageReq) (*pb.SendMessageRes, error) {
	logger := logging.FromContext(ctx)
	logger.InfoContext(ctx, "SendMessage: request", "req", fmt.Sprintf("%+v", req))
	res, err := h.chat.SendMessage(ctx, req)
	logger.InfoContext(ctx, "SendMessage: response", "res", fmt.Sprintf("%+v", res), "error", err)
	return res, err
}
