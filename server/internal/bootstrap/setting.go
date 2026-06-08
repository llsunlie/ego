package bootstrap

import (
	"ego-server/internal/setting"

	pb "ego-server/proto/ego"
)

func NewSettingHandler(p *Platform) pb.EgoServer {
	return setting.NewHandler(setting.Deps{
		DB: p.Pool,
	})
}
