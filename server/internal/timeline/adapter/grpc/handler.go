package grpc

import (
	"context"

	"ego-server/internal/timeline/app"
	writingdomain "ego-server/internal/writing/domain"

	pb "ego-server/proto/ego"
)

type Handler struct {
	pb.UnimplementedEgoServer
	listTraces       *app.ListTracesUseCase
	getTraceDetail   *app.GetTraceDetailUseCase
	getRandomMoments *app.GetRandomMomentsUseCase
}

func NewHandler(
	listTraces *app.ListTracesUseCase,
	getTraceDetail *app.GetTraceDetailUseCase,
	getRandomMoments *app.GetRandomMomentsUseCase,
) *Handler {
	return &Handler{
		listTraces:       listTraces,
		getTraceDetail:   getTraceDetail,
		getRandomMoments: getRandomMoments,
	}
}

func (h *Handler) ListTraces(ctx context.Context, req *pb.ListTracesReq) (*pb.ListTracesRes, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, writingdomain.ErrMomentNotFound
	}

	output, err := h.listTraces.Execute(ctx, app.ListTracesInput{
		UserID:   userID,
		Cursor:   req.Cursor,
		PageSize: req.PageSize,
	})
	if err != nil {
		return nil, err
	}

	pbTraces := make([]*pb.Trace, len(output.Traces))
	for i, t := range output.Traces {
		pbTraces[i] = traceToProto(t)
	}

	return &pb.ListTracesRes{
		Traces:     pbTraces,
		NextCursor: output.NextCursor,
		HasMore:    output.HasMore,
	}, nil
}

func (h *Handler) GetTraceDetail(ctx context.Context, req *pb.GetTraceDetailReq) (*pb.GetTraceDetailRes, error) {
	output, err := h.getTraceDetail.Execute(ctx, app.GetTraceDetailInput{
		TraceID: req.TraceId,
	})
	if err != nil {
		return nil, err
	}

	items := make([]*pb.TraceItem, len(output.Items))
	for i, item := range output.Items {
		items[i] = traceItemToProto(item)
	}

	return &pb.GetTraceDetailRes{
		Trace: traceToProto(output.Trace),
		Items: items,
	}, nil
}

func (h *Handler) GetRandomMoments(ctx context.Context, req *pb.GetRandomMomentsReq) (*pb.GetRandomMomentsRes, error) {
	userID, ok := ctx.Value("user_id").(string)
	if !ok || userID == "" {
		return nil, writingdomain.ErrMomentNotFound
	}

	output, err := h.getRandomMoments.Execute(ctx, app.GetRandomMomentsInput{
		UserID: userID,
		Count:  req.Count,
	})
	if err != nil {
		return nil, err
	}

	pbMoments := make([]*pb.Moment, len(output.Moments))
	for i, m := range output.Moments {
		pbMoments[i] = momentToProto(m)
	}

	return &pb.GetRandomMomentsRes{
		Moments: pbMoments,
	}, nil
}
