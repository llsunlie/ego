package grpc

import (
	"context"

	"ego-server/internal/timeline/domain"
	writingdomain "ego-server/internal/writing/domain"

	pb "ego-server/proto/ego"
)

type Handler struct {
	pb.UnimplementedEgoServer
	moments  domain.MomentReader
	traces   domain.TraceReader
	echos    domain.EchoReader
	insights domain.InsightReader
}

func NewHandler(
	moments domain.MomentReader,
	traces domain.TraceReader,
	echos domain.EchoReader,
	insights domain.InsightReader,
) *Handler {
	return &Handler{
		moments:  moments,
		traces:   traces,
		echos:    echos,
		insights: insights,
	}
}

func (h *Handler) ListTraces(ctx context.Context, req *pb.ListTracesReq) (*pb.ListTracesRes, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, writingdomain.ErrMomentNotFound
	}

	cursor := req.Cursor
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
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
		var insight *writingdomain.Insight
		if echo != nil {
			insight, _ = h.insights.FindByMomentID(ctx, m.ID)
		}

		var echos []writingdomain.Echo
		if echo != nil {
			echos = []writingdomain.Echo{*echo}
		}

		items[i] = traceItemToProto(writingdomain.TraceItem{
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

func (h *Handler) GetRandomMoments(ctx context.Context, req *pb.GetRandomMomentsReq) (*pb.GetRandomMomentsRes, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, writingdomain.ErrMomentNotFound
	}

	count := req.Count
	if count <= 0 {
		count = 3
	}

	moments, err := h.moments.RandomByUserID(ctx, userID, count)
	if err != nil {
		return nil, err
	}

	pbMoments := make([]*pb.Moment, len(moments))
	for i, m := range moments {
		pbMoments[i] = momentToProto(m)
	}

	return &pb.GetRandomMomentsRes{
		Moments: pbMoments,
	}, nil
}
