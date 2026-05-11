package bootstrap

import (
	"ego-server/internal/starmap"

	pb "ego-server/proto/ego"
)

func NewStarmapHandler(p *Platform) pb.EgoServer {
	return starmap.NewHandler(starmap.Deps{
		DB: p.Pool,
	})
}
