package grpc

import (
	"context"

	"ego-server/internal/writing/app"
	"ego-server/internal/writing/domain"

	pb "ego-server/proto/ego"
)

type Handler struct {
	pb.UnimplementedEgoServer
	createMoment    *app.CreateMomentUseCase
	generateInsight *app.GenerateInsightUseCase
	moments         domain.MomentReader
	traces          domain.TraceReader
	echos           domain.EchoRepository
	insights        domain.InsightRepository
}

func NewHandler(
	createMoment *app.CreateMomentUseCase,
	generateInsight *app.GenerateInsightUseCase,
	moments domain.MomentReader,
	traces domain.TraceReader,
	echos domain.EchoRepository,
	insights domain.InsightRepository,
) *Handler {
	return &Handler{
		createMoment:    createMoment,
		generateInsight: generateInsight,
		moments:         moments,
		traces:          traces,
		echos:           echos,
		insights:        insights,
	}
}

func (h *Handler) CreateMoment(ctx context.Context, req *pb.CreateMomentReq) (*pb.CreateMomentRes, error) {
	output, err := h.createMoment.Execute(ctx, app.CreateMomentInput{
		Content: req.Content,
		TraceID: req.TraceId,
	})
	if err != nil {
		return nil, err
	}

	return &pb.CreateMomentRes{
		Moment: momentToProto(output.Moment),
		Echo:   echoToProto(output.Echo),
	}, nil
}

func (h *Handler) GenerateInsight(ctx context.Context, req *pb.GenerateInsightReq) (*pb.GenerateInsightRes, error) {
	userID, _ := ctx.Value("user_id").(string)
	output, err := h.generateInsight.Execute(ctx, app.GenerateInsightInput{
		MomentID: req.MomentId,
		EchoID:   req.EchoId,
		UserID:   userID,
	})
	if err != nil {
		return nil, err
	}

	return &pb.GenerateInsightRes{
		Insight: insightToProto(output),
	}, nil
}

func (h *Handler) ListTraces(ctx context.Context, req *pb.ListTracesReq) (*pb.ListTracesRes, error) {
	cursor := req.Cursor
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}

	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, domain.ErrMomentNotFound // placeholder: should return auth error
	}

	traces, nextCursor, hasMore, err := h.traces.ListTracesByUserID(ctx, userID, cursor, pageSize)
	if err != nil {
		return nil, err
	}

	pbTraces := make([]*pb.Trace, len(traces))
	for i, t := range traces {
		pbTraces[i] = traceToProto(t)
	}

	return &pb.ListTracesRes{
		Traces:     pbTraces,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func (h *Handler) GetTraceDetail(ctx context.Context, req *pb.GetTraceDetailReq) (*pb.GetTraceDetailRes, error) {
	trace, err := h.traces.GetTraceByID(ctx, req.TraceId)
	if err != nil {
		return nil, err
	}

	moments, err := h.traces.ListMomentsByTraceID(ctx, req.TraceId)
	if err != nil {
		return nil, err
	}

	items := make([]*pb.TraceItem, len(moments))
	for i, m := range moments {
		echo, _ := h.echos.FindByMomentID(ctx, m.ID)
		var insight *domain.Insight
		if echo != nil {
			insight, _ = h.insights.FindByMomentID(ctx, m.ID)
		}

		var echos []domain.Echo
		if echo != nil {
			echos = []domain.Echo{*echo}
		}

		items[i] = traceItemToProto(domain.TraceItem{
			Moment:  m,
			Echos:   echos,
			Insight: insight,
		})
	}

	return &pb.GetTraceDetailRes{
		Trace: traceToProto(*trace),
		Items: items,
	}, nil
}
