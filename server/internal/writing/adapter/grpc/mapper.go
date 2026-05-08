package grpc

import (
	"ego-server/internal/writing/domain"

	pb "ego-server/proto/ego"
)

func momentToProto(m domain.Moment) *pb.Moment {
	return &pb.Moment{
		Id:        m.ID,
		Content:   m.Content,
		TraceId:   m.TraceID,
		CreatedAt: m.CreatedAt.UnixMilli(),
	}
}

func echoToProto(e *domain.Echo) *pb.Echo {
	if e == nil {
		return nil
	}
	similarities := make([]float32, len(e.Similarities))
	for i, s := range e.Similarities {
		similarities[i] = float32(s)
	}
	return &pb.Echo{
		Id:               e.ID,
		MomentId:         e.MomentID,
		MatchedMomentIds: e.MatchedMomentIDs,
		Similarities:     similarities,
	}
}

func insightToProto(i *domain.Insight) *pb.Insight {
	if i == nil {
		return nil
	}
	return &pb.Insight{
		Id:               i.ID,
		MomentId:         i.MomentID,
		EchoId:           i.EchoID,
		Text:             i.Text,
		RelatedMomentIds: i.RelatedMomentIDs,
	}
}

func traceToProto(t domain.Trace) *pb.Trace {
	return &pb.Trace{
		Id:         t.ID,
		Motivation: t.Motivation,
		Stashed:    t.Stashed,
		CreatedAt:  t.CreatedAt.UnixMilli(),
	}
}

func traceItemToProto(item domain.TraceItem) *pb.TraceItem {
	echos := make([]*pb.Echo, len(item.Echos))
	for i, e := range item.Echos {
		echos[i] = echoToProto(&e)
	}
	return &pb.TraceItem{
		Moment:  momentToProto(item.Moment),
		Echos:   echos,
		Insight: insightToProto(item.Insight),
	}
}
