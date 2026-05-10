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
}

func NewHandler(
	createMoment *app.CreateMomentUseCase,
	generateInsight *app.GenerateInsightUseCase,
	moments domain.MomentReader,
) *Handler {
	return &Handler{
		createMoment:    createMoment,
		generateInsight: generateInsight,
		moments:         moments,
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

func (h *Handler) GetMoments(ctx context.Context, req *pb.GetMomentsReq) (*pb.GetMomentsRes, error) {
	moments, err := h.moments.GetByIDs(ctx, req.Ids)
	if err != nil {
		return nil, err
	}

	pbMoments := make([]*pb.Moment, len(moments))
	for i, m := range moments {
		pbMoments[i] = momentToProto(m)
	}

	return &pb.GetMomentsRes{
		Moments: pbMoments,
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
