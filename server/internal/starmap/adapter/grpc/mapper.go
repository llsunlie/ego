package grpc

import (
	"ego-server/internal/starmap/domain"
	writingdomain "ego-server/internal/writing/domain"

	pb "ego-server/proto/ego"
)

func starToProto(s domain.Star) *pb.Star {
	return &pb.Star{
		Id:      s.ID,
		TraceId: s.TraceID,
		Topic:   s.Topic,
	}
}

func constellationToProto(c domain.Constellation) *pb.Constellation {
	return &pb.Constellation{
		Id:                   c.ID,
		Name:                 c.Name,
		ConstellationInsight: c.ConstellationInsight,
		StarIds:              c.StarIDs,
		TopicPrompts:         c.TopicPrompts,
		StarCount:            int32(len(c.StarIDs)),
		CreatedAt:            c.CreatedAt.UnixMilli(),
		UpdatedAt:            c.UpdatedAt.UnixMilli(),
	}
}

func momentToProto(m writingdomain.Moment) *pb.Moment {
	return &pb.Moment{
		Id:        m.ID,
		Content:   m.Content,
		CreatedAt: m.CreatedAt.UnixMilli(),
		TraceId:   m.TraceID,
	}
}
