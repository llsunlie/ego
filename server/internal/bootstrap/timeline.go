package bootstrap

import (
	"ego-server/internal/timeline"

	pb "ego-server/proto/ego"
)

func NewTimelineHandler(p *Platform) pb.EgoServer {
	return timeline.NewHandler(timeline.Deps{
		DB: p.Pool,
	})
}
