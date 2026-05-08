package grpc

import (
	"context"

	"ego-server/internal/writing/app"

	pb "ego-server/proto/ego"
)

type Handler struct {
	pb.UnimplementedEgoServer
	createMoment    *app.CreateMomentUseCase
	generateInsight *app.GenerateInsightUseCase
}

func NewHandler(
	createMoment *app.CreateMomentUseCase,
	generateInsight *app.GenerateInsightUseCase,
) *Handler {
	return &Handler{
		createMoment:    createMoment,
		generateInsight: generateInsight,
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
