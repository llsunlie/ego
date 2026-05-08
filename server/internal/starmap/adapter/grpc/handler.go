package grpc

import (
	"context"

	"ego-server/internal/starmap/app"
	"ego-server/internal/starmap/domain"

	pb "ego-server/proto/ego"
)

type StashTraceUseCase interface {
	Execute(ctx context.Context, input app.StashTraceInput) (*domain.Star, error)
}

type ListConstellationsUseCase interface {
	Execute(ctx context.Context) (*app.ListConstellationsOutput, error)
}

type GetConstellationUseCase interface {
	Execute(ctx context.Context, input app.GetConstellationInput) (*app.GetConstellationOutput, error)
}

type Handler struct {
	pb.UnimplementedEgoServer
	stashTrace        StashTraceUseCase
	listConstellations ListConstellationsUseCase
	getConstellation  GetConstellationUseCase
}

func NewHandler(
	stashTrace StashTraceUseCase,
	listConstellations ListConstellationsUseCase,
	getConstellation GetConstellationUseCase,
) *Handler {
	return &Handler{
		stashTrace:        stashTrace,
		listConstellations: listConstellations,
		getConstellation:  getConstellation,
	}
}

func (h *Handler) StashTrace(ctx context.Context, req *pb.StashTraceReq) (*pb.StashTraceRes, error) {
	star, err := h.stashTrace.Execute(ctx, app.StashTraceInput{TraceID: req.TraceId})
	if err != nil {
		return nil, err
	}

	return &pb.StashTraceRes{Star: starToProto(*star)}, nil
}

func (h *Handler) ListConstellations(ctx context.Context, req *pb.ListConstellationsReq) (*pb.ListConstellationsRes, error) {
	out, err := h.listConstellations.Execute(ctx)
	if err != nil {
		return nil, err
	}

	pbConstellations := make([]*pb.Constellation, len(out.Constellations))
	for i, c := range out.Constellations {
		pbConstellations[i] = constellationToProto(c)
	}

	return &pb.ListConstellationsRes{
		Constellations: pbConstellations,
		TotalStarCount: out.TotalStarCount,
	}, nil
}

func (h *Handler) GetConstellation(ctx context.Context, req *pb.GetConstellationReq) (*pb.GetConstellationRes, error) {
	out, err := h.getConstellation.Execute(ctx, app.GetConstellationInput{ConstellationID: req.ConstellationId})
	if err != nil {
		return nil, err
	}

	pbMoments := make([]*pb.Moment, len(out.Moments))
	for i, m := range out.Moments {
		pbMoments[i] = momentToProto(m)
	}

	pbStars := make([]*pb.Star, len(out.Stars))
	for i, s := range out.Stars {
		pbStars[i] = starToProto(s)
	}

	return &pb.GetConstellationRes{
		Constellation: constellationToProto(out.Constellation),
		Moments:       pbMoments,
		Stars:         pbStars,
	}, nil
}
