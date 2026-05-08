package grpc

import (
	"ego-server/internal/writing/domain"

	pb "ego-server/proto/ego"
)

func momentToProto(m domain.Moment) *pb.Moment {
	return &pb.Moment{
		Id:        m.ID,
		Content:   m.Content,
		CreatedAt: m.CreatedAt.UnixMilli(),
		Connected: m.Connected,
		TraceId:   m.TraceID,
	}
}

func echoToProto(e *domain.Echo) *pb.Echo {
	if e == nil {
		return nil
	}
	candidates := make([]*pb.Moment, len(e.Candidates))
	for i, c := range e.Candidates {
		candidates[i] = momentToProto(c)
	}
	return &pb.Echo{
		Id:           e.ID,
		TargetMoment: momentToProto(e.TargetMoment),
		Candidates:   candidates,
		Similarity:   float32(e.Similarity),
	}
}

func insightToProto(i *domain.Insight) *pb.Insight {
	if i == nil {
		return nil
	}
	return &pb.Insight{
		Id:              i.ID,
		Text:            i.Text,
		RelatedMomentIds: i.RelatedMomentIDs,
	}
}
