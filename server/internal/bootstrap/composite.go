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
	setting  pb.EgoServer
}

func NewEgoHandler(identity, writing, timeline, starmap, chat, setting pb.EgoServer) *EgoHandler {
	return &EgoHandler{
		identity: identity,
		writing:  writing,
		timeline: timeline,
		starmap:  starmap,
		chat:     chat,
		setting:  setting,
	}
}

func logRPCDispatch(ctx context.Context, rpc string, module string) {
	logging.FromContext(ctx).DebugContext(ctx, rpc+": dispatch", "module", module)
}

func logRPCDone(ctx context.Context, rpc string, err error) {
	logger := logging.FromContext(ctx)
	if err != nil {
		logger.ErrorContext(ctx, rpc+": failed", "error", err)
		return
	}
	logger.DebugContext(ctx, rpc+": done")
}

// Auth — delegated to identity.
func (h *EgoHandler) Login(ctx context.Context, req *pb.LoginReq) (*pb.LoginRes, error) {
	logRPCDispatch(ctx, "Login", "identity")
	res, err := h.identity.Login(ctx, req)
	logRPCDone(ctx, "Login", err)
	return res, err
}

func (h *EgoHandler) CheckPhone(ctx context.Context, req *pb.CheckPhoneReq) (*pb.CheckPhoneRes, error) {
	logger := logging.FromContext(ctx)
	logger.InfoContext(ctx, "CheckPhone: request", "req", fmt.Sprintf("%+v", req))
	res, err := h.identity.CheckPhone(ctx, req)
	logger.InfoContext(ctx, "CheckPhone: response", "res", fmt.Sprintf("%+v", res), "error", err)
	return res, err
}

func (h *EgoHandler) SendVerificationCode(ctx context.Context, req *pb.SendVerificationCodeReq) (*pb.SendVerificationCodeRes, error) {
	logger := logging.FromContext(ctx)
	logger.InfoContext(ctx, "SendVerificationCode: request", "req", fmt.Sprintf("%+v", req))
	res, err := h.identity.SendVerificationCode(ctx, req)
	logger.InfoContext(ctx, "SendVerificationCode: response", "res", fmt.Sprintf("%+v", res), "error", err)
	return res, err
}

func (h *EgoHandler) Register(ctx context.Context, req *pb.RegisterReq) (*pb.RegisterRes, error) {
	logger := logging.FromContext(ctx)
	logger.InfoContext(ctx, "Register: request", "req", fmt.Sprintf("%+v", req))
	res, err := h.identity.Register(ctx, req)
	logger.InfoContext(ctx, "Register: response", "res", fmt.Sprintf("%+v", res), "error", err)
	return res, err
}

func (h *EgoHandler) ResetPassword(ctx context.Context, req *pb.ResetPasswordReq) (*pb.ResetPasswordRes, error) {
	logger := logging.FromContext(ctx)
	logger.InfoContext(ctx, "ResetPassword: request", "req", fmt.Sprintf("%+v", req))
	res, err := h.identity.ResetPassword(ctx, req)
	logger.InfoContext(ctx, "ResetPassword: response", "res", fmt.Sprintf("%+v", res), "error", err)
	return res, err
}

// Moment — delegated to writing.
func (h *EgoHandler) CreateMoment(ctx context.Context, req *pb.CreateMomentReq) (*pb.CreateMomentRes, error) {
	logRPCDispatch(ctx, "CreateMoment", "writing")
	res, err := h.writing.CreateMoment(ctx, req)
	logRPCDone(ctx, "CreateMoment", err)
	return res, err
}

func (h *EgoHandler) GetMoments(ctx context.Context, req *pb.GetMomentsReq) (*pb.GetMomentsRes, error) {
	logRPCDispatch(ctx, "GetMoments", "writing")
	res, err := h.writing.GetMoments(ctx, req)
	logRPCDone(ctx, "GetMoments", err)
	return res, err
}

// Insight — delegated to writing.
func (h *EgoHandler) GenerateInsight(ctx context.Context, req *pb.GenerateInsightReq) (*pb.GenerateInsightRes, error) {
	logRPCDispatch(ctx, "GenerateInsight", "writing")
	res, err := h.writing.GenerateInsight(ctx, req)
	logRPCDone(ctx, "GenerateInsight", err)
	return res, err
}

// Trace — delegated to timeline.
func (h *EgoHandler) ListTraces(ctx context.Context, req *pb.ListTracesReq) (*pb.ListTracesRes, error) {
	logRPCDispatch(ctx, "ListTraces", "timeline")
	res, err := h.timeline.ListTraces(ctx, req)
	logRPCDone(ctx, "ListTraces", err)
	return res, err
}

func (h *EgoHandler) GetTraceDetail(ctx context.Context, req *pb.GetTraceDetailReq) (*pb.GetTraceDetailRes, error) {
	logRPCDispatch(ctx, "GetTraceDetail", "timeline")
	res, err := h.timeline.GetTraceDetail(ctx, req)
	logRPCDone(ctx, "GetTraceDetail", err)
	return res, err
}

// Memory Dot — delegated to timeline.
func (h *EgoHandler) GetRandomMoments(ctx context.Context, req *pb.GetRandomMomentsReq) (*pb.GetRandomMomentsRes, error) {
	logRPCDispatch(ctx, "GetRandomMoments", "timeline")
	res, err := h.timeline.GetRandomMoments(ctx, req)
	logRPCDone(ctx, "GetRandomMoments", err)
	return res, err
}

// Stash — delegated to starmap.
func (h *EgoHandler) StashTrace(ctx context.Context, req *pb.StashTraceReq) (*pb.StashTraceRes, error) {
	logRPCDispatch(ctx, "StashTrace", "starmap")
	res, err := h.starmap.StashTrace(ctx, req)
	logRPCDone(ctx, "StashTrace", err)
	return res, err
}

// Constellation — delegated to starmap.
func (h *EgoHandler) ListConstellations(ctx context.Context, req *pb.ListConstellationsReq) (*pb.ListConstellationsRes, error) {
	logRPCDispatch(ctx, "ListConstellations", "starmap")
	res, err := h.starmap.ListConstellations(ctx, req)
	logRPCDone(ctx, "ListConstellations", err)
	return res, err
}

func (h *EgoHandler) GetConstellation(ctx context.Context, req *pb.GetConstellationReq) (*pb.GetConstellationRes, error) {
	logRPCDispatch(ctx, "GetConstellation", "starmap")
	res, err := h.starmap.GetConstellation(ctx, req)
	logRPCDone(ctx, "GetConstellation", err)
	return res, err
}

// Chat — delegated to chat.
func (h *EgoHandler) StartChat(ctx context.Context, req *pb.StartChatReq) (*pb.StartChatRes, error) {
	logRPCDispatch(ctx, "StartChat", "chat")
	res, err := h.chat.StartChat(ctx, req)
	logRPCDone(ctx, "StartChat", err)
	return res, err
}

func (h *EgoHandler) SendMessage(ctx context.Context, req *pb.SendMessageReq) (*pb.SendMessageRes, error) {
	logRPCDispatch(ctx, "SendMessage", "chat")
	res, err := h.chat.SendMessage(ctx, req)
	logRPCDone(ctx, "SendMessage", err)
	return res, err
}

// Setting — delegated to setting.
func (h *EgoHandler) GetProfile(ctx context.Context, req *pb.GetProfileReq) (*pb.GetProfileRes, error) {
	logger := logging.FromContext(ctx)
	logger.InfoContext(ctx, "GetProfile: request", "req", fmt.Sprintf("%+v", req))
	res, err := h.setting.GetProfile(ctx, req)
	logger.InfoContext(ctx, "GetProfile: response", "res", fmt.Sprintf("%+v", res), "error", err)
	return res, err
}

func (h *EgoHandler) SubmitFeedback(ctx context.Context, req *pb.SubmitFeedbackReq) (*pb.SubmitFeedbackRes, error) {
	logger := logging.FromContext(ctx)
	logger.InfoContext(ctx, "SubmitFeedback: request", "req", fmt.Sprintf("%+v", req))
	res, err := h.setting.SubmitFeedback(ctx, req)
	logger.InfoContext(ctx, "SubmitFeedback: response", "res", fmt.Sprintf("%+v", res), "error", err)
	return res, err
}
